package main

type Config struct {
	HttpBind []string
	PidFile  string
}

var (
	Conf *Config
)

func InitConfig() error {
	Conf = &Config{
		HttpBind: {"127.0.0.1:8180"},
		PidFile:  "/tmp/gim-web.pid",
		User:     "nobody nobody",
		Dir:      "./",
	}

	return nil
}
