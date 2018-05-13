package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	var err error = nil

	driverName := "sqlite3"
	dataSourceName := "test.db"
	locationName := "Asia/Shanghai"
	myChat := NewMyChat()
	if err = myChat.Init(driverName, dataSourceName, locationName); err != nil {
		log.Println(err)
		os.Exit(100)
	}

	for {
		time.Sleep(time.Second)
		fmt.Println(time.Now())
	}
}
