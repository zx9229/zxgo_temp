package main

import (
	"log"
	"time"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
)

func main() {
	var err error = nil
	var engine *xorm.Engine = nil

	if engine, err = xorm.NewEngine("sqlite3", "./test.db"); err != nil {
		log.Println(err)
		return
	}

	if location, err := time.LoadLocation("Asia/Shanghai"); err != nil {
		log.Println(err)
		return
	} else {
		engine.TZLocation = location
	}

	engine.SetMapper(core.GonicMapper{}) //支持struct为驼峰式命名,表结构为下划线命名之间的转换,同时对于特定词支持更好.

	if err = CreateTablesAndSync2(engine); err != nil {
		return
	}

	//if affected, err := engine.Table(cmr2.MyTn).InsertOne(cmr); err != nil {
}
