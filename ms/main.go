package main

import (
	"fmt"
	"gim/common"
	"runtime"
	"time"
)

func main() {
	var err error

	fmt.Printf("gim ms start....\n")

	if err = InitConfig(); err != nil {
		fmt.Printf("InitConfig() error(%v)", err)
		return
	}
	// Set max routine
	max := runtime.NumCPU()
	runtime.GOMAXPROCS(max)

	// start http listen.
	StartMs()
	// init process
	// sleep one second, let the listen start
	time.Sleep(time.Second)
	if err = common.InitProcess(Conf.User, Conf.Dir, Conf.PidFile); err != nil {
		fmt.Printf("common.InitProcess() error(%v)", err)
		return
	}
	// init signals, block wait signals
	signalCH := common.InitSignal()
	common.HandleSignal(signalCH)
	fmt.Printf("gim ms stop\n")
}
