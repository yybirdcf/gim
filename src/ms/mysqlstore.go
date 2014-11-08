package main

import (
	"database/sql"
	"fmt"
	_ "mysql"
	"strconv"
)

type MysqlStore struct {
	db *sql.DB
}

func NewMysqlStore() *MysqlStore {
	mysqlStore := &MysqlStore{}

	db, err := sql.Open("mysql", Conf.Dsn)
	if err != nil {
		panic(err.Error())
	}

	mysqlStore.db = db
	return mysqlStore
}

func (self *MysqlStore) Read(to int, maxId int64, limit int) []*Message {
	rows, err := self.db.Query("SELECT * FROM message WHERE to=? AND id>? ORDER BY id ASC LIMIT "+strconv.Itoa(limit), to, maxId)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	ms := make([]*Message, limit)
	i := 0
	for rows.Next() {
		var (
			Id   int64
			Msg  string
			Type int
			Time int64
			From int
			To   int
		)
		err = rows.Scan(&Id, &Msg, &Type, &Time, &From, &To)
		if err != nil {
			panic(err.Error())
		}

		ms[i] = &Message{
			Id:   Id,
			Msg:  Msg,
			Type: Type,
			Time: Time,
			From: From,
			To:   To,
		}
	}

	return ms
}

func (self *MysqlStore) Save(m *Message) bool {
	stmt, err := self.db.Prepare("INSERT INTO message (id, msg, type, time, from, to) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		panic(err.Error())
	}
	defer stmt.Close()

	res, err := stmt.Exec(m.Id, m.Msg, m.Type, m.Time, m.From, m.To)
	if err != nil {
		panic(err.Error())
		return false
	}

	affect, err := res.RowsAffected()
	fmt.Printf("rows affect %d\n", affect)
	return true
}

func (self *MysqlStore) Close() {
	self.db.Close()
}
