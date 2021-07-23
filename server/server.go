package main

import (
	"fmt"
	"net"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "10.3.39.2:8002")
	if err != nil {
		fmt.Println("ResolveUDPAddr err:", err)
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("ListenUDP err:", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	n, cliAddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return
	}
	fmt.Println("Received from Client：", string(buf[:n]))

	_, err = conn.WriteToUDP([]byte("nice to see u in udp"), cliAddr)
	if err != nil {
		fmt.Println("WriteToUDP err:", err)
		return
	}
}
