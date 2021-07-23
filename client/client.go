package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("udp", "10.3.39.2:8002")
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
