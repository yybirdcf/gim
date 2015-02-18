package main

import (
	"errors"
	"gim/common"
)

const (
	MYSQL_STORE_TYPE = "mysql"
)

var (
	store Store
)

type Store interface {
	Save(m *common.Message) bool
	Read(who int, maxId int64, limit int) []common.Message
	GetGroupMembers(groupId int) []int
	GetUser(username string) common.User
}

func InitStore() error {
	if Conf.Store == MYSQL_STORE_TYPE {
		store = NewMysqlStore()
	} else {
		errors.New("unknown storage type")
	}

	return nil
}
