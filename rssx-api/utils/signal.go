package utils

import (
	"os"
	"os/signal"
	"syscall"
)

var signals = make(chan os.Signal, 1)

func init() {
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGTERM)
}

func WaitSignals() {
	for s := range signals {
		if s == os.Interrupt || s == os.Kill || s == syscall.SIGTERM {
			break
		}
	}
	signal.Stop(signals)
}
