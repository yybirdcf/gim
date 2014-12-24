package main

type Config struct {
	HttpBind  []string
	PidFile   string
	User      string
	Dir       string
	Redis     string
	Ms        string
	ZooKeeper []string //conn srvs
	ZkRoot    string
}

var (
	Conf *Config
)

func InitConfig() error {
	Conf = &Config{
		HttpBind: []string{"127.0.0.1:8180"},
		PidFile:  "/tmp/gim-web.pid",
		User:     "nobody nobody",
		Dir:      "./",
		Ms:       "127.0.0.1:8680",
		Redis:    "127.0.0.1:6379",
		ZooKeeper: []string{
			"127.0.0.1:2181",
		},
		ZkRoot: "/connsrv",
	}

	return nil
}
