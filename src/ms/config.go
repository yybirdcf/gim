package main

type Config struct {
	TcpBind string
	PidFile string
	User    string
	Dir     string
	Store   string
	Dsn     string
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
		Dsn:     "root:1160616612@tcp(127.0.0.1:3306)/chat?charset=utf8",
	}

	return nil
}
