// Copyright 2014. All rights reserved.
// Use of this source code is governed by a GPLv3
// Author: Wenming Zhang <zhgwenming@gmail.com>

package native

import (
	"github.com/zhgwenming/gbalancer/Godeps/_workspace/src/github.com/docker/spdystream"
	"net"
	"strings"
	"time"
	"github.com/zhgwenming/gbalancer/golog"
	"fmt"
)

const (
	STREAMPORT = "6900"
)

type connTunnel struct {
	conn      *spdystream.Connection
	tcpAddr   *net.TCPAddr
	switching bool
}

type spdySession struct {
	backend   *Backend
	spdy      *connTunnel
	connindex uint
}

func NewSpdySession(backend *Backend, index uint) *spdySession {
	return &spdySession{backend: backend, connindex: index}
}

func NewConnTunnel(conn net.Conn) (*connTunnel,error) {
	var spdyconn *connTunnel
	
	err := fmt.Errorf("create conn tunnel processing occurs error")

	if conn == nil {
		return nil, nil
	}

	addr := conn.LocalAddr()

	if tcpaddr, ok := addr.(*net.TCPAddr); !ok {
		return nil, err
	} else {
		spdy, err := spdystream.NewConnection(conn, false)
		if err != nil {
			golog.Error("Spdy", "NewConnTunnel", fmt.Sprintf("spdystream create connection error: %s", err), 0)
			return nil, err
		}

		go spdy.Serve(spdystream.NoOpStreamHandler)
		if _, err = spdy.Ping(); err != nil {
			return nil, err
		}

		spdyconn = &connTunnel{conn: spdy, tcpAddr: tcpaddr, switching: false}
	}

	return spdyconn, nil
}

func NewStreamConn(addr, port string) (*connTunnel, error) {
	conn, err := net.DialTimeout("tcp", addr+":"+port, time.Second)
	if err != nil {
		//log.Printf("dail spdy error: %s", err)
		return nil, err
	}

	connTunnel, err := NewConnTunnel(conn)

	return connTunnel, err
}

func CreateSpdySession(request *spdySession, ready chan<- *spdySession) {
	for {
		addrs := strings.Split(request.backend.address, ":")
		if conn, err := NewStreamConn(addrs[0], *streamPort); err == nil {
			request.spdy = conn
			golog.Info("Spdy", "CreateSpdySession", fmt.Sprintf("Created new session for: %s", request.backend.address), 0)
			break
		}
		time.Sleep(time.Second)
	}
	ready <- request
}

func (ct *connTunnel) Close() error {
	ct.conn.Close()
	return nil
}
