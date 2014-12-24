package main

import (
	"flag"
	"fmt"
	"gim/common"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

func main() {
	var cpuprofile = flag.String("cpuprofile", "", "--cpuprofile=<.prof file path>")
	var memprofile = flag.String("memprofile", "", "--memprofile=<.prof file path>")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			panic(err.Error())
		}
		pprof.WriteHeapProfile(f)
		defer f.Close()
	}

	fmt.Printf("connect server start\n")

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

	StartRpc()
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
	fmt.Printf("gim connect server stop\n")
}
