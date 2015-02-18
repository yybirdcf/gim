package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"gim/common"
	"io"
	"net"
	"strings"
	"time"
)

//客户端发过来的消息结构
type ClientMsg struct {
	UniqueId int64  //nano sec可以唯一区分消息
	Content  string //消息内容
	To       int    //发送给的目标
	Type     int    //消息类型,个人消息，还是群组消息
}

const (
	CMD_MSG     = "MSG"
	CMD_MSG_ACK = "MSGACK"
	CMD_PING    = "PING"
	CMD_PONG    = "PONG"
	CMD_AUTH    = "AUTH"
	CMD_UNKNOW  = "UNKOWN"

	CLIENT_INIT  = 0
	CLIENT_READY = 1

	ERR_CODE_MSG_FORMAT      = 1000
	ERR_CODE_MSG_UNKNOW_TYPE = 1001
	ERR_AUTH_FAILED          = 1002
	ERR_NEED_AUTH            = 1003
)

//每有一个客户端连接进来，就产生一个Client实例
type Client struct {
	conn        net.Conn
	out         chan string
	writer      *bufio.Writer
	in          chan *common.Message
	reader      *bufio.Reader
	id          int          //用户id
	quiting     chan *Client //gorouting监听是否退出，如果退出则放置到server的退出列表处理
	activating  chan *Client //激活
	ready       int          //0未初始化，1正常初始化
	lastAccTime int          //最近一次通信时间
	quited      bool         //退出
	username    string       //用户名
	password    string       //密码
}

//客户端发送命令
type ClientCmd struct {
	Cmd    string //命令类型
	Params string //参数&作为分隔符
}

type ClientResp struct {
	RetCode int
	RetType string
	RetMsg  string
	RetData interface{}
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
	self.lastAccTime = int(time.Now().Unix())
	self.ready = CLIENT_INIT
	self.quited = false
	self.id = 0

	self.Listen()
}

//启动gorouting处理消息读
func (self *Client) Listen() {
	go self.Read()
	go self.Write() //开启写goroute
}

//处理写缓冲
func (self *Client) Write() {
	for m := range self.out {
		if self.quited {
			return
		}

		if _, err := self.writer.WriteString(m + "\n"); err != nil {
			return
		}

		if err := self.writer.Flush(); err != nil {
			return
		}
	}
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

//处理客户端读
//认证通过以后才能正常收发消息
func (self *Client) Read() {
	var clientCmd ClientCmd
	var cm ClientMsg

	for {
		if self.quited {
			return
		}

		if line, err := self.reader.ReadBytes('\n'); err == nil {
			line = bytes.TrimRight(line, "\r\n")
			err := json.Unmarshal(line, &clientCmd)
			if err != nil {
				ret := RetJson(ERR_CODE_MSG_FORMAT, CMD_UNKNOW, "消息格式错误", nil)
				self.out <- ret
				continue
			}

			if self.ready == CLIENT_READY {
				//已经认证通过，可以正常收发消息
				switch clientCmd.Cmd {
				case CMD_MSG:
					err = json.Unmarshal([]byte(clientCmd.Params), &cm)
					if err != nil {
						ret := RetJson(ERR_CODE_MSG_FORMAT, CMD_UNKNOW, "消息格式错误", nil)
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
						m.Time = int(time.Now().Unix())
						m.From = self.id
						m.To = to
						m.Group = group

						self.in <- m

						self.lastAccTime = int(time.Now().Unix())
						//回写发送成功消息
						ret := RetJson(0, CMD_MSG_ACK, "OK", cm.UniqueId)
						self.out <- ret
					} else {
						ret := RetJson(ERR_CODE_MSG_UNKNOW_TYPE, CMD_UNKNOW, "未知消息的消息类型", nil)
						self.out <- ret
						continue
					}
				case CMD_AUTH:
					self.lastAccTime = int(time.Now().Unix())
					ret := RetJson(0, CMD_AUTH, "已经认证通过", nil)
					self.out <- ret
				case CMD_PING:
					self.lastAccTime = int(time.Now().Unix())
					ret := RetJson(0, CMD_PONG, "PONG", nil)
					self.out <- ret
				case CMD_PONG:
					//服务器发出的心跳响应
					self.lastAccTime = int(time.Now().Unix())
				default:
					ret := RetJson(ERR_CODE_MSG_UNKNOW_TYPE, CMD_UNKNOW, "未知的消息类型", nil)
					self.out <- ret
				}
			} else {
				//认证
				if clientCmd.Cmd == CMD_AUTH {
					//认证参数  {app_key}&{app_secret}&{user_name}&{user_password}
					params := strings.Split(clientCmd.Params, "&")
					if len(params) < 4 {
						ret := RetJson(ERR_AUTH_FAILED, CMD_AUTH, "认证失败，参数错误", nil)
						self.out <- ret
						return
					}

					if params[0] != common.APP_KEY || params[1] != common.APP_SECRET {
						ret := RetJson(ERR_AUTH_FAILED, CMD_AUTH, "认证失败，app key secret匹配错误", nil)
						self.out <- ret
						return
					}

					self.username = params[2]
					self.password = params[3]
					self.ready = CLIENT_READY
					self.activating <- self
				} else {
					ret := RetJson(ERR_NEED_AUTH, CMD_UNKNOW, "未知的消息类型，需要认证", nil)
					self.out <- ret
				}
			}
		} else if err == io.EOF {
			self.Quit()
			break
		} else {
			//读取数据失败，可能连接错误，活着连接关闭等等
			fmt.Printf("%v\n", err.Error())
			self.Quit()
			break
		}
	}

}

//定期发送心跳包到客户端
func (self *Client) KeepCliAlive() {
	go func() {
		for {
			if self.quited {
				return
			}

			time.Sleep(time.Second * 3)

			ret := RetJson(0, CMD_PING, "", nil)
			self.out <- ret
		}
	}()
}

//自检
func (self *Client) CheckSelf() {
	go func() {
		for {
			if self.quited {
				return
			}
			time.Sleep(time.Second * 30)
			//超过60秒，认为该连接已断开
			if self.lastAccTime+60 < int(time.Now().Unix()) {
				self.Quit()
			}
		}
	}()
}

func (self *Client) Quit() {
	if self.quited {
		return
	}

	self.quiting <- self
}

func (self *Client) Close() {
	if self.conn != nil {
		fmt.Printf("%d close\n", self.id)
		self.conn.Close()
	}

	self.writer = nil
	self.reader = nil
	self.conn = nil
	self.quited = true
}

func (self *Client) ShutDown() {
	close(self.in)
	close(self.out)
	close(self.activating)
	close(self.quiting)
}
