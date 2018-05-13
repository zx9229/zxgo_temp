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

	if err = myChat.addUser("a1", ""); err != nil {
		log.Println(err)
	}
	if err = myChat.addUser("b2", ""); err != nil {
		log.Println(err)
	}
	if err = myChat.AddFriends(1, 2); err != nil {
		log.Println(err)
	}

	for {
		time.Sleep(time.Second)
		fmt.Println(time.Now())
	}
}
