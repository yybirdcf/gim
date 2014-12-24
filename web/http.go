package main

import (
	"encoding/json"
	"fmt"
	"gim/common"
	"github.com/garyburd/redigo/redis"
	"github.com/samuel/go-zookeeper/zk"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

const (
	httpReadTimeout = 30 //sec
)

var (
	msClient  *rpc.Client
	zkConn    *zk.Conn
	redClient redis.Conn
)

// StartHTTP start listen http.
func StartHTTP() {
	// external
	httpServeMux := http.NewServeMux()
	// 1.0
	httpServeMux.HandleFunc("/public/send", SendPublicMsg)
	httpServeMux.HandleFunc("/sub/send", SendSubMsg)
	httpServeMux.HandleFunc("/get/connectserver", GetAvalConnSrv)
	httpServeMux.HandleFunc("/get/offlinemsg", GetOfflineMsgs)

	for _, bind := range Conf.HttpBind {
		fmt.Printf("start http listen addr:\"%s\"", bind)
		go httpListen(httpServeMux, bind)
	}

	//开启send srv rpc
	client, err := rpc.DialHTTP("tcp", Conf.Ms)
	if err != nil {
		panic(err.Error())
		fmt.Printf("web start rpc failed, connect %s failed\n", Conf.Ms)
		return
	}
	msClient = client

	conn, err := redis.Dial("tcp", Conf.Redis)
	if err != nil {
		panic(err.Error())
	}
	redClient = conn

	zkConn = common.ZkConnect(Conf.ZooKeeper)
}

func httpListen(mux *http.ServeMux, bind string) {
	server := &http.Server{Handler: mux, ReadTimeout: httpReadTimeout * time.Second}
	l, err := net.Listen("tcp", bind)
	if err != nil {
		fmt.Printf("net.Listen(\"tcp\", \"%s\") error(%v)", bind, err)
		panic(err.Error())
	}
	if err := server.Serve(l); err != nil {
		fmt.Printf("server.Serve() error(%v)", err)
		panic(err.Error())
	}
}

//json数据输出到客户端
func retWrite(w http.ResponseWriter, r *http.Request, res map[string]interface{}, callback string, start time.Time) {
	data, err := json.Marshal(res) //格式化json数据
	if err != nil {
		fmt.Printf("json.Marshal(\"%v\") error(%v)", res, err)
		return
	}
	dataStr := ""
	if callback == "" {
		// Normal json
		dataStr = string(data)
	} else {
		// Jsonp
		dataStr = fmt.Sprintf("%s(%s)", callback, string(data))
	}
	if n, err := w.Write([]byte(dataStr)); err != nil {
		fmt.Printf("w.Write(\"%s\") error(%v)", dataStr, err)
	} else {
		fmt.Printf("w.Write(\"%s\") write %d bytes", dataStr, n)
	}
	fmt.Printf("req: \"%s\", res:\"%s\", ip:\"%s\", time:\"%fs\"", r.URL.String(), dataStr, r.RemoteAddr, time.Now().Sub(start).Seconds())
}
