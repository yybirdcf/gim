package common

import (
	"net/rpc"
)

func InitRpcClient(host string) *rpc.Client {
	client, err := rpc.DialHTTP("tcp", host)
	if err != nil {
		panic(err.Error())
	}
	return client
}
