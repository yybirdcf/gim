package main

import (
	"common"
	"fmt"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

type Args struct {
	id     int
	ids    []int
	server string
}

type Cis struct {
	clientInfoMap *common.SafeMap
}

type Infomation struct {
	start  int
	server string
}

func NewCis() *Cis {
	cis := &Cis{
		clientInfoMap: common.NewSafeMap(),
		total:         0,
		lock:          new(sync.RWMutex),
	}

	return cis
}

func (self *Cis) GetClient(args *Args, reply *string) error {
	infomation := self.clientInfoMap.Get(args.id)
	if information != nil {
		*replay = information.server
	}

	*replay = nil
	return nil
}

func (self *Cis) GetClients(args *Args, reply *map[int]string) error {
	infos := make(map[int]string)
	for id := range args.ids {
		infomation := self.clientInfoMap.Get(id)
		if information != nil {
			infos[id] = information.server
		}
	}

	replay = &infos
	return nil
}

func (self *Cis) SetClient(args *Args, reply *bool) error {
	information := &Infomation{
		start:  time.Now().Unix(),
		server: args.server,
	}
	ok := self.clientInfoMap.Set(args.id, &information)
	*replay = ok
	return nil
}

func (self *Cis) CheckClient(args *Args, reply *bool) error {
	ok := self.clientInfoMap.Check(args.id)
	*replay = ok
	return nil
}

func (self *Cis) GetTotal(args *Args, reply *int) error {
	*reply = self.clientInfoMap.Len()
	return nil
}

func (self *Cis) DeleteClient(args *Args, reply *bool) error {
	*reply = true
	self.clientInfoMap.Delete(args.id)
	return nil
}

func StartCis() {
	cis := new(Cis)
	rpc.Register(cis)
	rpc.HandleHTTP()

	err := http.ListenAndServe(Conf.HttpBind, nil)
	if err != nil {
		fmt.Printf("cis rpc error: %s\n", err.Error())
	}
}
