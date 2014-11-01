package main

import (
	"glog"
	"runtime"
	"time"
)

func main() {
	glog.Infof("connect server start")
	defer glog.Flush()

	if err := InitConfig(); err != nil {
		glog.ErrorF("InitConfig() error(%v)", err)
		return
	}

	// Set max routine
	max := runtime.NumCPU()
	runtime.GOMAXPROCS(max)

	server := CreateServer()
	server.Start()
	defer server.Stop()
	// init process
	// sleep one second, let the listen start
	time.Sleep(time.Second)
	if err = common.InitProcess(Conf.User, Conf.Dir, Conf.PidFile); err != nil {
		glog.Errorf("common.InitProcess() error(%v)", err)
		return
	}
	// init signals, block wait signals
	signalCH := common.InitSignal()
	common.HandleSignal(signalCH)
	glog.Info("gim connect server stop")
}
