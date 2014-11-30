package main

type Config struct {
	TcpBind    string
	PidFile    string
	User       string
	Dir        string
	MaxClients int
	SendSrvTcp string
	RcpBind    string
	MS         string
	Redis      string
}

var (
	Conf *Config
)

func InitConfig() error {
	Conf = &Config{
		TcpBind:    "127.0.0.1:8280",
		PidFile:    "/tmp/gim-connsrv.pid",
		User:       "nobody nobody",
		Dir:        "./",
		MaxClients: 50,
		SendSrvTcp: "127.0.0.1:8380",
		RcpBind:    "127.0.0.1:8285",
		MS:         "127.0.0.1:8680",
		Redis:      "127.0.0.1:6379",
	}

	return nil
}
