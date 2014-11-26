package main

import (
	"common"
	"math/rand"
	"net/http"
	"time"
)

type UserArgs struct {
	id    int
	token string
}

type ClientArgs struct {
	guid    string
	connSrv string
	userId  int
}

//测试web服务器
func TestServer(w http.ResponseWriter, r *http.Request) {
	//限制get
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
	}

	//获取参数(无)
	res := map[string]interface{}{"ret": OK}
	retWrite(w, r, res, "", time.Now())
}

//分配一个新的客户端
func ClientDispatch(w http.ResponseWriter, r *http.Request) {
	//判断用户是否有效
	client, err := rpc.DialHTTP("tcp", Conf.MS)
	if err != nil {
		fmt.Printf("web connect ms, connect %s failed\n", Conf.MS)
		return
	}

	r.ParseForm()
	if len(r.Form["user_id"]) == 0 || len(r.Form["token"]) == 0 {
		res := map[string]interface{}{"ret": PARAM_ERR, "msg": "缺少用户ID或者token信息"}
		retWrite(w, r, res, "", time.Now())
		return
	}

	userId := r.Form["user_id"][0]
	token := r.Form["token"][0]

	userArgs := UserArgs{
		id:    userId,
		token: token,
	}

	var exist bool
	err = client.Call("MS.GetUser", userArgs, &exist)
	if err != nil {
		fmt.Printf("Web call MS, call MS.GetUser failed: %s\n", err.Error())
		return
	}

	if !exist {
		res := map[string]interface{}{"ret": -1, "msg": "用户不存在"}
		retWrite(w, r, res, "", time.Now())
		return
	}

	//生成guid
	rand.Seed(time.Now().UTC().UnixNano())
	guid, err := common.NewGuid(rand.Intn(9999999))
	//分配connect srv
	connSrv := "conn01"
	//新增一个客户端信息
	clientArgs := ClientArgs{
		guid:    guid,
		connSrv: connSrv,
		userId:  userId,
	}
	var success bool
	err = client.Call("MS.NewClientInformation", clientArgs, &success)
	if err != nil {
		fmt.Printf("Web call MS, call MS.NewClientInformation failed: %s\n", err.Error())
		return
	}

	if !success {
		res := map[string]interface{}{"ret": -1, "msg": "创建客户端信息失败"}
		retWrite(w, r, res, "", time.Now())
		return
	}

	res := map[string]interface{}{"ret": OK, "msg": "", "connsrv": connSrv, "guid": guid}
	retWrite(w, r, res, "", time.Now())
	return
}
