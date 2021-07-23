package main

import (
	"fmt"
	"net"
)

const (
	IP   = "0.0.0.0"
	Port = "8002"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", IP, Port))
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

	fmt.Println("Listening on port", Port)

	rxSeqNum := 0
	for {
		buf := make([]byte, 1024)
		n, cliAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("ReadFromUDP err:", err)
			return
		}
		txSeqNum := int(buf[n-1]) - 48
		if txSeqNum == rxSeqNum {
			fmt.Printf("Want %d, received %d, send back ACK %d\n", rxSeqNum, txSeqNum, rxSeqNum+1)
			rxSeqNum++
		} else {
			fmt.Printf("Want %d, received %d, drop it\n", rxSeqNum, txSeqNum)
		}

		_, err = conn.WriteToUDP([]byte(fmt.Sprintf("ACK %d", rxSeqNum)), cliAddr)
		if err != nil {
			fmt.Println("WriteToUDP err:", err)
			return
		}

		buf = nil
	}
}
