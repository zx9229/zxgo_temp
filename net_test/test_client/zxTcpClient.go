package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	var c net.Conn = nil
	var err error = nil
	c, err = net.Dial("tcp", "127.0.0.1:6889")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	time.Sleep(time.Second * 2)
	c.Write([]byte("qwertasdfg"))
	time.Sleep(time.Second * 1)
	c.Close()
	fmt.Println("will close")
}
