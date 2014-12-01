package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

//客户端发送命令
type ClientCmd struct {
	cmd    string //命令类型
	params string //参数&作为分隔符
}

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
				stdout.Flush()
			}
		}
	}()

	//主进程负责接收用户输入
	for {
		if line, _, err := stdin.ReadLine(); err == nil {
			connout.WriteString(string(line) + "\n")
			connout.Flush()
		}
	}

	//发起认证
	cc := ClientCmd{
		cmd:    "AUTH",
		params: "1000",
	}
	str, _ := json.Marshal(cc)
	connout.WriteString(string(str) + "\n")
	connout.Flush()
}
