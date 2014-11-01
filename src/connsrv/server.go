package main

import (
	"common"
	"glog"
	"net"
	"time"
)

//定义服务器结构
type Server struct {
	listener net.Listener
	clients  *common.SafeMap
	tokens   chan int //令牌，确保服务器同时服务的客户端数量一定
	pending  chan net.Conn
	quiting  chan net.Conn
	in       chan string
	out      chan string
}

//生成一个token，每一个token对应一个客户端
func (self *Server) GenerateToken() {
	self.tokens <- 0
}

//去掉一个token，客户端就可以连接，没有token可去掉，则阻塞，后续客户端连接不进来
func (self *Server) TakeToken() {
	<-self.tokens
}

func CreateServer() *Server {
	server := &Server{
		clients: common.NewSafeMap(map[net.Conn]*Client),
		tokens:  make(chan int),
		pending: make(chan net.Conn),
		quiting: make(chan net.Conn),
		in:      make(chan string),
		out:     make(chan string),
	}

	server.listen()
	return server
}

func (self *Server) listen() {
	go func() {
		for {
			select {
			case msg := <-self.in:
				HandleServerMsg(msg) //新消息处理
			case conn := <-self.pending:
				self.join(conn) //新客户端处理
			case conn := <-self.quiting:
				self.quit(conn) //退出客户端处理
			}
		}
	}()
}

//增加一个客户端
func (self *Server) join(conn net.Conn) {
	guid, err := common.NewGuid(time.Now().Second())
	if err != nil {
		glog.Errorf("new guid failed")
		self.quiting <- conn
		return
	}

	client := CreateClient(conn)
	self.clients.Set(conn, client)

	glog.Infof("new client join: \"%v\"", conn)

	//开一个gorouting处理这个客户端输入
	go func() {
		for {
			msg := <-client.in
			glog.Infof("Got msg: %s from client id: %d", msg, client.GetId())

			//处理消息
			//parse msg
			HandleClientMsg(msg)
		}
	}()

	//开一个gorouting处理客户端退出
	go func() {
		for {
			conn := <-client.quiting
			glog.Infof("client %d is quiting", client.GetId())
			self.quiting <- conn
		}
	}()
}

//删除一个客户端
func (self *Server) quit(conn net.Conn) {
	if conn != nil {
		conn.Close()
		self.clients.Delete(conn)
	}

	self.GenerateToken() //归还token
}

func (self *Server) Start() {
	self.listener, err = net.Listen("tcp", Conf.TcpBind)
	if err != nil {
		glog.Infof("server %s listen failed", Conf.TcpBind)
		return
	}

	glog.Infof("server %s start", Conf.TcpBind)

	//预先生成指定连接数的token
	for i := 0; i < Conf.MaxClients; i++ {
		self.GenerateToken()
	}

	go func() {
		for {
			conn, err := self.listener.Accept()

			if err != nil {
				glog.Errorf("server accept error : \"%v\"", err)
				return
			}

			glog.Infof("new client connect: \"%v\"", conn)
			self.TakeToken() //如果没有token了，会阻塞在这里
			self.pending <- conn
		}
	}()
}

func (self *Server) Stop() {
	close(self.tokens)
	close(self.pending)
	close(self.quiting)
	close(self.in)
	close(self.out)
	self.listener.Close()
}