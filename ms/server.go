package main

import (
	"fmt"
	"gim/common"
	"github.com/samuel/go-zookeeper/zk"
	"net/http"
	"net/rpc"
)

type RArgs struct {
	Who   int
	MaxId int64
	Limit int
}

type MS struct {
	buf chan *common.Message
}

var (
	zkConn *zk.Conn
)

func NewMS() *MS {

	ms := &MS{
		buf: make(chan *common.Message, 4096),
	}

	InitStore()

	go func() {
		for {
			m := <-ms.buf
			if m != nil {
				b := store.Save(m)
				if b {
					fmt.Printf("save message success: %v", m)
				} else {
					fmt.Printf("save message failed: %v", m)
				}
			}
		}
	}()

	return ms
}

//保存消息调用
func (self *MS) SaveMessage(args *common.Message, reply *bool) error {
	self.buf <- args

	*reply = true
	return nil
}

//获取消息调用
func (self *MS) ReadMessages(args *RArgs, reply *[]common.Message) error {
	msgs := store.Read(args.Who, args.MaxId, args.Limit)
	*reply = msgs
	return nil
}

type GroupArgs struct {
	GroupId int
}

func (self *MS) GetGroupMembers(args *GroupArgs, reply *[]int) error {
	*reply = store.GetGroupMembers(args.GroupId)
	return nil
}

func StartMs() {

	ms := NewMS()
	rpc.Register(ms)
	rpc.HandleHTTP()

	go func() {
		err := http.ListenAndServe(Conf.TcpBind, nil)
		if err != nil {
			fmt.Printf("ms rpc error: %s\n", err.Error())
		}
	}()

	//服务器初始化完成以后，开启zookeeper
	zkConn = common.ZkConnect(Conf.ZooKeeper)
	common.ZkCreateRoot(zkConn, Conf.ZkRoot)
	//为当前ms服务器创建一个节点，加入到ms集群中
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
}

func CloseMs() {
	zkConn.Delete(Conf.ZkRoot+"/"+Conf.TcpBind, 0)
	zkConn.Close()
}
