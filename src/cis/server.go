package main

import (
	"common"
	"fmt"
	"net/http"
	"net/rpc"
)

type Args struct {
	Id     int
	Ids    []int
	Server string
}

type Cis struct {
	clientInfoMap *common.SafeMap
}

func NewCis() *Cis {
	cis := &Cis{
		clientInfoMap: common.NewSafeMap(),
	}

	return cis
}

func (self *Cis) GetClient(args *Args, reply *string) error {
	server := self.clientInfoMap.Get(args.Id)
	if server != nil {
		if s, ok := server.(string); ok {
			*reply = s
		}
	}

	return nil
}

func (self *Cis) GetClients(args *Args, reply *map[int]string) error {
	infos := make(map[int]string)
	for _, id := range args.Ids {
		server := self.clientInfoMap.Get(id)
		if server != nil {
			if s, ok := server.(string); ok {
				infos[id] = s
			}
		}
	}

	*reply = infos
	return nil
}

func (self *Cis) SetClient(args *Args, reply *bool) error {
	ok := self.clientInfoMap.Set(args.Id, args.Server)
	*reply = ok
	return nil
}

func (self *Cis) CheckClient(args *Args, reply *bool) error {
	ok := self.clientInfoMap.Check(args.Id)
	*reply = ok
	return nil
}

func (self *Cis) GetTotal(args *Args, reply *int) error {
	*reply = self.clientInfoMap.Len()
	return nil
}

func (self *Cis) DeleteClient(args *Args, reply *bool) error {
	*reply = true
	self.clientInfoMap.Delete(args.Id)
	return nil
}

func StartCis() {
	cis := NewCis()
	rpc.Register(cis)
	rpc.HandleHTTP()

	err := http.ListenAndServe(Conf.HttpBind, nil)
	if err != nil {
		fmt.Printf("cis rpc error: %s\n", err.Error())
	}
}
