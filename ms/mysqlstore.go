package main

import (
	"database/sql"
	"fmt"
	"gim/common"
	_ "github.com/go-sql-driver/mysql"
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
		return nil
	}

	mysqlStore.db = db
	return mysqlStore
}

func (self *MysqlStore) Read(who int, maxId int64, limit int) []common.Message {
	rows, err := self.db.Query("SELECT msg_uid, msg_mid, msg_content, msg_type, msg_time, msg_from, msg_to, msg_group FROM message WHERE msg_uid=? AND msg_mid>? ORDER BY msg_mid ASC LIMIT "+strconv.Itoa(limit), who, maxId)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		panic(err.Error())
	}
	defer rows.Close()

	var ms []common.Message
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
		err = rows.Scan(&Uid, &Mid, &Content, &Type, &Time, &From, &To, &Group)
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
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(m.Uid, m.Mid, m.Content, m.Type, m.Time, m.From, m.To, m.Group)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return false
	}

	return true
}

func (self *MysqlStore) GetGroupMembers(groupId int) []int {
	rows, err := self.db.Query("SELECT user_id FROM user_group WHERE group_id=?", groupId)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return nil
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

func (self *MysqlStore) GetUser(username string) common.User {
	var user common.User

	rows, err := self.db.Query("SELECT id, username, password, avatar FROM user WHERE username=?", username)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return user
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&user.Id, &user.Username, &user.Password, &user.Avatar)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			panic(err.Error())
		}

		return user
	}

	return user
}
