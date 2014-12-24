package main

import (
	"encoding/json"
	"gim/common"
	"github.com/garyburd/redigo/redis"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const (
	RCP_TCP_HOST_PREFIX = "rcp&tcp#"
)

func retJson(w http.ResponseWriter, r *http.Request, retCode int, retMsg string, retData interface{}) {
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
	r.ParseForm()
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

	data, err := json.Marshal(m)
	if err != nil {
		retJson(w, r, -1, "发送消息失败", nil)
		return
	}
	redClient.Do("LPUSH", common.MSG_QUEUE_0, string(data))
	//输出结果到前台
	retJson(w, r, 0, "发送成功", nil)
	return
}

//http://127.0.0.1:8180/sub/send?subId=1001&msg=hello&from=2&to=10002
func SendSubMsg(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
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

	data, err := json.Marshal(m)
	if err != nil {
		retJson(w, r, -1, "发送消息失败", nil)
		return
	}
	redClient.Do("LPUSH", common.MSG_QUEUE_0, string(data))
	//输出结果到前台
	retJson(w, r, 0, "发送成功", nil)
	return
}

func GetAvalConnSrv(w http.ResponseWriter, r *http.Request) {
	//获取可用的conn srv子节点列表
	children := common.ZkGetChildren(zkConn, Conf.ZkRoot)
	if children != nil {
		//随机一个机器
		rand.Seed(time.Now().UnixNano())
		idx := rand.Intn(len(children))
		host, _ := redis.String(redClient.Do("GET", RCP_TCP_HOST_PREFIX+children[idx]))
		retJson(w, r, 0, host, nil)
		return
	}

	retJson(w, r, -1, "没有可用的连接服务器", nil)
	return
}

type RArgs struct {
	Who   int
	MaxId int64
	Limit int
}

//获取离线消息
func GetOfflineMsgs(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	uid := r.Form.Get("uid")
	lastMaxMsgId := r.Form.Get("maxid")
	limit := r.Form.Get("limit")

	who, _ := strconv.Atoi(uid)
	maxId, _ := strconv.Atoi(lastMaxMsgId)
	size, _ := strconv.Atoi(limit)
	if size > 20 {
		size = 20
	}
	//调用ms srv rpc将消息获取离线消息
	args := RArgs{
		Who:   who,
		MaxId: int64(maxId),
		Limit: size,
	}
	var msgs []common.Message
	err := msClient.Call("MS.ReadMessages", args, &msgs)
	if err != nil {
		panic(err.Error())
		retJson(w, r, -1, "获取离线消息失败", nil)
		return
	}

	retJson(w, r, 0, "OK", msgs)
	return
}
