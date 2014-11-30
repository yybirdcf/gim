package main

import (
	"runtime"
)

type Config struct {
	TcpBind   string
	PidFile   string
	User      string
	Dir       string
	MaxThread int
	MS        string
	Redis     string
	PushSrv   string
}

var (
	Conf *Config
)

func InitConfig() error {
	Conf = &Config{
		TcpBind:   "127.0.0.1:8380",
		PidFile:   "/tmp/gim-sendsrv.pid",
		User:      "nobody nobody",
		Dir:       "./",
		MaxThread: runtime.NumCPU(),
		MS:        "127.0.0.1:8680",
		Redis:     "127.0.0.1:6379",
		PushSrv:   "127.0.0.1:8980",
	}

	return nil
}
