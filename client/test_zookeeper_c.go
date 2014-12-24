package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

func main() {
	conn := connect()
	defer conn.Close()

	flags := int32(0)
	acl := zk.WorldACL(zk.PermAll)

	found, _, _, err := conn.ExistsW("/nodes")
	if err != nil {
		panic(err.Error())
	}

	if !found {
		fmt.Printf("create /nodes\n")
		path, err := conn.Create("/nodes", []byte("nodes"), flags, acl)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("created %s\n", path)
	}

	for {
		snapshot, _, events, err := conn.ChildrenW("/nodes")
		if err != nil {
			panic(err.Error())
			return
		}

		fmt.Printf("%+v ", snapshot)
		evt := <-events
		if evt.Err != nil {
			return
		}

		if evt.Type == zk.EventNodeChildrenChanged {
			fmt.Printf("nodes changed\n")
		}

		fmt.Printf(" is ok\n")
	}
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
