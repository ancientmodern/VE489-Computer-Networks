package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

var num = flag.Int("num", 5, "Input how many times")

func main() {
	flag.Parse()

	conn, err := net.Dial("udp", "10.3.39.2:8002")
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Dial complete")

	count := 0
	for i := 0; i < 20; i++ {
		_, err = conn.Write([]byte(fmt.Sprintf("Message %d", count)))
		count++
		if err != nil {
			fmt.Println("conn.Write err:", err)
		}

		err = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if err != nil {
			fmt.Println("conn.SetReadDeadline err:", err)
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				fmt.Println("Waiting for ACK timeout, resend...")
				count--
			} else {
				return
			}
		}
		fmt.Println("Received from Server:", string(buf[:n]))

		time.Sleep(1 * time.Second)
	}

	fmt.Println("Client ends.")
}
