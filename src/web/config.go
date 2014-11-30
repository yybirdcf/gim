package main

type Config struct {
	HttpBind []string
	PidFile  string
	User     string
	Dir      string
	SendSrv  string
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
		SendSrv:  "127.0.0.1:8380",
	}

	return nil
}
