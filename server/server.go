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

	fmt.Println("Listen complete")

	count := 0
	for {
		buf := make([]byte, 1024)
		n, cliAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("ReadFromUDP err:", err)
			return
		}
		fmt.Printf("Count: %d, Content: %s\n", count, string(buf[:n]))
		count++

		_, err = conn.WriteToUDP([]byte("0123456789"), cliAddr)
		if err != nil {
			fmt.Println("WriteToUDP err:", err)
			return
		}

		buf = nil
	}
}
