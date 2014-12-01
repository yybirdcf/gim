package main

import (
	"encoding/json"
	"fmt"
	"gim/common"
	"github.com/garyburd/redigo/redis"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"
	"time"
)

var (
	redClient     redis.Conn
	sendSrvClient *rpc.Client
	in            chan *common.Message
)

const (
	USER_ONLINE_PREFIX = "useron#"
)

//定义服务器结构
type Server struct {
	listener   net.Listener
	clients    map[int]*Client
	lock       *sync.RWMutex
	pending    chan net.Conn
	quiting    chan *Client
	activating chan *Client
	out        chan *common.Message
}

func CreateServer() *Server {
	server := &Server{
		clients:    make(map[int]*Client),
		pending:    make(chan net.Conn),
		quiting:    make(chan *Client),
		activating: make(chan *Client),
		in:         make(chan *common.Message),
		out:        make(chan *common.Message),
	}

	conn, _ := redis.Dial("tcp", Conf.Redis)
	redClient = conn

	server.listen()
	return server
}

func (self *Server) listen() {
	var resp ClientResp
	go func() {
		for {
			select {
			case msg := <-in:
				//新消息处理
				//获取客户端
				client := self.clients[msg.Uid]
				if client != nil {
					resp.retCode = 0
					resp.retType = CMD_MSG
					resp.retMsg = "OK"
					resp.retData = *msg

					client.lastAccTime = int(time.Now().Unix())
					str, _ := json.Marshal(resp)
					client.out <- string(str)
				}
			case msg := <-self.out:
				//客户端需要发出去的消息
				s, _ := json.Marshal(*msg)
				redClient.Do("LPUSH", "msg_queue_0", string(s))
			case conn := <-self.pending:
				self.join(conn) //新客户端处理
			case client := <-self.quiting:
				self.quit(client) //退出客户端处理
			case client := <-self.activating:
				self.activate(client) //激活客户端处理
			}
		}
	}()
}

//增加一个客户端
func (self *Server) join(conn net.Conn) {
	client := CreateClient(conn)
	//开一个gorouting处理客户端激活
	go func() {
		c := <-client.activating
		//c!=nil代表时激活，否则就结束本次激活goroute
		if c != nil {
			fmt.Printf("client %d is activating", c.id)
			self.activating <- c
		}
	}()

	//开一个gorouting处理客户端退出
	go func() {
		c := <-client.quiting
		fmt.Printf("client %d is quiting", c.id)
		self.quiting <- c
	}()
}

//删除一个客户端
func (self *Server) quit(client *Client) {
	if client != nil {
		//删除客户端map信息
		self.lock.Lock()
		delete(self.clients, client.id)
		self.lock.Unlock()

		//删除客户端在线信息
		redClient.Do("DEL", USER_ONLINE_PREFIX+strconv.Itoa(client.id))

		//如果客户端没有激活，需要关闭激活goroute
		client.activating <- nil
		client.Close()
	}
}

//激活一个客户端,如果重连，删除之前的客户端结构
func (self *Server) activate(client *Client) {
	if client != nil {
		self.lock.Lock()

		if _, ok := self.clients[client.id]; ok {
			self.quit(self.clients[client.id])
		}

		self.clients[client.id] = client
		self.lock.Unlock()

		//激活的客户端开启goroute处理客户端发出的消息
		go func() {
			for {
				if msg := <-client.in; msg != nil {
					self.out <- msg
				}
			}
		}()
	}
}

func (self *Server) Start() {
	listener, err := net.Listen("tcp", Conf.TcpBind)
	self.listener = listener
	if err != nil {
		fmt.Printf("server %s listen failed", Conf.TcpBind)
		return
	}

	fmt.Printf("server %s start\n", Conf.TcpBind)

	go func() {
		fmt.Printf("begin to accept connects\n")

		for {
			conn, err := self.listener.Accept()

			if err != nil {
				fmt.Printf("server accept error : \"%v\"", err)
				return
			}

			fmt.Printf("new client connect: \"%v\"", conn)
			self.pending <- conn
		}
	}()
}

func (self *Server) Stop() {
	self.listener.Close()
}

type ConnRpcServer struct {
}

//push server 调用的rpc服务，负责推送消息
func (self *ConnRpcServer) SendMsg(args *common.Message, reply *bool) error {
	in <- args

	return nil
}

func StartRpc() {
	go func() {
		rpcSrv := &RpcServer{}
		rpc.Register(rpcSrv)
		rpc.HandleHTTP()

		err := http.ListenAndServe(Conf.RcpBind, nil)
		if err != nil {
			fmt.Printf("connect server rpc error: %s\n", err.Error())
		}
	}()
}
