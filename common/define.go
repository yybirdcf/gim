package common

//redis前缀
const (
	APP_KEY    = "gim#test#key"
	APP_SECRET = "gim#test#key#secret"

	USER_ONLINE_PREFIX      = "useron#"
	USER_ONLINE_HOST_PREFIX = "userhoston#"
	RCP_TCP_HOST_PREFIX     = "rcp&tcp#"
	MSG_QUEUE_0             = "msg_queue_0"
)

type User struct {
	Id       int
	Username string
	Password string
	Avatar   string
}
