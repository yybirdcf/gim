package sendsrv

type Config struct {
	TcpBind    string
	PidFile    string
	User       string
	Dir        string
	MaxClients int
}

var (
	Conf *Config
)

func InitConfig() error {
	Conf = &Config{
		TcpBind:    "127.0.0.1:8380",
		PidFile:    "/tmp/gim-sendsrv.pid",
		User:       "nobody nobody",
		Dir:        "./",
		MaxClients: 5,
	}

	return nil
}
