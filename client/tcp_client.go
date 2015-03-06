package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gim/common"
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
	// var uid int
	var ip = "127.0.0.1:8280"
	fmt.Printf("start to connect %s\n", ip)

	conn, err := net.Dial("tcp", ip)
	if err != nil {
		fmt.Printf("connect to %s failed\n", ip)
	}
	fmt.Printf("connect success\n")
	defer conn.Close()

	// stdin := bufio.NewReader(os.Stdin)
	stdout := bufio.NewWriter(os.Stdout)
	connin := bufio.NewReader(conn)
	connout := bufio.NewWriter(conn)

	//发起认证
	cc := ClientCmd{
		Cmd:    "AUTH",
		Params: "gim#test#key&gim#test#key#secret&Jack&123456",
	}
	str, _ := json.Marshal(cc)
	connout.WriteString(string(str) + "\n")
	connout.Flush()

	if line, _, err := connin.ReadLine(); err == nil {
		//如果是服务器ping，则需要回复，保持心跳
		var resp Resp
		err = json.Unmarshal(line, &resp)
		if err != nil {
			panic(err.Error())
		}

		if resp.RetCode != 0 {
			fmt.Printf("%s\n", "认证失败")
			return
		}

		if resp.RetCode == 0 && resp.RetType == "AUTH" {
			// uid = int(resp.RetData.(float64))
		}
	} else {
		return
	}

	//gorouting负责接收服务器消息
	go func() {
		var resp Resp
		for {
			if line, _, err := connin.ReadLine(); err == nil {
				//如果是服务器ping，则需要回复，保持心跳
				err = json.Unmarshal(line, &resp)
				if err != nil {
					fmt.Printf("%v\n", line)
					fmt.Printf("%s\n", err.Error())
				}

				fmt.Printf("%v\n", resp)
				if resp.RetType == "PING" {
					clc := ClientCmd{
						Cmd:    "PONG",
						Params: "",
					}
					str, _ := json.Marshal(clc)
					connout.WriteString(string(str) + "\n")
					connout.Flush()
				} else if resp.RetType == "MSG" {
					msg := resp.RetData.(map[string]interface{})
					mid := int64(msg["Mid"].(float64))
					fmt.Printf("%v\n", mid)
					params := fmt.Sprintf("%d", mid)
					clc := ClientCmd{
						Cmd:    "MSRECEIVEACK",
						Params: params,
					}

					str, _ := json.Marshal(clc)
					connout.WriteString(string(str) + "\n")
					connout.Flush()
				} else if resp.RetType == "MSSENDGACK" {
					fmt.Printf("%v\n", "msg send success")
				}

				// stdout.WriteString(string(line))
				stdout.Flush()
			} else {
				fmt.Printf("read msg failed\n")
				break
			}
		}
	}()

	// 负责接收用户输入
	go func() {
		for {
			clc := ClientCmd{
				Cmd:    "MSG",
				Params: "",
			}
			cm := ClientMsg{
				UniqueId: time.Now().UnixNano(),
				Content:  "say hello world from Jack",
				To:       1001,
				Type:     4,
			}
			str, _ := json.Marshal(cm)
			clc.Params = string(str)
			str2, _ := json.Marshal(clc)

			connout.WriteString(string(str2) + "\n")
			connout.Flush()

			time.Sleep(4 * time.Second)
		}
	}()

	signalCH := common.InitSignal()
	common.HandleSignal(signalCH)
	fmt.Printf("client stop\n")
}
