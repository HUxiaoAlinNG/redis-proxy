#poolproxy

## Introduction
redis-proxy是一个使用golang编写的简单reids代理工具
原仓库：https://github.com/bjdgyc/poolproxy

该工具提供了连接池功能，并可设置最大连接数、连接最大空闲时间、定时检测并断开空闲连接

该工具提供了一个透明的代理接口，可以为下游程序提供带连接池的代理功能

## 功能
目前实现了redis的代理功能

本地使用时，建议监听`Unix domain socket`
可以有效减少TCP握手消耗，提高系统性能

## Toml config

``` toml

[options.redis]
    # 代理监听设置
    # 可以设置为Unix socket
    # 如: /var/run/poolproxy.socket
    # 也可以设置为TCP端口
    addr = ":8080"
    read_timeout = 0
    write_timeout = 0
    pool_timeout = 0

    #远程连接设置
    #redis-server地址
    raddr = "192.168.56.102:6379"
    #redis-server AUTH密码
    rpass = ""
    #实际代理的redis-server 数据库，若为""，则默认使用redis默认的0数据库
    rdb   = ""
    rpool_size = 0
    # 获取空闲连接的排队超时时间（秒）
    ridle_timeout = 0
    # 定期检测空闲连接的时间（秒）
    ridle_check_frequency = 120
    
```

## Start

`go build && ./poolproxy -c ./config.toml`
