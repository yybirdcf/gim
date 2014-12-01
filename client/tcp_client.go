package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

//客户端发送命令
type ClientCmd struct {
	Cmd    string //命令类型
	Params string //参数&作为分隔符
}

type ClientMsg struct {
	UniqueId int64  //nano sec可以唯一区分消息
	Content  string //消息内容
	To       int    //发送给的目标
	Type     int    //消息类型,个人消息，还是群组消息
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

	//发起认证
	cc := ClientCmd{
		Cmd:    "AUTH",
		Params: "1000",
	}
	str, _ := json.Marshal(cc)
	fmt.Printf("%s\n", string(str))
	connout.WriteString(string(str) + "\n")
	connout.Flush()

	//负责接收用户输入

	for {
		cm := ClientMsg{
			UniqueId: time.Nanosecond(),
			Content:  "hello world",
			To:       10001,
			Type:     4,
		}
		str, _ := json.Marshal(cm)

		connout.WriteString(string(str) + "\n")
		connout.Flush()

		time.Sleep(1 * time.Second)
	}
}
