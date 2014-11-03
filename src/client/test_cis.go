package main

import (
	"fmt"
	"net/rpc"
)

type Args struct {
	id     int
	ids    []int
	server string
}

func main() {
	server := "127.0.0.1:8580"
	client, err := rpc.DialHTTP("tcp", server)
	if err != nil {
		fmt.Printf("cis test, connect %s failed\n", server)
		return
	}

	args := Args{
		id:     12345,
		server: "127.0.0.1:8280",
		ids:    []int{12345},
	}

	var reply_b bool
	err = client.Call("Cis.SetClient", args, &reply_b)
	if err != nil {
		fmt.Printf("cis test, call Cis.SetClient failed\n")
		return
	}

	fmt.Printf("cis test, Cis.SetClient reply: %b\n", reply_b)

	err = client.Call("Cis.CheckClient", args, &reply_b)
	if err != nil {
		fmt.Printf("cis test, call Cis.CheckClient failed\n")
		return
	}

	fmt.Printf("cis test, Cis.CheckClient reply: %b\n", reply_b)

	var reply_str string
	err = client.Call("Cis.GetClient", args, &reply_str)
	if err != nil {
		fmt.Printf("cis test, call Cis.GetClient failed\n")
		return
	}

	fmt.Printf("cis test, Cis.GetClient reply: %s\n", reply_str)

	var reply_m map[int]string
	err = client.Call("Cis.GetClients", args, &reply_m)
	if err != nil {
		fmt.Printf("cis test, call Cis.GetClients failed\n")
		return
	}

	fmt.Printf("cis test, Cis.GetClients reply: %v\n", reply_m)

	var reply_i int
	err = client.Call("Cis.GetTotal", args, &reply_i)
	if err != nil {
		fmt.Printf("cis test, call Cis.GetTotal failed\n")
		return
	}

	fmt.Printf("cis test, Cis.GetTotal reply: %v\n", reply_i)

	err = client.Call("Cis.DeleteClient", args, &reply_b)
	if err != nil {
		fmt.Printf("cis test, call Cis.DeleteClient failed\n")
		return
	}

	fmt.Printf("cis test, Cis.DeleteClient reply: %v\n", reply_b)
}