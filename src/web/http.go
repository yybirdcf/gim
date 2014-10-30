package main

import (
	"encoding/json"
	"fmt"
	"glog"
	"net"
	"net/http"
	"time"
)

const (
	httpReadTimeout = 30 //sec
)

// StartHTTP start listen http.
func StartHTTP() {
	// external
	httpServeMux := http.NewServeMux()
	// 1.0
	httpServeMux.HandleFunc("/server/test", TestServer)

	for _, bind := range Conf.HttpBind {
		glog.Infof("start http listen addr:\"%s\"", bind)
		go httpListen(httpServeMux, bind)
	}
}

func httpListen(mux *http.ServeMux, bind string) {
	server := &http.Server{Handler: mux, ReadTimeout: httpReadTimeout * time.Second}
	l, err := net.Listen("tcp", bind)
	if err != nil {
		glog.Errorf("net.Listen(\"tcp\", \"%s\") error(%v)", bind, err)
		panic(err)
	}
	if err := server.Serve(l); err != nil {
		glog.Errorf("server.Serve() error(%v)", err)
		panic(err)
	}
}

//json数据输出到客户端
func retWrite(w http.ResponseWriter, r *http.Request, res map[string]interface{}, callback string, start time.Time) {
	data, err := json.Marshal(res) //格式化json数据
	if err != nil {
		glog.Errorf("json.Marshal(\"%v\") error(%v)", res, err)
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
		glog.Errorf("w.Write(\"%s\") error(%v)", dataStr, err)
	} else {
		glog.V(1).Infof("w.Write(\"%s\") write %d bytes", dataStr, n)
	}
	glog.Infof("req: \"%s\", res:\"%s\", ip:\"%s\", time:\"%fs\"", r.URL.String(), dataStr, r.RemoteAddr, time.Now().Sub(start).Seconds())
}
