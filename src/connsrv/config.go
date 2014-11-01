package main

type Config struct {
	TcpBind    string
	PidFile    string
	User       string
	Dir        string
	MaxClients int
	SendSrvTcp string
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
	}

	return nil
}
