package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"net/http"
	_ "net/http/pprof"
	"os"
)

var (
	connPool *ConnPool
	SLogger  zerolog.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
)

func main() {
	// 读取配置
	config, err := LoadConfig("./config.toml")
	if err != nil {
		fmt.Println("读取配置失败", err)
	}

	for _, opt := range config.Options {
		connPool = NewConnPool(opt)
		go StartProxy(connPool, opt.Addr)
		fmt.Println(connPool)
	}

	go func() {
		http.ListenAndServe("127.0.0.1:8090", nil)
	}()

	//connPool.Close()
	InitSignal()

}
