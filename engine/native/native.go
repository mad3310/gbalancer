// Copyright 2014. All rights reserved.
// Use of this source code is governed by a GPLv3
// Author: Wenming Zhang <zhgwenming@gmail.com>

package native

import (
	"flag"
	"github.com/zhgwenming/gbalancer/golog"
	"github.com/zhgwenming/gbalancer/config"
	"net"
	"sync"
)

var (
	tunnels    = flag.Uint("tunnels", 0, "number of tunnels per server")
	streamPort = flag.String("streamport", "6900", "port of the remote stream server")
	failover   = flag.Bool("failover", false, "whether to enable failover mode for scheduling")
	shuffle    = flag.Bool("shuffle", true, "whether to enable shuffle for server list")
)

func Serve(settings *config.Configuration, wgroup *sync.WaitGroup, done chan struct{}, status chan map[string]int) {
	job := make(chan *Request)

	// start the scheduler
	sch := NewScheduler(*failover, *tunnels)
	go sch.Schedule(job, status)

	listenAddrs, err := settings.GetListenAddrs()
	if err != nil {
		golog.Fatal("Native", "Serve", "" , 0, err)
	}

	for _, listenAddr := range listenAddrs {
		listener, err := listenAddr.Listen()

		// close the listener makes the unix socket file got removed
		wgroup.Add(1)
		go func() {
			<-done

			golog.Info("Native", "Serve", "starting clean up connection...." , 0)
			//close the backends connection for spdy
			for addr, _ := range sch.backends {
				sch.RemoveBackend(addr)
			}

			listener.Close()
		}()

		if err != nil {
			golog.Fatal("Native", "Serve", "", 0, err)
		}

		// tcp/unix listener
		go func(listen config.ListenAddr) {

			for {
				if conn, err := listener.Accept(); err == nil {
					//log.Println("main: got a connection")
					req := &Request{Conn: conn}
					job <- req
				} else {
					if neterr, ok := err.(net.Error); ok && neterr.Temporary() {
						golog.Info("Native", "Serve", "%s\n", 0, err)
					} else {
						// we should got a errClosing
						golog.Info("Native", "Serve", "stop listening for %s:%s\n", 0, listen.Net, listen.Addr)
						wgroup.Done()
						return
					}
				}
			}
		}(listenAddr)
	}
}
