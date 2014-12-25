package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
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
	for i := 15000; i < 20000; i++ {
		if i%500 == 0 {
			time.Sleep(time.Second * 5)
		}
		func() {
			fmt.Printf("start to connect %s\n", ip)

			conn, err := net.Dial("tcp", ip)
			if err != nil {
				panic(err.Error())
				fmt.Printf("connect to %s failed\n", ip)
			}

			// defer conn.Close()

			// stdin := bufio.NewReader(os.Stdin)
			stdout := bufio.NewWriter(os.Stdout)
			connin := bufio.NewReader(conn)
			connout := bufio.NewWriter(conn)

			//gorouting负责接收服务器消息
			go func() {
				defer func() {
					if err := recover(); err != nil {
						conn.Close()
					}
				}()

				var resp Resp
				for {
					if line, err := connin.ReadBytes('\n'); err == nil {
						//如果是服务器ping，则需要回复，保持心跳
						line = bytes.TrimRight(line, "\r\n")
						err = json.Unmarshal(line, &resp)

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
					} else if err == io.EOF {
						fmt.Printf("read msg failed: %v\n", err.Error())
						break
					} else {
						panic(err.Error())
					}
				}
			}()

			//发起认证
			cc := ClientCmd{
				Cmd:    "AUTH",
				Params: strconv.Itoa(i + 1),
			}
			str, _ := json.Marshal(cc)

			connout.WriteString(string(str) + "\n")
			connout.Flush()
		}()
	}

	for {
		time.Sleep(4 * time.Second)
	}
}
