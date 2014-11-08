package main

import (
	"fmt"
	"net/http"
	"net/rpc"
)

type WArgs struct {
	Id   int64
	Msg  string
	Type int
	Time int64
	From int
	To   int
}

type RArgs struct {
	To    int
	MaxId int64
	Limit int
}

type MS struct {
	store *Store
	buf   chan *Message
}

func NewMS() *MS {

	ms := &MS{
		store: NewStore(Conf.Dsn),
		buf:   make(chan *Message, 1024),
	}

	go func() {
		for {
			m := <-ms.buf
			if m != nil {
				b := ms.store.Save(m)
				if b {
					fmt.Printf("save message success: %v", m)
				} else {
					fmt.Printf("save message failed: %v", m)
				}
			}
		}
	}()

	return ms
}

func (self *MS) SaveMessage(args *WArgs, reply *bool) error {
	m := &Message{
		Id:   args.Id,
		Msg:  args.Msg,
		Type: args.Type,
		Time: args.Time,
		From: args.From,
		To:   args.To,
	}
	self.buf <- m

	*reply = true
	return nil
}

func (self *MS) ReadMessages(args *RArgs, reply *[]*Message) error {
	*reply = self.store.Read(args.To, args.MaxId, args.Limit)

	return nil
}

func StartMs() {
	ms := NewMS()
	rpc.Register(ms)
	rpc.HandleHTTP()

	err := http.ListenAndServe(Conf.HttpBind, nil)
	if err != nil {
		fmt.Printf("ms rpc error: %s\n", err.Error())
	}
}
