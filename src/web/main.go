package main

import (
	"common"
	l4g "log4go"
	"runtime"
	"time"
)

func main() {
	var err error

	l4g.Trace("gim web start....")

	if err = InitConfig(); err != nil {
		l4g.Error("InitConfig() error(%v)", err)
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
		l4g.Error("common.InitProcess() error(%v)", err)
		return
	}
	// init signals, block wait signals
	signalCH := common.InitSignal()
	common.HandleSignal(signalCH)
	l4g.Trace("gim web stop")
}
