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

type Resp struct {
	RetCode int
	RetType string
	RetMsg  string
	RetData interface{}
}

func main() {
	var ip = "115.29.241.118:8280"
	fmt.Printf("start to connect %s\n", ip)

	conn, err := net.Dial("tcp", ip)
	if err != nil {
		fmt.Printf("connect to %s failed\n", ip)
	}

	defer conn.Close()

	// stdin := bufio.NewReader(os.Stdin)
	stdout := bufio.NewWriter(os.Stdout)
	connin := bufio.NewReader(conn)
	connout := bufio.NewWriter(conn)

	//gorouting负责接收服务器消息
	go func() {
		var resp Resp
		for {
			if line, _, err := connin.ReadLine(); err == nil {
				//如果是服务器ping，则需要回复，保持心跳
				err = json.Unmarshal(line, &resp)
				if err != nil {
					panic(err.Error())
				}

				if resp.RetType == "PING" {
					clc := ClientCmd{
						Cmd:    "PONG",
						Params: "",
					}
					str, _ := json.Marshal(clc)
					connout.WriteString(string(str) + "\n")
					connout.Flush()
				}

				stdout.WriteString(string(line))
				stdout.Flush()
			} else {
				fmt.Printf("read msg failed\n")
				break
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

	// 负责接收用户输入

	for {
		clc := ClientCmd{
			Cmd:    "MSG",
			Params: "",
		}
		cm := ClientMsg{
			UniqueId: time.Now().UnixNano(),
			Content:  "group say hello world from 1000",
			To:       1,
			Type:     3,
		}
		str, _ := json.Marshal(cm)
		clc.Params = string(str)
		str2, _ := json.Marshal(clc)

		connout.WriteString(string(str2) + "\n")
		connout.Flush()

		time.Sleep(4 * time.Second)
	}
}
