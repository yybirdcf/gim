package main

import (
	"fmt"
	"gim/common"
	"github.com/garyburd/redigo/redis"
	"net/http"
	"net/rpc"
	"strconv"
)

const (
	USER_ONLINE_PREFIX = "useron#"
)

var (
	connClient *rpc.Client
	redClient  redis.Conn
)

type PushSrv struct {
	buf chan *common.Message
}

func (self *PushSrv) SendMsg(args *common.Message, reply *bool) error {
	self.buf <- args
	return nil
}

func NewPushSrv() *PushSrv {
	ps := &PushSrv{
		buf: make(chan *common.Message, 2048),
	}

	client, err := rpc.DialHTTP("tcp", Conf.ConnSrv)
	if err != nil {
		panic(err.Error())
		return nil
	}
	connClient = client

	conn, _ := redis.Dial("tcp", Conf.Redis)
	redClient = conn

	for i := 0; i < Conf.MaxThread; i++ {
		go func() {
			for {
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
	exist, _ := redClient.Do("GET", USER_ONLINE_PREFIX+strconv.Itoa(m.Uid))
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
