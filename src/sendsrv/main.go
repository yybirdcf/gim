package sendsrv

import (
	"common"
	"fmt"
	"runtime"
	"time"
)

func main() {

	fmt.Printf("send server start\n")

	if err := InitConfig(); err != nil {
		fmt.Printf("InitConfig() error(%v)\n", err)
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
	if err := common.InitProcess(Conf.User, Conf.Dir, Conf.PidFile); err != nil {
		fmt.Printf("common.InitProcess() error(%v)\n", err)
		return
	}
	// init signals, block wait signals
	signalCH := common.InitSignal()
	common.HandleSignal(signalCH)
	fmt.Printf("gim send server stop\n")
}
