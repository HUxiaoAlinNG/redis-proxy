package main

import (
	"testing"
)

func getRedis(t *testing.T) *Redis {
	conn, err := NewConn(Option{})
	if err != nil {
		t.Fatal(err)
	}
	return &Redis{conn: conn}
}

func TestPing(t *testing.T) {
	redis := getRedis(t)
	err := redis.Ping()
	if err != nil {
		t.Error(err)
	}
	redis.conn.Close()
	t.Log("Ping sucess")
}

func TestAuth(t *testing.T) {
	redis := getRedis(t)
	err := redis.Auth("", "")
	if err != nil {
		t.Error(err)
	}
	redis.conn.Close()
	t.Log("Auth sucess")
}
