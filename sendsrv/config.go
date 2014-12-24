package main

import (
	"runtime"
)

type Config struct {
	PidFile   string
	User      string
	Dir       string
	MaxThread int
	Redis     string
	ZooKeeper []string
	MsZkRoot  string
	PsZkRoot  string
}

var (
	Conf *Config
)

func InitConfig() error {
	Conf = &Config{
		PidFile:   "/tmp/gim-sendsrv.pid",
		User:      "nobody nobody",
		Dir:       "./",
		MaxThread: runtime.NumCPU(),
		Redis:     "127.0.0.1:6379",
		ZooKeeper: []string{
			"127.0.0.1:2181",
		},
		MsZkRoot: "/ms",
		PsZkRoot: "/pushsrv",
	}

	return nil
}
