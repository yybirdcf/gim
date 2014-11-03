package sendsrv

import (
	"bufio"
	"fmt"
	"net"
)

//每有一个客户端连接进来，就产生一个Client实例
type Client struct {
	conn    net.Conn
	in      chan string
	out     chan string
	reader  *bufio.Reader
	writer  *bufio.Writer
	id      int64
	quiting chan net.Conn //gorouting监听是否退出，如果退出则放置到server的退出列表处理
}

func (self *Client) GetId() int64 {
	return self.id
}

//初始化一个客户端结构
func CreateClient(conn net.Conn, id int64) *Client {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	client := &Client{
		conn:   conn,
		in:     make(chan string),
		out:    make(chan string),
		reader: reader,
		writer: writer,
		id:     id,
	}

	client.Listen()
	return client
}

//启动两个gorouting分别处理消息读写
func (self *Client) Listen() {
	go self.Read()
	go self.Write()
}

//处理读缓冲
func (self *Client) Read() {
	for {
		if line, _, err := self.reader.ReadLine(); err == nil {
			self.in <- string(line) //放到每一个client输入chan，每一个客户端输入消息阻塞处理
		} else {
			fmt.Printf("client %d read error: \"%v\"\n", self.id, err)
			self.quit()
			return
		}
	}
}

//处理写缓冲
func (self *Client) Write() {
	for msg := range self.out {
		if _, err := self.writer.WriteString(msg + "\n"); err != nil {
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

func (self *Client) GetInMsg() string {
	return <-self.in
}

func (self *Client) PutOutMsg(msg string) {
	self.out <- msg
}
