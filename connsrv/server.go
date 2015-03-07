package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"gim/common"
	"github.com/garyburd/redigo/redis"
	"github.com/samuel/go-zookeeper/zk"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"
	"time"
)

var (
	redisPool *redis.Pool          //redis 连接池
	msClients [20]*rpc.Client      //ms rpc客户端列表
	isMsStop  bool                 //message store是否暂停
	msLen     int                  //message store节点数量
	in        chan *common.Message //输入消息
	out       chan *common.Message //输出消息
	zkConn    *zk.Conn             //zookeeper
	getCli    chan *Client         //获取缓存客户端结构
	putCli    chan *Client         //放回缓存客户端结构
	getMsg    chan *common.Message //获取缓存消息结构
	putMsg    chan *common.Message //放回缓存消息结构
)

//定义服务器结构
type Server struct {
	listener   net.Listener
	clients    map[int]*Client //客户端映射表
	lock       *sync.RWMutex   //map锁
	pending    chan net.Conn   //待处理连接channel
	quiting    chan *Client    //退出客户端channel
	activating chan *Client    //待激活客户端channel
}

//ms获取账户信息参数结构
type UserArgs struct {
	Username string //用户名或id
}

//ms拉取消息参数结构
type RArgs struct {
	Who   int   //拉取对象
	MaxId int64 //最大消息id
	Limit int   //拉取数量
}

//创建服务器
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

	//初始化redis连接池
	redisPool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", Conf.Redis)
			if err != nil {
				//redis连接失败直接报错
				panic(err.Error())
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	getCli, putCli = makeClientRecycler()
	getMsg, putMsg = common.MakeMessageRecycler()

	server.listen()
	return server
}

