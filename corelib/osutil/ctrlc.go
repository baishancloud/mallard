package osutil

import (
	"os"
	"os/signal"
	"syscall"
)

// WaitSignals wait signals capture
func WaitSignals(signals ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	<-c
}

// Wait is short name for WaitInterrupt
func Wait() {
	WaitSignals(os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
}
