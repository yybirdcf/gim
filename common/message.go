package common

const (
	MESSAGE_TYPE_PUBLIC = 1
	MESSAGE_TYPE_SUB    = 2
	MESSAGE_TYPE_GROUP  = 3
	MESSAGE_TYPE_USER   = 4
)

type Message struct {
	Mid     int
	Uid     int
	Content string
	Type    int
	Time    int
	From    int
	To      int
	Group   int
}
