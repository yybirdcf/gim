package main

import (
	"fmt"
	"gim/common"
	"github.com/astaxie/goredis"
	"net/http"
	"net/rpc"
	"sync"
)

const (
	USER_ONLINE_PREFIX = "useron#"
)

var (
	connClient *rpc.Client
	redClient  goredis.Client
)

type PushSrv struct {
	buf chan *common.Message
}

func (self *PushSrv) SendMsg(args *common.Message, reply *bool) error {
	m := &Message{
		Mid:     args.Mid,
		Uid:     args.Uid,
		Content: args.Content,
		Type:    args.Type,
		Time:    args.Time,
		From:    args.From,
		To:      args.To,
		Group:   args.Group,
	}

	self.buf <- m
	return nil
}

func NewPushSrv() *PushSrv {
	ps := PushSrv{
		buf: make(chan *common.Message, 2048),
	}

	client, err := rpc.DialHTTP("tcp", Conf.ConnSrv)
	if err != nil {
		panic(err.Error())
		return
	}
	connClient = client

	redClient.Addr = Conf.Redis

	for i := 0; i < Conf.MaxThread; i++ {
		go func() {
			for {
				s := <-ps.buf
				if s != nil {
					//分发消息
					dispathMsg(s)
				}
			}
		}()
	}

	return ps
}

//connect server or apns
func dispatchMsg(m *common.Message) {
	exist, _ := redClient.Get(USER_ONLINE_PREFIX + m.Uid)
	if exist != nil {
		var ok bool
		connClient.Call("Server.SendMsg", *m, &ok)
	} else {
		//apns
	}
}

func StartPushSrv() {
	ps := NewPushSrv()
	rpc.Register(ps)
	rpc.HandleHTTP()

	err := http.ListenAndServe(Conf.TcpBind, nil)
	if err != nil {
		fmt.Printf("pushsrv rpc error: %s\n", err.Error())
	}
}
