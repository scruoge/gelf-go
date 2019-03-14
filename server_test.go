package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	go server()
	time.Sleep(time.Millisecond)
	raddr, err := net.ResolveUDPAddr("udp", ":12201")
	if err != nil {
		t.Error(err)
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		t.Error(err)
	}
	p := make([]byte, 8192)
	s:="1f8b0800b7628a5c0003cb48cdc9c9070086a6103605000000"
	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	p = data;
	n, err := conn.Write(p);
	if err != nil {
		t.Error(err)
	}
	fmt.Println(n)
	time.Sleep(time.Millisecond)
	//err = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	//if err != nil {
	//	t.Error(err)
	//}
}
