package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type Redis struct {
	conn *Conn
}

var _ Conner = (*Redis)(nil)

//redis的ping
func (cn *Redis) Ping() error {
	err := cn.conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	if err != nil {
		return nil
	}
	cb := <-cn.conn.GetReadChan()
	if cb.Err != nil {
		return cb.Err
	}
	if strings.ToUpper(string(cb.Byte)) != "+PONG\r\n" {
		return fmt.Errorf("error ping")
	}
	return nil
}

//redis 权限验证
func (cn *Redis) Auth(user, pass string) error {
	if pass == "" {
		return nil
	}

	data := fmt.Sprintf("*2\r\n$4\r\nAUTH\r\n$%d\r\n%s\r\n", len(pass), pass)
	err := cn.conn.Write([]byte(data))
	if err != nil {
		return err
	}
	cb := <-cn.conn.GetReadChan()
	if cb.Err != nil {
		return cb.Err
	}
	if strings.ToUpper(string(cb.Byte)) != "+OK\r\n" {
		return fmt.Errorf("auth error")
	}

	return nil
}

//redis 选择 分组
func (cn *Redis) Select(db string) error {
	if db == "" {
		return nil
	}
	// check
	_, err := strconv.Atoi(db)
	if err != nil {
		return err
	}
	data := fmt.Sprintf("*2\r\n$6\r\nSELECT\r\n$%d\r\n%s\r\n", len(db), db)
	err = cn.conn.Write([]byte(data))
	if err != nil {
		return err
	}
	cb := <-cn.conn.GetReadChan()
	if cb.Err != nil {
		return cb.Err
	}
	if strings.ToUpper(string(cb.Byte)) != "+OK\r\n" {
		return fmt.Errorf("db error")
	}

	return nil
}

//redis 数据读取
func (cn *Redis) ReadData() {
	var (
		line []byte
		err  error
		cb   *ChanBuf
	)
	for {
		line, err = cn.conn.BufReader.ReadBytes('\n')
		cn.conn.UsedAt = time.Now()
		cb = &ChanBuf{Byte: line, Err: err}
		cn.conn.ChanRead <- cb
		if err != nil {
			break
		}
	}
}

//redis数据交换
func (cn *Redis) SwapData(local net.Conn, opt Option) bool {
	localRead := bufio.NewReader(local)
	var (
		err           error
		forceClose    = false
		exitChanProxy = make(chan struct{})
		cb            *ChanBuf
	)

	//读取数据
	go func() {
		var (
			line []byte
			err  error
		)
		firstItem := ""
		arrLen := 0
		lines := ""
		for {
			// 第一次读取
			line, err = localRead.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					SLogger.Error().AnErr("local read error1:", err)
				}
				exitChanProxy <- struct{}{}
				break
			}
			if firstItem == "" {
				firstItem = string(line)
				// 命令起点，则在里面继续进行读取
				if len(firstItem) > 1 && string(firstItem[0]) == "*" {
					arrLen, err = strconv.Atoi(firstItem[1 : len(firstItem)-2])
					// 截取失败，直接交换数据
					if err != nil {
						firstItem = ""
						SLogger.Error().AnErr("截取 error:", err)
						err = connWrite(cn, line, &forceClose)
						if err != nil {
							exitChanProxy <- struct{}{}
						}
						// next
						continue
					}
					lines += string(line)
					// 截取成功后
					if arrLen > 0 {
						isAuth := false
						isSelect := false
						for i := 0; i < 2*arrLen; i++ {
							// 继续读取
							line, err = localRead.ReadBytes('\n')
							if err != nil {
								if err != io.EOF {
									SLogger.Error().AnErr("local read error2:", err)
								}
								break
							}
							if i == 1 && strings.ToUpper(string(line[0:len(line)-2])) == "AUTH" {
								isAuth = true
							}
							if opt.RDb != "" && i == 1 && strings.ToUpper(string(line[0:len(line)-2])) == "SELECT" {
								isSelect = true
							}
							if isAuth {
								if i == 2 {
									data := fmt.Sprintf("AUTH\r\n$%d\r\n%s\r\n", len(opt.RPass), opt.RPass)
									lines += data
								}
								if i == 3 {
									isAuth = false
								}
							} else if isSelect {
								if i == 2 {
									data := fmt.Sprintf("SELECT\r\n$%d\r\n%s\r\n", len(opt.RDb), opt.RDb)
									lines += data
								}
								if i == 3 {
									isSelect = false
								}
							} else {
								lines += string(line)
							}
						}
						b := bytes.NewBuffer([]byte(lines))
						for i := 0; i < 2*arrLen+1; i++ {
							b, _ := b.ReadBytes('\n')
							err = connWrite(cn, b, &forceClose)
							if err != nil {
								exitChanProxy <- struct{}{}
							}
						}
						lines = ""
					}
				} else {
					err = connWrite(cn, line, &forceClose)
					if err != nil {
						exitChanProxy <- struct{}{}
					}
				}
				firstItem = ""
			}
		}
	}()

	//客户端写回数据
	readChan := cn.conn.GetReadChan()
	for {
		select {
		case <-exitChanProxy:
			goto FAIL
		case cb = <-readChan:
			if cb.Err != nil {
				forceClose = true
				SLogger.Error().AnErr("remote readChan error:", err)
				goto FAIL
			}
			//fmt.Println(string(cb.Byte))
			_, err = local.Write(cb.Byte)
			if err != nil {
				SLogger.Error().AnErr("local write error:", err)
				goto FAIL
			}
		}
	}
FAIL:
	return forceClose
}

func connWrite(cn *Redis, b []byte, forceClose *bool) error {
	err := cn.conn.Write(b)
	if err != nil {
		isTrue := true
		forceClose = &isTrue
		SLogger.Error().AnErr("remote write error:", err)
		return err
	}
	return nil
}
