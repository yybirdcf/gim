package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"gim/common"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	//定义来自客户端的命令
	CMD_C_MSG  = "MSG"  //客户端普通聊天消息
	CMD_C_PING = "PING" //客户端ping消息
	CMD_C_AUTH = "AUTH" //客户端认证消息
	CMD_C_RACK = "RACK" //客户端收到服务器消息，发回的确认消息

	//定义服务器产生的命令
	CMD_S_MSG  = "MSG"  //服务器普通聊天消息
	CMD_S_SACK = "SACK" //服务器收到客户端消息，发回的确认消息
	CMD_S_PONG = "PONG" //服务器响应客户端的PING消息
	CMD_S_AUTH = "AUTH" //服务器响应客户端AUTH消息
	CMD_S_KICK = "KICK" //服务器剔除客户端消息

	//未知的消息
	CMD_UNKNOW = "UNKOWN"

	//客户端状态
	CLIENT_INIT  = 0 //客户端未认证成功
	CLIENT_READY = 1 //客户端认证成功，就绪

	USER_MAX_MSGID_PREFIX = "maxmsgid#"
)

//每有一个客户端连接进来，就产生一个Client实例
type Client struct {
	conn        net.Conn    //tcp链接
	out         chan string //消息输出到客户端
	writer      *bufio.Writer
	in          chan *common.Message //客户端消息输入
	reader      *bufio.Reader
	id          int          //用户id作为结构唯一ID
	quiting     chan *Client //gorouting监听是否退出，如果退出则放置到server的退出列表处理
	activating  chan *Client //激活
	ready       int          //0未初始化，1正常初始化
	lastAccTime int64        //最近一次通信时间，心跳维持，客户端结构自检
	quited      bool         //退出
	username    string       //用户名
	password    string       //密码
}

//客户端发过来的消息结构
type ClientMsg struct {
	UniqueId int64  //nano sec可以唯一区分消息
	Content  string //消息内容
	To       int    //发送给的目标
	Type     int    //消息类型,个人消息，还是群组消息
}

//客户端发送结构
type ClientCmd struct {
	Cmd    string //命令类型
	Params string //参数&作为分隔符
}

//响应客户端接口
type ClientResp struct {
	RetCode int         //返回码
	RetType string      //响应类型
	RetMsg  string      //响应消息
	RetData interface{} //响应数据
}

//初始化一个客户端结构
func CreateClient() *Client {
	client := &Client{
		out:        make(chan string),
		in:         make(chan *common.Message),
		quiting:    make(chan *Client),
		activating: make(chan *Client),
	}

	return client
}

func (self *Client) Init(conn net.Conn) {
	self.writer = bufio.NewWriter(conn)
	self.reader = bufio.NewReader(conn)

	self.conn = conn
	self.lastAccTime = time.Now().Unix()
	self.ready = CLIENT_INIT
	self.quited = false
	self.id = 0

	self.Listen()
}

//启动gorouting处理消息读写
func (self *Client) Listen() {
	go self.Read()  //开启读goroute
	go self.Write() //开启写goroute
}

//处理写缓冲
func (self *Client) Write() {
	//发生写异常直接断开客户端
	defer func() {
		if err := recover(); err != nil {
			self.Quit()
		}
	}()

	for {
		m := <-self.out

		if self.quited {
			return
		}
		fmt.Printf("%v\n", m)
		self.writer.WriteString(m + "\n")
		self.writer.Flush()
	}
}

