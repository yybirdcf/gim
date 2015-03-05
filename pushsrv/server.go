package main

import (
	"fmt"
	"gim/common"
	"github.com/garyburd/redigo/redis"
	"github.com/samuel/go-zookeeper/zk"
	"net/http"
	"net/rpc"
	"strconv"
	"time"
)

const (
	USER_ONLINE_PREFIX      = "useron#"
	USER_ONLINE_HOST_PREFIX = "userhoston#"
)

var (
	connClient     *rpc.Client
	redClient      redis.Conn
	zkConn         *zk.Conn
	connZkConn     *zk.Conn
	avalConnClient map[string]*rpc.Client //可用conn srv列表
	isStopSecond   bool
)

type PushSrv struct {
	buf chan *common.Message
}

func (self *PushSrv) SendMsg(args *common.Message, reply *bool) error {
	self.buf <- args
	*reply = true
	return nil
}

func NewPushSrv() *PushSrv {
	isStopSecond = true
	ps := &PushSrv{
		buf: make(chan *common.Message, 2048),
	}

	avalConnClient = make(map[string]*rpc.Client)

	conn, _ := redis.Dial("tcp", Conf.Redis)
	redClient = conn

	for i := 0; i < Conf.MaxThread; i++ {
		go func() {
			for {
				if isStopSecond {
					time.Sleep(time.Second * 1)
					continue
				}
				s := <-ps.buf
				if s != nil {
					//分发消息
					dispatchMsg(s)
				}
			}
		}()
	}

	return ps
}

//connect server or apns
func dispatchMsg(m *common.Message) {
	//用户是否在线
	exist, _ := redis.Int(redClient.Do("GET", USER_ONLINE_PREFIX+strconv.Itoa(m.Uid)))
	//用户长连接机器是否在可用机器列表
	host, _ := redis.String(redClient.Do("GET", USER_ONLINE_HOST_PREFIX+strconv.Itoa(m.Uid)))
	var ok bool
	connClient, ok = avalConnClient[host]
	if exist > 0 && ok {
		err := connClient.Call("ConnRpcServer.SendMsg", *m, &ok)
		if err != nil {
			fmt.Printf("call conn server rpc failed %s\n", err.Error())
		}
	} else {
		//apns
	}
}

func StartPushSrv() {
	ps := NewPushSrv()
	rpc.Register(ps)
	rpc.HandleHTTP()

	go func() {
		err := http.ListenAndServe(Conf.TcpBind, nil)
		if err != nil {
			fmt.Printf("pushsrv rpc error: %s\n", err.Error())
		}
	}()

	//服务器初始化完成以后，开启zookeeper
	zkConn = common.ZkConnect(Conf.ZooKeeper)
	common.ZkCreateRoot(zkConn, Conf.ZkRoot)
	//为当前push服务器创建一个节点，加入到push集群中
	path := Conf.ZkRoot + "/" + Conf.TcpBind
	common.ZkCreateTempNode(zkConn, path)
	go func() {
		exist, _, watch, err := zkConn.ExistsW(path)
		if err != nil {
			//发生错误，当前节点退出
			fmt.Printf("%s occur error\n", path)
			common.KillSelf()
		}

		if !exist {
			//节点不存在了
			fmt.Printf("%s not exist\n", path)
			common.KillSelf()
		}

		event := <-watch
		fmt.Printf("%s receiver a event %v\n", path, event)
	}()

	//初始化conn srv可用列表
	connZkConn := common.ZkConnect(Conf.ConnZooKeeper)
	children := common.ZkGetChildren(connZkConn, Conf.ConnZkRoot)
	if children != nil {
		//更新可用客户端列表
		for _, host := range children {
			avalConnClient[host] = initRpcClient(host)
		}
	}

	//开一个goroute处理conn srv服务器节点变化
	go func() {
		for {
			_, _, events, err := connZkConn.ChildrenW(Conf.ConnZkRoot)
			if err != nil {
				time.Sleep(time.Second * 1)
				continue
			}

			evt := <-events
			if evt.Type == zk.EventNodeChildrenChanged {
				//conn节点有变化，更新conn列表
				isStopSecond = true
				time.Sleep(time.Second * 1)

				for host, rcpCli := range avalConnClient {
					rcpCli.Close()
					delete(avalConnClient, host)
				}

				children := common.ZkGetChildren(connZkConn, Conf.ConnZkRoot)
				for _, host := range children {
					avalConnClient[host] = initRpcClient(host)
				}

				isStopSecond = false
			}
		}
	}()

	//可以开始处理消息
	isStopSecond = false
}

func ClosePs() {
	zkConn.Delete(Conf.ZkRoot+"/"+Conf.TcpBind, 0)
	zkConn.Close()
}

func initRpcClient(host string) *rpc.Client {
	client, err := rpc.DialHTTP("tcp", host)
	if err != nil {
		panic(err.Error())
	}
	return client
}
