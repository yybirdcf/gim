package main

import (
	"gim/common"
	"net/http"
	"strconv"
	"time"
)

func retJson(w http.ResponseWriter, r *http.Request, retCode int, retMsg string, retData map[interface{}]interface{}) {
	res := map[string]interface{}{
		"retCode": retCode,
		"retMsg":  retMsg,
		"retData": retData,
	}
	retWrite(w, r, res, "", time.Now())
}

//http://127.0.0.1:8180/public/send?msg=hello&from=1&to=1000
//http://127.0.0.1:8180/public/send?msg=hello&from=1&to=1001
func SendPublicMsg(w http.ResponseWriter, r *http.Request) {
	msg := r.Form.Get("msg")
	from := r.Form.Get("from")
	to := r.Form.Get("to")
	if msg == "" || from == "" || to == "" {
		retJson(w, r, -1, "缺少参数", nil)
		return
	}

	From, _ := strconv.Atoi(from)
	Uid, _ := strconv.Atoi(to)
	//构造一个公开消息结构
	m := common.Message{
		Mid:     0,
		Uid:     Uid,
		Content: msg,
		Type:    common.MESSAGE_TYPE_PUBLIC,
		Time:    int(time.Now().Unix()),
		From:    From,
		To:      Uid,
		Group:   0,
	}

	//调用send srv rpc将消息发过去
	var reply bool
	err := sendSrvClient.Call("SendSrv.SendMsg", m, &reply)
	if err != nil || !reply {
		retJson(w, r, -1, "发送消息失败", nil)
		return
	}
	//输出结果到前台
	retJson(w, r, 0, "发送成功", nil)
	return
}

//http://127.0.0.1:8180/sub/send?subId=1001&msg=hello&from=2&to=10002
func SendSubMsg(w http.ResponseWriter, r *http.Request) {
	subId := r.Form.Get("subId")
	msg := r.Form.Get("msg")
	from := r.Form.Get("from")
	to := r.Form.Get("to")
	if msg == "" || subId == "" || from == "" || to == "" {
		retJson(w, r, -1, "缺少参数", nil)
		return
	}

	Group, _ := strconv.Atoi(subId)
	From, _ := strconv.Atoi(from)
	Uid, _ := strconv.Atoi(to)
	//构造一个订阅消息结构
	m := common.Message{
		Mid:     0,
		Uid:     Uid,
		Content: msg,
		Type:    common.MESSAGE_TYPE_SUB,
		Time:    int(time.Now().Unix()),
		From:    From,
		To:      Uid,
		Group:   Group,
	}

	//调用send srv rpc将消息发过去
	var reply bool
	err := sendSrvClient.Call("SendSrv.SendMsg", m, &reply)
	if err != nil || !reply {
		retJson(w, r, -1, "发送消息失败", nil)
		return
	}
	//输出结果到前台
	retJson(w, r, 0, "发送成功", nil)
	return
}
