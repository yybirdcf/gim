package main

import (
	"common"
	"fmt"
	"net/http"
	"net/rpc"
)

type RArgs struct {
	To    int
	MaxId int64
	Limit int
}

type MS struct {
	buf chan *common.Message
}

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
	msgs := store.Read(args.To, args.MaxId, args.Limit)
	*reply = msgs
	return nil
}

type GroupArgs struct {
	groupId int
}

func (self *MS) GetGroupMembers(args *GroupArgs, reply *[]int) error {
	*reply = store.GetGroupMembers(args.groupId)
	return nil
}

func StartMs() {
	ms := NewMS()
	rpc.Register(ms)
	rpc.HandleHTTP()

	err := http.ListenAndServe(Conf.TcpBind, nil)
	if err != nil {
		fmt.Printf("ms rpc error: %s\n", err.Error())
	}
}
