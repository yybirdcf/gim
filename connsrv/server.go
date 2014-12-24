package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"gim/common"
	"github.com/garyburd/redigo/redis"
	"github.com/samuel/go-zookeeper/zk"
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
	out           chan *common.Message
	zkConn        *zk.Conn
	get           chan *Client
	put           chan *Client
	getMsg        chan *common.Message
	putMsg        chan *common.Message
)

//定义服务器结构
type Server struct {
	listener   net.Listener
	clients    map[int]*Client
	lock       *sync.RWMutex
	pending    chan net.Conn
	quiting    chan *Client
	activating chan *Client
}

func CreateServer() *Server {
	server := &Server{
		clients:    make(map[int]*Client),
		lock:       new(sync.RWMutex),
		pending:    make(chan net.Conn),
		quiting:    make(chan *Client),
		activating: make(chan *Client),
	}

	in = make(chan *common.Message)
	out = make(chan *common.Message)
	conn, err := redis.Dial("tcp", Conf.Redis)
	if err != nil {
		panic(err.Error())
	}
	redClient = conn

	get, put = makeClientRecycler()
	getMsg, putMsg = common.MakeMessageRecycler()

	server.listen()
	return server
}

func (self *Server) listen() {
	go func() {
		for {
			select {
			case msg := <-in:
				//新消息处理
				//获取客户端
				client := self.clients[msg.Uid]
				if client != nil {
					client.lastAccTime = int(time.Now().Unix())

					ret := RetJson(0, CMD_MSG, "OK", *msg)
					client.out <- ret
				}
			case msg := <-out:
				//客户端需要发出去的消息
				s, _ := json.Marshal(*msg)
				redClient.Do("LPUSH", common.MSG_QUEUE_0, string(s))
				putMsg <- msg
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
	client := <-get
	client.Init(conn)
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
		if c != nil {
			fmt.Printf("client %d is quiting", c.id)
			self.quiting <- c
		}
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
		_, err := redClient.Do("DEL", common.USER_ONLINE_PREFIX+strconv.Itoa(client.id))
		if err != nil {
			fmt.Printf("delete %d online map status failed\n", client.id)
		}
		_, err2 := redClient.Do("DEL", common.USER_ONLINE_HOST_PREFIX+strconv.Itoa(client.id))
		if err2 != nil {
			fmt.Printf("delete %d online host map status failed\n", client.id)
		}

		client.Close()
		put <- client
		fmt.Printf("client quited\n")
	}
}

//激活一个客户端,如果重连，删除之前的客户端结构
func (self *Server) activate(client *Client) {
	if client != nil {
		self.lock.Lock()
		if _, ok := self.clients[client.id]; ok {
			//删除客户端在线信息
			_, err := redClient.Do("DEL", common.USER_ONLINE_PREFIX+strconv.Itoa(client.id))
			if err != nil {
				fmt.Printf("delete %d online map status failed\n", client.id)
			}
			_, err2 := redClient.Do("DEL", common.USER_ONLINE_HOST_PREFIX+strconv.Itoa(client.id))
			if err2 != nil {
				fmt.Printf("delete %d online host map status failed\n", client.id)
			}
			//如果客户端没有激活，需要关闭激活goroute
			fmt.Printf("server %d close\n", client.id)
			self.clients[client.id].Close()
			delete(self.clients, client.id)
		}

		self.clients[client.id] = client
		self.lock.Unlock()

		//用户上线
		redClient.Do("SET", common.USER_ONLINE_PREFIX+strconv.Itoa(client.id), 1)
		//写用户在线的机器
		redClient.Do("SET", common.USER_ONLINE_HOST_PREFIX+strconv.Itoa(client.id), Conf.RcpBind)
		//激活的客户端开启goroute处理客户端发出的消息
		go func() {
			for {
				if client.quited {
					return
				}

				if msg := <-client.in; msg != nil {
					out <- msg
				}
			}
		}()

		//开启客户端心跳维持
		client.KeepCliAlive()
		//开启客户端自检
		client.CheckSelf()
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
			fmt.Printf("new client into pending\n")
		}
	}()

	//服务器初始化完成以后，开启zookeeper
	zkConn = common.ZkConnect(Conf.ZooKeeper)
	common.ZkCreateRoot(zkConn, Conf.ZkRoot)
	//为当前ms服务器创建一个节点，加入到ms集群中
	path := Conf.ZkRoot + "/" + Conf.RcpBind
	common.ZkCreateTempNode(zkConn, path)
	go func() {
		for {
			exist, _, watch, err := zkConn.ExistsW(path)
			if err != nil {
				//发生错误，当前节点退出
				fmt.Printf("%s occur error\n", path)
				common.KillSelf()
				return
			}

			if !exist {
				//节点不存在了
				fmt.Printf("%s not exist\n", path)
				common.KillSelf()
				return
			}

			event := <-watch
			fmt.Printf("%s receiver a event %v\n", path, event)
		}
	}()
}

func (self *Server) Stop() {
	self.listener.Close()
	zkConn.Delete(Conf.ZkRoot+"/"+Conf.RcpBind, 0)
	zkConn.Close()
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
		rpcSrv := &ConnRpcServer{}
		rpc.Register(rpcSrv)
		rpc.HandleHTTP()

		err := http.ListenAndServe(Conf.RcpBind, nil)
		if err != nil {
			fmt.Printf("connect server rpc error: %s\n", err.Error())
		}
	}()

	//将机器机器rpc对应的tcp关系写了redis，让前端分配tcp服务器时候查找
	redClient.Do("SET", common.RCP_TCP_HOST_PREFIX+Conf.RcpBind, Conf.TcpBind)
}

//重用client结构
func makeClientRecycler() (get, put chan *Client) {
	get = make(chan *Client)
	put = make(chan *Client)

	go func() {
		queue := new(list.List)
		for {
			if queue.Len() == 0 {
				queue.PushFront(CreateClient())
			}

			ct := queue.Front()

			timeout := time.NewTimer(time.Minute)
			select {
			case b := <-put:
				timeout.Stop()
				b.lastAccTime = int(time.Now().Unix())
				queue.PushFront(b)
			case get <- ct.Value.(*Client):
				timeout.Stop()
				queue.Remove(ct)
			case <-timeout.C:
				ct := queue.Front()
				for ct != nil {
					n := ct.Next()
					if (int(time.Now().Unix()) - ct.Value.(*Client).lastAccTime) > int(time.Second*60) {
						ct.Value.(*Client).ShutDown()
						queue.Remove(ct)
						ct.Value = nil
					}
					ct = n
				}
			}
		}
	}()

	return
}
