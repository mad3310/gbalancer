// Copyright 2014. All rights reserved.
// Use of this source code is governed by a GPLv3
// Author: Wenming Zhang <zhgwenming@gmail.com>

package main

import (
	"flag"
	"fmt"
	"github.com/zhgwenming/gbalancer/Godeps/_workspace/src/github.com/docker/spdystream"
	"github.com/zhgwenming/gbalancer/golog"
	"github.com/zhgwenming/gbalancer/utils"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
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
	pidFile     = flag.String("pidfile", "", "pid file")
	listenAddr  = flag.String("listen", ":6900", "port number")
	serviceAddr = flag.String("to", "/var/lib/mysql/mysql.sock", "service address")
	sigChan     = make(chan os.Signal, 1)
	wgroup      = &sync.WaitGroup{}
	logLevel     = flag.String("log-level", "", "log level [debug|info|warn|error], default error")
	logPath     =  flag.String("log-path", "/var/log/streamd/", "specify the log path")
)

func init() {
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
	
	//when the log file size greater than 1GB, kingshard will generate a new file
	if *logPath != "" {
		sysFilePath := path.Join(*logPath, sysLogName)
		sysFile, err := golog.NewRotatingFileHandler(sysFilePath, MaxLogSize, 1)
		if err != nil {
			fmt.Printf("new log file error:%v\n", err.Error())
			return
		}
		golog.GlobalSysLogger = golog.New(sysFile, golog.Lfile|golog.Ltime|golog.Llevel)
	}

	if *logLevel != "" {
		setLogLevel(*logLevel)
	} else {
		setLogLevel("error")
	}

	if *pidFile != "" {
		if err := utils.WritePid(*pidFile); err != nil {
			golog.Error("Main", "main", fmt.Sprintf("%s", err), 0)
			os.Exit(1)
		}
		defer func() {
			if err := os.Remove(*pidFile); err != nil {
				golog.Error("Main", "main", fmt.Sprintf("error to remove pidfile %s:", err), 0)
			}
		}()
	}

	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		golog.Error("Main", "main", fmt.Sprintf("Listen error: %s", err), 0)
		os.Exit(1)
	}
	
	var spdyConns = make([] *spdystream.Connection, 128)
	
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				golog.Error("Main", "main", fmt.Sprintf("Accept error: %s", err), 0)
			}
			spdyConn, err := spdystream.NewConnection(conn, true)
			if err != nil {
				conn.Close()
				golog.Error("Main", "main", fmt.Sprintf("New spdyConnection error, %s", err), 0)
			}
			spdyConns = append(spdyConns, spdyConn)
			go spdyConn.Serve(AgentStreamHandler)
		}
	}()
	
	golog.Info("Main", "main", fmt.Printf("starting clean up connections..."), 0)
	
	// waiting for exit signals
	for sig := range sigChan {
		golog.Info("Main", "main", fmt.Sprintf("captured %v, exiting..", sig), 0)
		
		for _, spdyConn := range spdyConns {
			if nil != spdyConn{
				spdyConn.Close()
			}
		}
		
		golog.GlobalSysLogger.Close()
		
		return
	}
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
