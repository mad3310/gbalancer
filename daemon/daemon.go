// Copyright 2014. All rights reserved.
// Use of this source code is governed by a GPLv3
// Author: Wenming Zhang <zhgwenming@gmail.com>

package daemon

import (
	"flag"
	"fmt"
	"github.com/zhgwenming/gbalancer/golog"
	"github.com/zhgwenming/gbalancer/utils"
	"os"
	"os/signal"
	"syscall"
)

var (
	pidFile = flag.String("pidfile", "", "pid file")
	sigChan = make(chan os.Signal, 1)
)

func init() {
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
}

func CreatePidfile() {
	if *pidFile != "" {
		if err := utils.WritePid(*pidFile); err != nil {
			fmt.Printf("error: %s\n", err)
			golog.Fatal("Daemon", "CreatePidfile", "error:" , 0, err)
		}
	}
}

func RemovePidfile() {
	if *pidFile != "" {
		if err := os.Remove(*pidFile); err != nil {
			golog.Info("Daemon", "RemovePidfile", "error to remove pidfile %s:" , 0, err)
		}
	}
}

func WaitSignal(cleanup func()) {
	// waiting for exit signals
	for sig := range sigChan {
		golog.Info("Daemon", "WaitSignal", "captured %v, exiting.." , 0, sig)
		// exit if we get any signal
		// Todo - catch signal other than SIGTERM/SIGINT
		break
	}

	cleanup()
	RemovePidfile()
	return
}
