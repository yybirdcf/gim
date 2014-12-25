package main

type Config struct {
	TcpBind   string
	PidFile   string
	User      string
	Dir       string
	Store     string
	Dsn       string
	ZooKeeper []string //sendsrv 和ms servers
	ZkRoot    string
}

var (
	Conf *Config
)

func InitConfig() error {
	Conf = &Config{
		TcpBind: "127.0.0.1:8680",
		PidFile: "/tmp/gim-ms.pid",
		User:    "nobody nobody",
		Dir:     "./",
		Store:   "mysql",
		Dsn:     "gotest:gotest@tcp(10.160.80.113:3306)/chat?charset=utf8",
		ZooKeeper: []string{
			"127.0.0.1:2181",
		},
		ZkRoot: "/ms",
	}

	return nil
}