//处理客户端读
//认证通过以后才能正常收发消息
func (self *Client) Read() {
	var cc ClientCmd
	var cm ClientMsg

	//redis连接池获取redis客户端
	redClient := redisPool.Get()
	defer redClient.Close()

	for {
		if self.quited {
			return
		}

		if line, err := self.reader.ReadBytes('\n'); err == nil {
			line = bytes.TrimRight(line, "\r\n")
			err := json.Unmarshal(line, &cc)
			if err != nil {
				ret := RetJson(ERR_CLIENT_MSG_FORMAT, CMD_UNKNOW, "未知消息格式", nil)
				self.out <- ret
				continue
			}

			if self.ready == CLIENT_READY {
				//已经认证通过，可以正常收发消息
				switch cc.Cmd {
				case CMD_C_MSG:
					//客户端聊天消息
					err = json.Unmarshal([]byte(cc.Params), &cm)
					if err != nil {
						ret := RetJson(ERR_CLIENT_MSG_FORMAT, CMD_UNKNOW, "未知消息格式", nil)
						self.out <- ret
						continue
					}

					if cm.Type == common.MESSAGE_TYPE_USER || cm.Type == common.MESSAGE_TYPE_GROUP {
						//将消息发送给send srv
						//构造消息结构
						group := 0
						to := cm.To
						if cm.Type == common.MESSAGE_TYPE_GROUP {
							group = cm.To
							to = 0
						}

						m := <-getMsg
						m.Mid = 0
						m.Uid = 0
						m.Content = cm.Content
						m.Type = cm.Type
						m.Time = time.Now().Unix()
						m.From = self.id
						m.To = to
						m.Group = group

						self.in <- m

						self.lastAccTime = time.Now().Unix()
						//回写发送成功消息
						ret := RetJson(0, CMD_S_SACK, "OK", cm.UniqueId)
						self.out <- ret
					} else {
						ret := RetJson(ERR_CLIENT_MSG_UNKNOW_TYPE, CMD_UNKNOW, "未知消息的消息类型", nil)
						self.out <- ret
						continue
					}
				case CMD_C_AUTH:
					//认证消息
					self.lastAccTime = time.Now().Unix()
					ret := RetJson(0, CMD_S_AUTH, "已认证通过", nil)
					self.out <- ret
				case CMD_C_PING:
					//客户端维持心跳消息
					self.lastAccTime = time.Now().Unix()
					ret := RetJson(0, CMD_S_PONG, "PONG", nil)
					self.out <- ret
				case CMD_C_RACK:
					//客户端消息回执, redis存储已读消息id
					redClient.Do("SET", USER_MAX_MSGID_PREFIX+strconv.Itoa(self.id), cc.Params)
				default:
					ret := RetJson(ERR_CLIENT_MSG_UNKNOW_TYPE, CMD_UNKNOW, "未知的消息类型", nil)
					self.out <- ret
				}
			} else {
				//认证
				if cc.Cmd == CMD_C_AUTH {
					//认证参数  {app_key}&{app_secret}&{user_name}&{user_password}
					params := strings.Split(cc.Params, "&")
					if len(params) < 4 {
						ret := RetJson(ERR_CLIENT_AUTH_FAILED, CMD_S_AUTH, "认证失败，参数错误", nil)
						self.out <- ret
						return
					}

					if params[0] != common.APP_KEY || params[1] != common.APP_SECRET {
						ret := RetJson(ERR_CLIENT_AUTH_FAILED, CMD_S_AUTH, "认证失败，应用信息错误", nil)
						self.out <- ret
						return
					}

					//待认证的账户信息
					self.username = params[2]
					self.password = params[3]
					self.activating <- self
				} else {
					ret := RetJson(ERR_CLIENT_MSG_UNKNOW_TYPE, CMD_UNKNOW, "未知的消息类型，需要认证", nil)
					self.out <- ret
				}
			}
		} else if err == io.EOF {
			self.Quit()
			break
		} else {
			//读取数据失败，可能连接错误，活着连接关闭等等
			self.Quit()
			break
		}
	}

}

//自检，如果lastAccTime超过60秒就断开客户端连接
func (self *Client) CheckSelf() {
	go func() {
		for {
			if self.quited {
				return
			}
			time.Sleep(time.Second * 10)
			//超过60秒，认为该连接已断开
			if self.lastAccTime+60 < time.Now().Unix() {
				self.Quit()
			}
		}
	}()
}

//退出
func (self *Client) Quit() {
	if self.quited {
		return
	}

	self.quiting <- self
}

//关闭客户端结构资源
func (self *Client) Close() {
	if self.conn != nil {
		self.conn.Close()
	}

	self.writer = nil
	self.reader = nil
	self.conn = nil
	self.quited = true
}

//剔出，发送剔出命令到客户端
func (self *Client) Kickout() {
	//被踢，先发送一个被踢命令到客户端，然后关闭连接
	ret := RetJson(0, CMD_S_KICK, "OK", nil)
	self.writer.WriteString(ret + "\n")
	self.writer.Flush()
	self.Close()
}

//彻底清除资源
func (self *Client) ShutDown() {
	close(self.in)
	close(self.out)
	close(self.activating)
	close(self.quiting)
}

func RetJson(retCode int, retType string, retMsg string, retData interface{}) string {
	resp := ClientResp{
		RetCode: retCode,
		RetType: retType,
		RetMsg:  retMsg,
		RetData: retData,
	}

	str, _ := json.Marshal(resp)
	return string(str)
}
