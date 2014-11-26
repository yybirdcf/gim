package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

//每有一个客户端连接进来，就产生一个Client实例
type Client struct {
	conn    net.Conn
	out     chan *Message
	writer  *bufio.Writer
	id      int           //用户id
	quiting chan net.Conn //gorouting监听是否退出，如果退出则放置到server的退出列表处理
}

func (self *Client) GetId() int64 {
	return self.id
}

//初始化一个客户端结构
func CreateClient(conn net.Conn, id int) *Client {
	writer := bufio.NewWriter(conn)

	client := &Client{
		conn:   conn,
		out:    make(chan *Message),
		writer: writer,
		id:     id,
	}

	client.Listen()
	return client
}

//启动两个gorouting分别处理消息读写
func (self *Client) Listen() {
	go self.Write()
}

//处理写缓冲
func (self *Client) Write() {
	for msg := range self.out {
		m := json.Marshal(*msg)
		if _, err := self.writer.WriteString(m + "\n"); err != nil {
			self.quit()
			return
		}

		if err := self.writer.Flush(); err != nil {
			self.quit()
			return
		}
	}
}

func (self *Client) quit() {
	self.quiting <- self.conn
}

func (self *Client) close() {
	self.conn.Close()
}
