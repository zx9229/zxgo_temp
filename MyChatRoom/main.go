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

	var userAlias string = "a1"
	if err = myChat.AddUserWithLock(userAlias, "pwd"); err != nil {
		log.Println(err)
	}

	for i := 0; i < 30; i++ {
		time.Sleep(time.Second)

		nmr := PushMessageRaw{}
		nmr.RecverId = []int64{0}
		nmr.Message = fmt.Sprintf("msg_%v", i)
		if err = myChat.RecvPushMessageRaw(nmr); err != nil {
			log.Println(err)
		}

		time.Sleep(time.Second)

		if err = myChat.HandlePushMessage(nil, &userAlias); err != nil {
			log.Println(err)
		}

		log.Println("for, i:", i)
	}

	for {
		time.Sleep(time.Second)
		fmt.Println(time.Now())
	}
}
