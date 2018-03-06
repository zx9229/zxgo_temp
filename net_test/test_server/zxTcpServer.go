package main

import (
	"fmt"
	"net"
	"time"
)

//go-tcpsock/server.go
func handleConn(c net.Conn) {
	defer c.Close()
	fmt.Println(time.Now(), "RemoteAddr:", c.RemoteAddr().String(), "LocalAddr:", c.LocalAddr().String())
	buf := make([]byte, 1024)
	for {
		lenght, err := c.Read(buf)
		if err != nil {
			break
		}
		if lenght > 0 {
			buf[lenght] = 0
		}
		fmt.Println(time.Now(), "Rec[", c.RemoteAddr().String(), "] Say :", string(buf[0:lenght]))
	}
	fmt.Println(time.Now(), "is_Closeed")
}

func main() {
	l, err := net.Listen("tcp", ":6889")
	if err != nil {
		fmt.Println("listen error:", err)
		return
	}

	for {
		var c net.Conn = nil
		c, err = l.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			break
		}
		// start a new goroutine to handle
		// the new connection.
		go handleConn(c)
	}
}
