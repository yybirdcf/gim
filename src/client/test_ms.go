package main

import (
	"common"
	"fmt"
	"net/rpc"
	"time"
)

type RArgs struct {
	To    int
	MaxId int64
	Limit int
}

func main() {
	server := "127.0.0.1:8680"
	client, err := rpc.DialHTTP("tcp", server)
	if err != nil {
		fmt.Printf("ms test, connect %s failed\n", server)
		return
	}

	args := common.Message{
		Mid:     time.Now().Unix(),
		Uid:     1000,
		Content: "test message",
		Type:    common.MESSAGE_TYPE_USER,
		Time:    time.Now().Unix(),
		From:    1,
		To:      1000,
		Group:   0,
	}

	var reply_b bool
	err = client.Call("MS.SaveMessage", args, &reply_b)
	if err != nil {
		fmt.Printf("MS test, call MS.SaveMessage failed: %s\n", err.Error())
		return
	}

	args2 := RArgs{
		Who:   1000,
		MaxId: 0,
		Limit: 10,
	}
	var reply_messages []common.Message
	err = client.Call("MS.ReadMessages", args2, &reply_messages)
	if err != nil {
		fmt.Printf("MS test, call MS.ReadMessages failed: %s\n", err.Error())
		return
	}

	fmt.Printf("ok\n")
	for _, m := range reply_messages {
		fmt.Printf("%v\n", m)
	}
}
