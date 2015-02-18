package main

type Config struct {
	TcpBind    string
	PidFile    string
	User       string
	Dir        string
	MaxClients int
	RcpBind    string
	Redis      string
	ZooKeeper  []string //conn srvs
	ZkRoot     string
	MsZkRoot   string
}

var (
	Conf *Config
)

func InitConfig() error {
	Conf = &Config{
		TcpBind:    ":8280",
		PidFile:    "/tmp/gim-connsrv.pid",
		User:       "nobody nobody",
		Dir:        "./",
		MaxClients: 50,
		RcpBind:    "127.0.0.1:8285",
		Redis:      "127.0.0.1:6379",
		ZooKeeper: []string{
			"127.0.0.1:2181",
		},
		ZkRoot:   "/connsrv",
		MsZkRoot: "/ms",
	}

	return nil
}
