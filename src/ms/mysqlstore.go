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
	rows, err := self.db.Query("SELECT * FROM message WHERE msg_to=? AND msg_id>? ORDER BY msg_id ASC LIMIT "+strconv.Itoa(limit), to, maxId)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		panic(err.Error())
	}
	defer rows.Close()

	var ms []Message
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
			fmt.Printf("%s\n", err.Error())
			panic(err.Error())
		}

		ms = append(ms, Message{
			Id:   Id,
			Msg:  Msg,
			Type: Type,
			Time: Time,
			From: From,
			To:   To,
		})
	}

	return ms
}

func (self *MysqlStore) Save(m *Message) bool {
	stmt, err := self.db.Prepare("INSERT INTO message (msg_id, msg_content, msg_type, msg_time, msg_from, msg_to) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		panic(err.Error())
	}
	defer stmt.Close()

	res, err := stmt.Exec(m.Id, m.Msg, m.Type, m.Time, m.From, m.To)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
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

func (self *MysqlStore) IsUserValid(id int, token string) bool {
	row, err := self.db.QueryRow("SELECT * FROM User WHERE id=? AND token=? LIMIT 1", id, token)
	defer row.Close()
	if err != nil && err == sql.ErrNoRows {
		return false
	} else if err != nil {
		fmt.Printf("%s\n", err.Error())
		panic(err.Error())
	} else {
		return true
	}
}

func (self *MysqlStore) NewClientInfomation(guid string, connSrv string, userId int) bool {
	stmt, err := self.db.Prepare("INSERT INTO client_information (guid, connSrv, genTime, userId) VALUES (?, ?, ?, ?)")
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		panic(err.Error())
	}
	defer stmt.Close()

	res, err := stmt.Exec(guid, connSrv, time.Now().Unix(), userId)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		panic(err.Error())
		return false
	}

	affect, err := res.RowsAffected()
	fmt.Printf("rows affect %d\n", affect)
	return true
}

func (self *MysqlStore) DeleteClientInformation(guid string) bool {
	stmt, err := self.db.Prepare("DELETE FROM client_information WHERE guid=?")
	defer stmt.Close()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		panic(err.Error())
	}

	res, err := stmt.Exec(guid)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		panic(err.Error())
	}

	return true
}

func (self *MysqlStore) ActiveClientInformation(guid string, connSrv string, userId int) bool {
	stmt, err := self.db.Prepare("UPDATE client_information SET status=1 WHERE guid=? AND userId=? AND connSrv=?")

	defer stmt.Close()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		panic(err.Error())
	}

	res, err := stmt.Exec(guid, userId, connSrv)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		panic(err.Error())
	}

	rowCnt, err := res.RowsAffected()
	if rowCnt > 0 {
		return true
	}
	return false
}
