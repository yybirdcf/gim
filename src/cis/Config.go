package main

type Config struct {
	HttpBind string
	PidFile  string
	User     string
	Dir      string
}

var (
	Conf *Config
)

func InitConfig() error {
	Conf = &Config{
		TcpBind: "127.0.0.1:8580",
		PidFile: "/tmp/gim-cis.pid",
		User:    "nobody nobody",
		Dir:     "./",
	}

	return nil
}
