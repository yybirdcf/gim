package main

import (
	"gim/common"
	"math/rand"
	"net/http"
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
	if msg == nil || from == nil || to == nil {
		retJson(w, r, -1, "缺少参数", nil)
		return
	}

	//构造一个公开消息结构
	m := common.Message{
		Mid:     0,
		Uid:     to,
		Content: msg,
		Type:    common.MESSAGE_TYPE_PUBLIC,
		Time:    time.Now().Unix(),
		From:    from,
		To:      to,
		Group:   0,
	}

	//调用send srv rpc将消息发过去
	reply bool
	err := sendSrvClient.Call("SendSrv.SendMsg", m, &reply)
	if err != nil || !reply{
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
	if msg == nil || subId == nil || from == nil || to == nil{
		retJson(w, r, -1, "缺少参数", nil)
		return
	}

	//构造一个订阅消息结构
	m := common.Message{
		Mid:     0,
		Uid:     to,
		Content: msg,
		Type:    common.MESSAGE_TYPE_SUB,
		Time:    time.Now().Unix(),
		From:    from,
		To:      to,
		Group:   subId,
	}

	//调用send srv rpc将消息发过去
	reply bool
	err := sendSrvClient.Call("SendSrv.SendMsg", m, &reply)
	if err != nil || !reply{
		retJson(w, r, -1, "发送消息失败", nil)
		return
	}
	//输出结果到前台
	retJson(w, r, 0, "发送成功", nil)
	return
}
