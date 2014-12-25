package main

import (
	"encoding/json"
	"fmt"
	"gim/common"
	"github.com/garyburd/redigo/redis"
	"github.com/samuel/go-zookeeper/zk"
	"math/rand"
	"net/rpc"
	"time"
)

var (
	msClients      [20]*rpc.Client //ms rcp客户端列表
	redClient      redis.Conn
	pushSrvClients [20]*rpc.Client //push rcp客户端列表
	msgPool        chan *common.Message
	zkConn         *zk.Conn
	isMsStop       bool
	isPsStop       bool
	msLen          int
	psLen          int
	msgGet         chan *common.Message
	msgPut         chan *common.Message
)

func StartSendSrv() {
	isMsStop = true
	isPsStop = true
	msLen = 0
	psLen = 0

	rand.Seed(time.Now().UnixNano())

	conn, err := redis.Dial("tcp", Conf.Redis)
	if err != nil {
		fmt.Printf("tcp redis %v\n", err.Error())
		return
	}
	redClient = conn

	msgGet, msgPut = common.MakeMessageRecycler()

	msgPool = make(chan *common.Message, 4096)

	go func() {
		msIdx := 0
		psIdx := 0
		for {
			//ms列表有变化
			if isMsStop || msLen == 0 || isPsStop || psLen == 0 {
				time.Sleep(time.Second * 1)
				continue
			}
			msIdx = rand.Intn(msLen)
			psIdx = rand.Intn(psLen)
			m, err := redis.String(redClient.Do("RPOP", "msg_queue_0"))
			if m == "" {
				time.Sleep(time.Second)
				continue
			}
			if err == nil {
				var msg common.Message
				err = json.Unmarshal([]byte(m), &msg)
				if err == nil {
					HandleServerMsg(&msg, msClients[msIdx], pushSrvClients[psIdx])
				}
			}
		}
	}()

	go func() {
		var success bool
		msIdx := 0
		for {
			//ms列表有变化
			if isMsStop || msLen == 0 {
				time.Sleep(time.Second * 1)
				continue
			}

			if m := <-msgPool; m != nil {
				//MS落地存储
				msIdx = rand.Intn(msLen)
				err := msClients[msIdx].Call("MS.SaveMessage", *m, &success)
				if err != nil || !success {
					fmt.Printf("send server call MS SaveMessage failed")
				}
				msgPut <- m
			}
		}
	}()

	//服务器初始化完成，启动zookeeper
	zkConn = common.ZkConnect(Conf.ZooKeeper)
	size := 0
	//获取子节点列表
	children := common.ZkGetChildren(zkConn, Conf.MsZkRoot)
	if children != nil {
		//开启新的客户端
		for _, host := range children {
			msClients[size] = initRpcClient(host)
			size++
		}

		msLen = size
		isMsStop = false
		fmt.Printf("ms len %d\n", msLen)
	}
	//获取push子节点列表
	psSize := 0
	psChildren := common.ZkGetChildren(zkConn, Conf.PsZkRoot)
	if psChildren != nil {
		//开启新的客户端
		for _, psHost := range psChildren {
			pushSrvClients[psSize] = initRpcClient(psHost)
			psSize++
		}

		psLen = psSize
		isPsStop = false
		fmt.Printf("ps len %d\n", psLen)
	}

	//开一个goroute处理ms服务器节点变化
	go func() {
		for {
			_, _, events, err := zkConn.ChildrenW(Conf.MsZkRoot)
			if err != nil {
				panic(err.Error())
				return
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
					msClients[size] = initRpcClient(host)
					size++
				}

				msLen = size
				isMsStop = false
				fmt.Printf("ms len %d\n", msLen)
			}
		}
	}()

	//开一个goroute处理push服务器节点变化
	go func() {
		for {
			_, _, events, err := zkConn.ChildrenW(Conf.PsZkRoot)
			if err != nil {
				panic(err.Error())
				return
			}

			evt := <-events
			if evt.Type == zk.EventNodeChildrenChanged {
				// push节点有变化，更新push列表
				// 停止当前的push处理
				isPsStop = true
				time.Sleep(time.Second * 1) //暂停一秒，让在处理中的push处理完成

				psSize = 0
				//关闭原先push客户端
				for _, psc := range pushSrvClients {
					if psc != nil {
						psc.Close()
						pushSrvClients[psSize] = nil
					}
					psSize++
				}

				//获取子节点列表
				children := common.ZkGetChildren(zkConn, Conf.PsZkRoot)
				if children == nil {
					fmt.Printf("no ps rpc servers\n")
					continue
				}

				psSize = 0
				//开启新的客户端
				for _, host := range children {
					pushSrvClients[psSize] = initRpcClient(host)
					psSize++
				}

				psLen = psSize
				isPsStop = false
				fmt.Printf("ps len %d\n", psLen)
			}
		}
	}()
}

func initRpcClient(host string) *rpc.Client {
	client, err := rpc.DialHTTP("tcp", host)
	if err != nil {
		panic(err.Error())
	}
	return client
}

func CloseSendSrv() {
	zkConn.Close()
}
