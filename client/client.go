package main

import (
	"flag"
	"fmt"
	"net"
	"time"
)

var num = flag.Int("num", 5, "Input how many times")

func main() {
	flag.Parse()

	conn, err := net.DialTimeout("udp", "10.3.39.2:8002", time.Second)
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte("0123456789"))
	if err != nil {
		fmt.Println("conn.Write err:", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}
	fmt.Println("Received from Server:", string(buf[:n]))
}
