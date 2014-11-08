package main

import (
	"errors"
)

type Message struct {
	Id   int64
	Msg  string
	Type int
	Time int64
	From int
	To   int
}

const (
	MYSQL_STORE_TYPE = "mysql"
)

var (
	store Store
)

type Store interface {
	Save(m *Message) bool
	Read(to int, maxId int64, limit int) []Message
}

func InitStore() error {
	if Conf.Store == MYSQL_STORE_TYPE {
		store = NewMysqlStore()
	} else {
		errors.New("unknown storage type")
	}

	return nil
}
