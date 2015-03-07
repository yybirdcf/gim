package common

import (
	"container/list"
	"time"
)

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
	Time    int64
	From    int
	To      int
	Group   int
}

type MessageQueued struct {
	when    time.Time
	message *Message
}

//重用client结构
func MakeMessageRecycler() (get, put chan *Message) {
	get = make(chan *Message)
	put = make(chan *Message)

	go func() {
		queue := new(list.List)
		for {
			if queue.Len() == 0 {
				queue.PushFront(MessageQueued{
					when:    time.Now(),
					message: &Message{},
				})
			}

			ct := queue.Front()

			timeout := time.NewTimer(time.Minute)
			select {
			case b := <-put:
				timeout.Stop()

				b.Mid = 0
				b.Uid = 0
				b.Content = ""
				b.Type = 0
				b.Time = 0
				b.From = 0
				b.To = 0
				b.Group = 0

				mq := MessageQueued{
					when:    time.Now(),
					message: b,
				}
				queue.PushFront(mq)
			case get <- ct.Value.(MessageQueued).message:
				timeout.Stop()
				queue.Remove(ct)
			case <-timeout.C:
				ct := queue.Front()
				for ct != nil {
					n := ct.Next()
					if time.Since(ct.Value.(MessageQueued).when) > time.Minute {
						queue.Remove(ct)
						ct.Value = nil
					}
					ct = n
				}
			}
		}
	}()

	return
}
