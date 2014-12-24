package main

import (
	"runtime"
)

type Config struct {
	TcpBind       string
	PidFile       string
	User          string
	Dir           string
	MaxThread     int
	ConnSrv       string
	Redis         string
	ZooKeeper     []string //sendsrv 和push servers
	ZkRoot        string
	ConnZooKeeper []string
	ConnZkRoot    string
}

var (
	Conf *Config
)

func InitConfig() error {
	Conf = &Config{
		TcpBind:   "127.0.0.1:8980",
		PidFile:   "/tmp/gim-pushsrv.pid",
		User:      "nobody nobody",
		Dir:       "./",
		MaxThread: runtime.NumCPU(),
		ConnSrv:   "127.0.0.1:8285",
		Redis:     "127.0.0.1:6379",
		ZooKeeper: []string{
			"127.0.0.1:2181",
		},
		ConnZooKeeper: []string{
			"127.0.0.1:2181",
		},
		ZkRoot:     "/pushsrv",
		ConnZkRoot: "/connsrv",
	}

	return nil
}
