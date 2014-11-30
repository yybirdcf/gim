package main

import (
	"bufio"
	"encoding/json"
	"gim/common"
	"io"
	"net"
	"strconv"
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
	CMD_AUTH    = "AUTH"
	CMD_UNKNOW  = "UNKOWN"

	CLIENT_INIT  = 0
	CLIENT_READY = 1
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
}

//客户端发送命令
type ClientCmd struct {
	cmd    string //命令类型
	params string //参数&作为分隔符
}

type ClientResp struct {
	retCode int
	retType string
	retMsg  string
	retData interface{}
}

//初始化一个客户端结构
func CreateClient(conn net.Conn) *Client {
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	client := &Client{
		conn:        conn,
		out:         make(chan string),
		writer:      writer,
		in:          make(chan *common.Message),
		reader:      reader,
		quiting:     make(chan *Client),
		activating:  make(chan *Client),
		lastAccTime: int(time.Now().Unix()),
		ready:       CLIENT_INIT,
	}

	client.Listen()
	return client
}

//启动gorouting处理消息读
func (self *Client) Listen() {
	go self.Read()
}

//处理写缓冲
func (self *Client) Write() {
	for m := range self.out {
		if _, err := self.writer.WriteString(m + "\n"); err != nil {
			self.Quit()
			return
		}

		if err := self.writer.Flush(); err != nil {
			self.Quit()
			return
		}
	}
}

//处理客户端读
//认证通过以后才能正常收发消息
func (self *Client) Read() {
	var clientCmd ClientCmd
	var resp ClientResp
	var cm ClientMsg

	for {
		if line, _, err := self.reader.ReadLine(); err == nil {
			if self.ready == CLIENT_READY {
				//已经认证通过，可以正常收发消息
				err := json.Unmarshal(line, &clientCmd)
				if err != nil {
					resp.retCode = -1
					resp.retType = CMD_UNKNOW
					resp.retMsg = "解析消息失败"
					resp.retData = nil
					str, _ := json.Marshal(resp)
					self.out <- string(str)
					continue
				}

				if clientCmd.cmd == CMD_MSG {
					//正常的用户消息
					err = json.Unmarshal([]byte(clientCmd.params), &cm)
					if err != nil {
						resp.retCode = -1
						resp.retType = CMD_UNKNOW
						resp.retMsg = "解析消息失败"
						resp.retData = nil
						str, _ := json.Marshal(resp)
						self.out <- string(str)
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

						m := &common.Message{
							Mid:     0,
							Uid:     0,
							Content: cm.Content,
							Type:    cm.Type,
							Time:    int(time.Now().Unix()),
							From:    self.id,
							To:      to,
							Group:   group,
						}

						self.in <- m

						//回写发送成功消息
						resp.retCode = 0
						resp.retType = CMD_MSG_ACK
						resp.retMsg = "OK"
						resp.retData = cm.UniqueId //用户客户端确认消息是否发送成功

						self.lastAccTime = int(time.Now().Unix())
						str, _ := json.Marshal(resp)
						self.out <- string(str)
					} else {
						resp.retCode = -1
						resp.retType = CMD_UNKNOW
						resp.retMsg = "解析消息失败"
						resp.retData = nil
						str, _ := json.Marshal(resp)
						self.out <- string(str)
						continue
					}

				} else if clientCmd.cmd == CMD_AUTH {
					//不需要认证
					resp.retCode = 0
					resp.retType = CMD_AUTH
					resp.retMsg = "已经认证通过"
					resp.retData = nil
					self.lastAccTime = int(time.Now().Unix())

					str, _ := json.Marshal(resp)
					self.out <- string(str)
				} else if clientCmd.cmd == CMD_PING {
					//客户端ping，返回pong
					resp.retCode = 0
					resp.retType = CMD_PING
					resp.retMsg = "PONG"
					resp.retData = nil

					self.lastAccTime = int(time.Now().Unix())
					str, _ := json.Marshal(resp)
					self.out <- string(str)
				} else {
					//未知的消息类型
					resp.retCode = -1
					resp.retType = CMD_UNKNOW
					resp.retMsg = "未知的消息类型"
					resp.retData = nil
					str, _ := json.Marshal(resp)
					self.out <- string(str)
				}
			} else {
				//认证
				err := json.Unmarshal(line, &clientCmd)
				if clientCmd.cmd == CMD_AUTH {
					//暂时没有认证过程,参数{uid}//{uid}&{token}
					uid, _ := strconv.Atoi(clientCmd.params)
					self.id = uid
					self.ready = CLIENT_READY
					self.activating <- self

					resp.retCode = 0
					resp.retType = CMD_AUTH
					resp.retMsg = "认证成功"
					resp.retData = nil

					go self.Write() //开启写goroute
					str, _ := json.Marshal(resp)
					self.out <- string(str)
				} else {
					resp.retCode = -1
					resp.retType = CMD_UNKNOW
					resp.retMsg = "未知的消息类型，需要认证"
					resp.retData = nil
					str, _ := json.Marshal(resp)
					self.out <- string(str)
				}
			}
		} else if err == io.EOF {
			self.Quit()
		}
	}
}

func (self *Client) Quit() {
	self.quiting <- self
}

func (self *Client) Close() {
	self.conn.Close()
}
