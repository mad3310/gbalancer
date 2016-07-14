// Copyright 2014. All rights reserved.
// Use of this source code is governed by a GPLv3
// Author: Wenming Zhang <zhgwenming@gmail.com>

package main

import (
	"flag"
	"fmt"
	"github.com/zhgwenming/gbalancer/config"
	"github.com/zhgwenming/gbalancer/daemon"
	"github.com/zhgwenming/gbalancer/engine"
	"github.com/zhgwenming/gbalancer/golog"
	"os"
	"runtime"
	"sync"
	"path"
	"strings"
)

const (
	VERSION = "0.6.6"
	sqlLogName = "sql.log"
	sysLogName = "sys.log"
	MaxLogSize = 1024 * 1024 * 1024
)

var (
	wgroup       = &sync.WaitGroup{}
	configFile   = flag.String("config", "gbalancer.json", "Configuration file")
	daemonMode   = flag.Bool("daemon", false, "daemon mode")
	printVersion = flag.Bool("version", false, "print gbalancer version")
	logLevel     = flag.String("log-level", "", "log level [debug|info|warn|error], default error")
	logPath     =  flag.String("log-path", "/var/log/gbalancer/", "specify the log path")
)

func PrintVersion() {
	fmt.Printf("gbalancer version: %s\n", VERSION)
	os.Exit(0)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	if *printVersion {
		PrintVersion()
	}

	if *daemonMode {
		os.Chdir("/")
	}
	
	//when the log file size greater than 1GB, kingshard will generate a new file
	if *logPath != "" {
		sysFilePath := path.Join(*logPath, sysLogName)
		sysFile, err := golog.NewRotatingFileHandler(sysFilePath, MaxLogSize, 1)
		if err != nil {
			fmt.Printf("new log file error:%v\n", err.Error())
			return
		}
		golog.GlobalSysLogger = golog.New(sysFile, golog.Lfile|golog.Ltime|golog.Llevel)

		sqlFilePath := path.Join(*logPath, sqlLogName)
		sqlFile, err := golog.NewRotatingFileHandler(sqlFilePath, MaxLogSize, 1)
		if err != nil {
			fmt.Printf("new log file error:%v\n", err.Error())
			return
		}
		golog.GlobalSqlLogger = golog.New(sqlFile, golog.Lfile|golog.Ltime|golog.Llevel)
	}

	if *logLevel != "" {
		setLogLevel(*logLevel)
	} else {
		setLogLevel("error")
	}

	// Load configurations
	settings, err := config.LoadConfig(*configFile)
	if err != nil {
		golog.Fatal("Main", "main", fmt.Sprintf("%s", err), 0)
	}
	golog.Info("Main", "main", settings.ListenInfo() , 0)

	daemon.CreatePidfile()

	// create the service goroutine
	done := engine.Serve(settings, wgroup)

	// wait the exit signal then do cleanup
	daemon.WaitSignal(func() {
		close(done)
		golog.GlobalSysLogger.Close()
		golog.GlobalSqlLogger.Close()
		wgroup.Wait()
	})
}

func setLogLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		golog.GlobalSysLogger.SetLevel(golog.LevelDebug)
	case "info":
		golog.GlobalSysLogger.SetLevel(golog.LevelInfo)
	case "warn":
		golog.GlobalSysLogger.SetLevel(golog.LevelWarn)
	case "error":
		golog.GlobalSysLogger.SetLevel(golog.LevelError)
	default:
		golog.GlobalSysLogger.SetLevel(golog.LevelError)
	}
}
