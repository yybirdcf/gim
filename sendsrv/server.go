package main

import (
	"encoding/json"
	"fmt"
	"gim/common"
	"github.com/garyburd/redigo/redis"
	"net/rpc"
)

var (
	msClient      *rpc.Client
	redClient     redis.Conn
	pushSrvClient *rpc.Client
	msgPool       chan *common.Message
)

func StartSendSrv() {
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

	msgPool = make(chan *common.Message, 4096)

	for i := 0; i < Conf.MaxThread; i++ {
		go func() {
			for {
				m, err := redis.String(redClient.Do("RPOP", "msg_queue_0"))
				if err == nil {
					var msg common.Message
					err = json.Unmarshal([]byte(m), &msg)
					if err == nil {
						HandleServerMsg(&msg)
					}
				}
			}
		}()
	}

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
}
