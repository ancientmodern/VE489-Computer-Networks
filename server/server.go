package main

import (
	"fmt"
	"net"
	"ve489/util"
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

	rxSeqNum, count := false, 0
	for {
		buf := make([]byte, 1024)
		n, cliAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("ReadFromUDP err:", err)
			return
		}
		txSeqNum := util.Int2Bool(int(buf[n-1]) - 48)
		if txSeqNum == rxSeqNum {
			fmt.Printf("Want %d, received %d, send back ACK %d\n", util.Bool2Int(rxSeqNum), util.Bool2Int(txSeqNum), util.Bool2Int(!rxSeqNum))
			rxSeqNum = !rxSeqNum
			count++
			fmt.Println("Totally received", count)
		} else {
			fmt.Printf("Want %d, received %d, drop it\n", util.Bool2Int(rxSeqNum), util.Bool2Int(txSeqNum))
		}

		_, err = conn.WriteToUDP([]byte(fmt.Sprintf("ACK %d", util.Bool2Int(rxSeqNum))), cliAddr)
		if err != nil {
			fmt.Println("WriteToUDP err:", err)
			return
		}

		buf = nil
	}
}
