package main

import (
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

func main() {
	acl := zk.WorldACL(zk.PermAll)
	flags := int32(zk.FlagEphemeral) //临时节点不能有children
	conn2 := connect()
	defer conn2.Close()
	_, err := conn2.Create("/nodes/192.168.1.100:8309", []byte("192.168.1.100:8309"), flags, acl)
	if err != nil {
		panic(err.Error())
	}
	time.Sleep(time.Second)

	_, err = conn2.Create("/nodes/192.168.1.101:8309", []byte("192.168.1.101:8309"), flags, acl)
	if err != nil {
		panic(err.Error())
	}

	time.Sleep(10 * time.Second)
}

func connect() *zk.Conn {
	zks := []string{
		"127.0.0.1:2181",
	}

	conn, _, err := zk.Connect(zks, time.Second)
	if err != nil {
		panic(err.Error())
	}

	return conn
}
