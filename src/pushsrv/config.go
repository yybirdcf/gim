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
	ConnSrv   string
	Redis     string
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
		MS:        "127.0.0.1:8680",
		ConnSrv:   "127.0.0.1:8280",
		Redis:     "127.0.0.1:6379",
	}

	return nil
}
