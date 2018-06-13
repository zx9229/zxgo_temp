package main

import (
	"errors"
	"log"
	"time"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

type MyXormDb struct {
	engine *xorm.Engine
}

func new_MyXormDb() *MyXormDb {
	curData := new(MyXormDb)
	curData.engine = nil
	return curData
}

func (self *MyXormDb) Init() {
	var err error
	if self.engine, err = xorm.NewEngine("sqlite3", "./test.db"); err != nil {
		log.Println(err)
		return
	}

	self.engine.SetMapper(core.GonicMapper{}) //支持struct为驼峰式命名,表结构为下划线命名之间的转换,同时对于特定词支持更好.

	if location, err := time.LoadLocation("Asia/Shanghai"); err != nil {
		log.Println(err)
		return
	} else {
		self.engine.TZLocation = location
	}

	beans := make([]interface{}, 0)
	beans = append(beans, new(ReportReq))
	beans = append(beans, new(ReportRsp))
	beans = append(beans, new(ReportData))

	if err = self.engine.CreateTables(beans...); err != nil { //应该是:只要存在这个tablename,就跳过它.
		log.Println(err)
	}
	if err = self.engine.Sync2(beans...); err != nil { //同步数据库结构
		log.Println(err)
	}
}

func QueryAndGetId(session *xorm.Session, userId int64, refId int64) (id int64, err error) {
	dataSlice := make([]*ReportData, 0)
	queryCond := new(ReportData)
	queryCond.UserId = userId
	queryCond.RefId = refId
	if err = session.Find(&dataSlice, queryCond); err != nil {
		return
	}
	if len(dataSlice) != 1 {
		err = errors.New("个数有问题")
		return
	}
	if dataSlice[0].Id <= 0 {
		err = errors.New("数据有问题")
		return
	}
	id = dataSlice[0].Id
	return
}
