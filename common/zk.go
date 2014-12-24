package common

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

func ZkConnect(zks []string) *zk.Conn {
	conn, _, err := zk.Connect(zks, time.Second)
	if err != nil {
		panic(err.Error())
	}

	return conn
}

func ZkCreateRoot(zkConn *zk.Conn, root string) {
	flags := int32(0)
	acl := zk.WorldACL(zk.PermAll)

	found, _, _, err := zkConn.ExistsW(root)
	if err != nil {
		panic(err.Error())
	}

	if !found {
		fmt.Printf("create %s\n", root)
		path, err := zkConn.Create(root, []byte(""), flags, acl)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("created %s\n", path)
	}
}

func ZkCreateTempNode(zkConn *zk.Conn, path string) {
	flags := int32(zk.FlagEphemeral)
	acl := zk.WorldACL(zk.PermAll)

	fmt.Printf("%s before create \n", path)
	p, err := zkConn.Create(path, []byte(""), flags, acl)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("%s after create \n", p)
}

func ZkGetChildren(zkConn *zk.Conn, root string) []string {
	children, stat, err := zkConn.Children(root)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil
		}
		fmt.Printf("zk.ChildrenW(\"%s\") error(%v)", root, err)
		return nil
	}
	if stat == nil {
		return nil
	}
	if len(children) == 0 {
		return nil
	}

	return children
}
