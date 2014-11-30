package main

import (
	"fmt"
	"gim/common"
	"github.com/garyburd/redigo/redis"
	"net/http"
	"net/rpc"
)

var (
	msClient      *rpc.Client
	redClient     redis.Conn
	pushSrvClient *rpc.Client
)

type SendSrv struct {
	buf     chan *common.Message
	msgPool chan *common.Message
}

func NewSendSrv() *SendSrv {

	ss := &SendSrv{
		buf:     make(chan *common.Message, 2048),
		msgPool: make(chan *common.Message, 2048),
	}

	for i := 0; i < Conf.MaxThread; i++ {
		go func() {
			for {
				s := <-ss.buf
				if s != nil {
					HandleServerMsg(ss, s)
				}
			}
		}()
	}

	client, err := rpc.DialHTTP("tcp", Conf.MS)
	if err != nil {
		panic(err.Error())
	}
	msClient = client

	pclient, err := rpc.DialHTTP("tcp", Conf.PushSrv)
	if err != nil {
		panic(err.Error())
	}
	pushSrvClient = pclient

	conn, _ := redis.Dial("tcp", Conf.Redis)
	redClient = conn
	go func() {
		for {
			if m := <-msgPool; m != nil {
				//MS落地存储
				var success bool
				err = msClient.Call("MS.SaveMessage", *m, &success)
				if err != nil || !success {
					fmt.Printf("send server call MS SaveMessage failed")
				}
			}
		}
	}()

	return ms
}

func (self *SendSrv) SendMsg(msg *common.Message, reply *bool) error {
	if m != nil {
		self.buf <- m
	}
}

func StartSendSrv() {
	ss := NewSendSrv()
	rpc.Register(ss)
	rpc.HandleHTTP()

	err := http.ListenAndServe(Conf.TcpBind, nil)
	if err != nil {
		fmt.Printf("sendsrv rpc error: %s\n", err.Error())
	}
}
