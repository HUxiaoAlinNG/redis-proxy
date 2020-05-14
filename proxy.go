package main

import (
	"net"
	"strings"
)

//监听连接
func StartProxy(connPool *ConnPool, addr string) {
	n := "unix"
	if strings.Contains(addr, ":") {
		n = "tcp"
	}

	l, err := net.Listen(n, addr)
	if err != nil {
		SLogger.Error().AnErr("监听失败", err)
		return
	}
	defer l.Close()

	for {
		local, err := l.Accept()
		if err != nil {
			SLogger.Warn().AnErr("accept 失败", err)
			continue
		}

		go HandlerData(connPool, local)
	}

}

//数据交换方法
func HandlerData(connPool *ConnPool, local net.Conn) {

	SLogger.Debug().Msg(local.RemoteAddr().String())

	conn, err := connPool.Get()
	if err != nil {
		local.Close()
		SLogger.Error().AnErr("pool get error", err)
		return
	}

	forceClose := conn.SwapData(local, connPool.opt)
	local.Close()
	connPool.Put(conn, forceClose)

}
