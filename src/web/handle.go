package main

import (
	"glog"
	"net/http"
	"time"
)

//测试web服务器
func TestServer(w http.ResponseWriter, r *http.Request) {
	//限制get
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
	}

	//获取参数(无)
	res := map[string]interface{}{"ret": OK}
	retWrite(w, r, res, callback, start)
}
