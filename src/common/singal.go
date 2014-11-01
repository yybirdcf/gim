package common

import (
	l4g "log4go"
	"os"
	"os/signal"
	"syscall"
)

// InitSignal register signals handler.
func InitSignal() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	return c
}

// HandleSignal fetch signal from chan then do exit or reload.
func HandleSignal(c chan os.Signal) {
	// Block until a signal is received.
	for {
		s := <-c
		l4g.Trace("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			return
		case syscall.SIGHUP:
			// TODO reload
			//return
		default:
			return
		}
	}
}