func (self *Server) listen() {
	go func() {
		redClient := redisPool.Get()
		defer redClient.Close()
		for {
			select {
			case msg := <-in:
				//新消息处理
				//获取客户端
				client, ok := self.clients[msg.Uid]
				if ok {
					ret := RetJson(0, CMD_S_MSG, "OK", *msg)
					client.out <- ret
				}
			case msg := <-out:
				//客户端需要发出去的消息，写到redis队列，由sendsrv分发
				s, _ := json.Marshal(*msg)
				redClient.Do("LPUSH", common.MSG_QUEUE_0, string(s))
				putMsg <- msg
			}
		}
	}()

	go func() {
		for {
			select {
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
	client := <-getCli
	client.Init(conn)
	//开一个gorouting处理客户端激活
	go func() {
		c := <-client.activating
		//c!=nil代表时激活，否则就结束本次激活goroute
		if c != nil {
			//ms列表有变化
			if isMsStop || msLen == 0 {
				ret := RetJson(ERR_CLIENT_AUTH_FAILED, CMD_S_AUTH, "认证失败，连接数据库服务器失败", nil)
				c.out <- ret
				return
			}

			msIdx := rand.Intn(msLen)

			userArgs := UserArgs{
				Username: c.username,
			}

			var user common.User
			err := msClients[msIdx].Call("MS.GetUser", userArgs, &user)
			if err != nil {
				fmt.Printf("conn server call MS GetUser failed\n")
			}

			if user.Id == 0 || user.Password != c.password {
				ret := RetJson(ERR_CLIENT_AUTH_FAILED, CMD_S_AUTH, "认证失败，账户密码错误", nil)
				c.out <- ret
				return
			}

			//认证通过
			c.id = user.Id
			c.ready = CLIENT_READY

			ret := RetJson(0, CMD_S_AUTH, "认证成功", nil)
			c.out <- ret

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

//退出，删除一个客户端
func (self *Server) quit(client *Client) {
	redClient := redisPool.Get()
	defer redClient.Close()
	if client == nil {
		return
	}
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
	putCli <- client
	fmt.Printf("client %d quited\n", client.id)
}

//激活一个客户端,如果重连，删除之前的客户端结构
func (self *Server) activate(client *Client) {
	redClient := redisPool.Get()
	defer redClient.Close()

	if client == nil {
		return
	}

	self.lock.Lock()
	//如果该用户已经连接一个终端，剔出
	if oldClient, ok := self.clients[client.id]; ok {
		//删除客户端在线信息
		_, err := redClient.Do("DEL", common.USER_ONLINE_PREFIX+strconv.Itoa(client.id))
		if err != nil {
			fmt.Printf("delete %d online map status failed\n", client.id)
		}
		_, err2 := redClient.Do("DEL", common.USER_ONLINE_HOST_PREFIX+strconv.Itoa(client.id))
		if err2 != nil {
			fmt.Printf("delete %d online host map status failed\n", client.id)
		}
		self.clients[client.id].Kickout()
		delete(self.clients, client.id)
		putCli <- oldClient
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

	//开启客户端自检
	client.CheckSelf()

	//下发未读信息，包括多少未读信息，以及最多最近1000条未读消息
	clientMaxMsgKey := USER_MAX_MSGID_PREFIX + strconv.Itoa(client.id)
	maxMsgId, _ := redis.Int64(redClient.Do("GET", clientMaxMsgKey))
	//获取当前生成的msg id
	maxGenId, _ := redis.Int64(redClient.Do("GET", clientMaxMsgKey))
	//获取多少未读信息
	offTotal := (int)(maxGenId - maxMsgId)
	//最多最近10条未读消息
	limit := 0
	if offTotal > 1000 {
		limit = 1000
		maxMsgId = maxGenId - 1000
	} else {
		limit = offTotal
	}

	args := RArgs{
		Who:   client.id,
		MaxId: maxMsgId,
		Limit: limit,
	}

	var msgs []common.Message

	msIdx := rand.Intn(msLen)
	err := msClients[msIdx].Call("MS.ReadMessages", args, &msgs)
	if err != nil {
		fmt.Printf("conn server call MS ReadMessages failed\n")
	}

	//离线消息发送到客户端
	for _, msg := range msgs {
		ret := RetJson(0, CMD_S_MSG, "OK", msg)
		client.out <- ret
	}
}

//服务器启动
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
	//为当前服务器创建一个节点，加入到集群中
	path := Conf.ZkRoot + "/" + Conf.RcpBind
	common.ZkCreateTempNode(zkConn, path)
	go func() {
		for {
			exist, _, watch, err := zkConn.ExistsW(path)
			if err != nil {
				//发生错误，当前节点退出
				common.KillSelf()
				return
			}

			if !exist {
				//节点不存在了
				common.KillSelf()
				return
			}

			event := <-watch
			fmt.Printf("%s receiver a event %v\n", path, event)
		}
	}()

	//获取存储子节点列表
	var size int
	children := common.ZkGetChildren(zkConn, Conf.MsZkRoot)
	if children != nil {
		//开启新的客户端
		for _, host := range children {
			msClients[size] = common.InitRpcClient(host)
			size++
		}

		msLen = size
		isMsStop = false
		fmt.Printf("ms len %d\n", msLen)
	}

	//开一个goroute处理ms服务器节点变化
	go func() {
		for {
			_, _, events, err := zkConn.ChildrenW(Conf.MsZkRoot)
			if err != nil {
				time.Sleep(time.Second * 1)
				continue
			}

			evt := <-events
			if evt.Type == zk.EventNodeChildrenChanged {
				// ms节点有变化，更新ms列表
				// 停止当前的ms处理
				isMsStop = true
				time.Sleep(time.Second * 1) //暂停一秒，让在处理中的ms处理完成

				size = 0
				//关闭原先ms客户端
				for _, mc := range msClients {
					if mc != nil {
						mc.Close()
						msClients[size] = nil
					}
					size++
				}

				//获取子节点列表
				children := common.ZkGetChildren(zkConn, Conf.MsZkRoot)
				if children == nil {
					fmt.Printf("no ms rpc servers\n")
					continue
				}

				size = 0
				//开启新的客户端
				for _, host := range children {
					msClients[size] = common.InitRpcClient(host)
					size++
				}

				msLen = size
				isMsStop = false
				fmt.Printf("ms len %d\n", msLen)
			}
		}
	}()
}

func (self *Server) Stop() {
	redisPool.Close()
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

	redClient := redisPool.Get()
	defer redClient.Close()
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
				b.lastAccTime = time.Now().Unix()
				queue.PushFront(b)
			case get <- ct.Value.(*Client):
				timeout.Stop()
				queue.Remove(ct)
			case <-timeout.C:
				ct := queue.Front()
				for ct != nil {
					n := ct.Next()
					if time.Now().Unix()-ct.Value.(*Client).lastAccTime > int64(time.Second*60) {
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
