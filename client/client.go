package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"
	. "ve489/util"
)

// var num = flag.Int("n", 20, "Input how many times")
var ip = flag.String("i", "10.3.80.2", "Server IP")
var port = flag.Int("p", 8002, "Server Port")
var bytes = flag.Int("b", 1000, "Bytes per message")

func main() {
	flag.Parse()

	data, err := ioutil.ReadFile("/root/shakespeare.txt")
	if err != nil {
		fmt.Println("ReadFile error:", err)
		return
	}

	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", *ip, *port))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Dial complete")

	txSeqNum, count, length := false, 0, 0
	for i := 0; i < len(data); i += *bytes {
		if i+*bytes < len(data) {
			length = *bytes
		} else {
			length = len(data) - i
		}
		msg := make([]byte, length+1)
		for j := 0; j < length; j++ {
			msg[j] = data[j+i]
		}
		msg[length] = Bool2Byte(txSeqNum)

		_, err = conn.Write(msg)
		if err != nil {
			fmt.Println("conn.Write err:", err)
		}
		fmt.Printf("Sent Message %d (Seq: %d)\n", count, Bool2Int(txSeqNum))

		err = conn.SetReadDeadline(time.Now().Add(800 * time.Millisecond))
		if err != nil {
			fmt.Println("conn.SetReadDeadline err:", err)
		}

		buf := make([]byte, 2048)
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				fmt.Printf("Waiting for ACK timeout, will resend Message %d (Seq: %d)\n", count, Bool2Int(txSeqNum))
				i -= *bytes
			} else {
				fmt.Println("Read error:", err)
				return
			}
		} else {
			fmt.Println("Received from Server:", string(buf[:n]))
			count++
			txSeqNum = !txSeqNum
		}

		// time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("Client exits")
}
