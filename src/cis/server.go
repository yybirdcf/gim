package main

import (
	"common"
	"fmt"
	"net/http"
	"net/rpc"
	"time"
	"unsafe"
)

type Args struct {
	id     int
	ids    []int
	server string
}

type Cis struct {
	clientInfoMap *common.SafeMap
}

type Information struct {
	start  int64
	server string
}

func NewCis() *Cis {
	cis := &Cis{
		clientInfoMap: common.NewSafeMap(),
	}

	return cis
}

func (self *Cis) GetClient(args *Args, reply *string) error {
	information := self.clientInfoMap.Get(args.id)
	if information != nil {
		info := (* Information)unsafe.Pointer(information)
		*reply = info.server
	}

	*reply = nil
	return nil
}

func (self *Cis) GetClients(args *Args, reply *map[int]string) error {
	infos := make(map[int]string)
	for id := range args.ids {
		information := self.clientInfoMap.Get(id)
		if information != nil {
			info := (* Information)unsafe.Pointer(information)
			infos[id] = info.server
		}
	}

	reply = &infos
	return nil
}

func (self *Cis) SetClient(args *Args, reply *bool) error {
	information := &Information{
		start:  time.Now().Unix(),
		server: args.server,
	}
	ok := self.clientInfoMap.Set(args.id, &information)
	*reply = ok
	return nil
}

func (self *Cis) CheckClient(args *Args, reply *bool) error {
	ok := self.clientInfoMap.Check(args.id)
	*reply = ok
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
