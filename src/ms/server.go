package main

import (
	"fmt"
	"net/http"
	"net/rpc"
)

type WArgs struct {
	Id   int64
	Msg  string
	Type int
	Time int64
	From int
	To   int
}

type RArgs struct {
	To    int
	MaxId int64
	Limit int
}

type MS struct {
	buf chan *Message
}

func NewMS() *MS {

	ms := &MS{
		buf: make(chan *Message, 1024),
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
func (self *MS) SaveMessage(args *WArgs, reply *bool) error {
	m := &Message{
		Id:   args.Id,
		Msg:  args.Msg,
		Type: args.Type,
		Time: args.Time,
		From: args.From,
		To:   args.To,
	}
	self.buf <- m

	*reply = true
	return nil
}

//获取消息调用
func (self *MS) ReadMessages(args *RArgs, reply *[]Message) error {
	msgs := store.Read(args.To, args.MaxId, args.Limit)
	*reply = msgs
	return nil
}

type UserArgs struct {
	id    int
	token string
}

//获取用户调用
func (self *MS) GetUser(args *UserArgs, reply *bool) error {
	*reply = store.IsUserValid(args.id, args.token)
	return nil
}

type ClientArgs struct {
	guid    string
	connSrv string
	userId  int
}

//申请客户端，新增client information
func (self *MS) NewClientInformation(args *ClientArgs, reply *bool) error {
	*reply = store.NewClientInfomation(args.guid, args.connSrv, args.userId)
	return nil
}

type ClientSim struct {
	guid string
}

//删除客户端申请信息
func (self *MS) DeleteClientInfomation(args *ClientSim, reply *bool) error {
	*reply = store.DeleteClientInformation(args.guid)
	return nil
}

//激活客户端
func (self *MS) ActiveClientInfomation(args *ClientArgs, reply *bool) error {
	*reply = store.ActiveClientInformation(args.guid, args.connSrv, args.userId)
	return nil
}

func StartMs() {
	ms := NewMS()
	rpc.Register(ms)
	rpc.HandleHTTP()

	err := http.ListenAndServe(Conf.HttpBind, nil)
	if err != nil {
		fmt.Printf("ms rpc error: %s\n", err.Error())
	}
}
