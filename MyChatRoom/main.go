package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
)

func main() {
	var err error = nil
	var engine *xorm.Engine = nil
	driverName := "sqlite3"
	dataSourceName := "test.db"
	if engine, err = xorm.NewEngine(driverName, dataSourceName); err != nil {
		log.Println(err)
		return
	}
	engine.SetMapper(core.GonicMapper{})
	if err = engine.CreateTables(UserData{}, GroupData{}); err != nil {
		log.Println(err)
		return
	}
	if err = engine.Sync2(UserData{}, GroupData{}); err != nil {
		log.Println(err)
		return
	}

	usergroup := NewUserGroup()
	if err = usergroup.LoadFromDb(engine); err != nil {
		log.Println(err)
		return
	}
	if err = usergroup.AddUserWithLock(engine, "a1", "pwd"); err != nil {
		log.Println(err)
		return
	}
}

func main_bak() {
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

		nmr := &PushMessageRaw{}
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
