package main

import (
	"fmt"
	"gim/common"
	"net/rpc"
	"time"
)

func main() {
	server := "127.0.0.1:8980"
	client, err := rpc.DialHTTP("tcp", server)
	if err != nil {
		fmt.Printf("pushsrv test, connect %s failed\n", server)
		return
	}

	var reply_b bool

	for {
		args := common.Message{
			Mid:     int(time.Now().Unix()),
			Uid:     1000,
			Content: "test message from 10001 to 1000",
			Type:    common.MESSAGE_TYPE_USER,
			Time:    int(time.Now().Unix()),
			From:    10001,
			To:      1000,
			Group:   0,
		}

		err = client.Call("PushSrv.SendMsg", args, &reply_b)
		if err != nil {
			fmt.Printf("pushsrv test, call PushSrv.SendMsg failed: %s\n", err.Error())
			return
		}

		fmt.Printf("send :%v\n", args)

		time.Sleep(2 * time.Second)
	}
}
