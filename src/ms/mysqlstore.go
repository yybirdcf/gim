package main

import (
	"database/sql"
	"fmt"
	_ "mysql"
	"strconv"
	"time"
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

func (self *MysqlStore) Read(to int, maxId int64, limit int) []Message {
	rows, err := self.db.Query("SELECT msg_uid, msg_mid, msg_content, msg_type, msg_time, msg_from, msg_to, msg_group FROM message WHERE to=? AND mid>? ORDER BY mid ASC LIMIT "+strconv.Itoa(limit), to, maxId)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		panic(err.Error())
	}
	defer rows.Close()

	var ms []Message
	for rows.Next() {
		var (
			Mid     int
			Uid     int
			Content string
			Type    int
			Time    int
			From    int
			To      int
			Group   int
		)
		err = rows.Scan(&Mid, &Uid, &Content, &Type, &Time, &From, &To, &Group)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			panic(err.Error())
		}

		ms = append(ms, common.Message{
			Mid:     Mid,
			Uid:     Uid,
			Content: Content,
			Type:    Type,
			Time:    Time,
			From:    From,
			To:      To,
			Group:   Group,
		})
	}

	return ms
}

func (self *MysqlStore) Save(m *common.Message) bool {
	stmt, err := self.db.Prepare("INSERT INTO message (msg_uid, msg_mid, msg_content, msg_type, msg_time, msg_from, msg_to, msg_group) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		panic(err.Error())
	}
	defer stmt.Close()

	res, err := stmt.Exec(m.Uid, m.Mid, m.Content, m.Type, m.Time, m.From, m.To, m.Group)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		panic(err.Error())
		return false
	}

	affect, err := res.RowsAffected()
	fmt.Printf("rows affect %d\n", affect)
	return true
}

func (self *MysqlStore) GetGroupMembers(groupId int) []int {
	rows, err := self.db.Query("SELECT user FROM group WHERE group=?", groupId)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		panic(err.Error())
	}
	defer rows.Close()

	var members []int
	for rows.Next() {
		var (
			user int
		)
		err = rows.Scan(&user)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			panic(err.Error())
		}

		members = append(members, user)
	}

	return members
}
