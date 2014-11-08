package main

type Message struct {
	Id   int64
	Msg  string
	Type int
	Time int64
	From int
	To   int
}

type Store interface {
	Save(m *Message) bool
	Read(to int, maxId int64, limit int) []*Message
}
