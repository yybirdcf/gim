package main

import (
	"fmt"
	"net/rpc"
	"time"
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

type Message struct {
	Id   int64
	Msg  string
	Type int
	Time int64
	From int
	To   int
}

func main() {
	server := "127.0.0.1:8680"
	client, err := rpc.DialHTTP("tcp", server)
	if err != nil {
		fmt.Printf("ms test, connect %s failed\n", server)
		return
	}

	args := WArgs{
		Id:   time.Now().Unix(),
		Msg:  "message from 1 to 2",
		Type: 1,
		Time: time.Now().Unix(),
		From: 1,
		To:   2,
	}

	var reply_b bool
	err = client.Call("MS.SaveMessage", args, &reply_b)
	if err != nil {
		fmt.Printf("MS test, call MS.SaveMessage failed: %s\n", err.Error())
		return
	}

	args2 := RArgs{
		To:    2,
		MaxId: 0,
		Limit: 10,
	}
	var reply_messages []Message
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
