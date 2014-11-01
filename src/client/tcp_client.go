package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	var ip = "127.0.0.1:8280"
	fmt.Printf("start to connect %s\n", ip)

	conn, err := net.Dial("tcp", ip)
	if err != nil {
		fmt.Printf("connect to %s failed\n", ip)
	}

	defer conn.Close()

	stdin := bufio.NewReader(os.Stdin)
	stdout := bufio.NewWriter(os.Stdout)
	connin := bufio.NewReader(conn)
	connout := bufio.NewWriter(conn)

	//gorouting负责接收服务器消息
	go func() {
		for {
			if line, _, err := connin.ReadLine(); err == nil {
				stdout.WriteString(string(line))
			}
		}
	}()

	//主进程负责接收用户输入
	for {
		if line, _, err := stdin.ReadLine(); err == nil {
			connout.WriteString(string(line) + "\n")
		}
	}
}
