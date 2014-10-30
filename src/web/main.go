package main

import (
	"common"
	"glog"
	"runtime"
	"time"
)

func main() {
	var err error

	glog.Infof("gim web start....")
	defer glog.Flush()
	if err = InitConfig(); err != nil {
		glog.Errorf("InitConfig() error(%v)", err)
		return
	}
	// Set max routine
	max := runtime.NumCPU()
	runtime.GOMAXPROCS(max)

	// start http listen.
	StartHTTP()
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
	glog.Info("gim web stop")
}
