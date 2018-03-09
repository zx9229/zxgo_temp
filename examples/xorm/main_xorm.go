//go get -u -v github.com/go-xorm/xorm
//go get -u -v github.com/mattn/go-sqlite3

//使用Table和Tag改变名称映射
//http://gobook.io/read/github.com/go-xorm/manual-zh-CN/chapter-02/3.tags.html
//Column属性定义
//http://gobook.io/read/github.com/go-xorm/manual-zh-CN/chapter-02/4.columns.html
//Go与字段类型对应表
//http://gobook.io/read/github.com/go-xorm/manual-zh-CN/chapter-02/5.types.html
//同步数据库结构
//http://gobook.io/read/github.com/go-xorm/manual-zh-CN/chapter-03/4.sync.html

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

type KeyValue struct {
	Key   string `xorm:"pk notnull"`
	Value string
}

type PushMessage struct {
	Id         int64     `xorm:"pk autoincr unique notnull"`
	Dttm       time.Time `xorm:"updated"`
	Groups     []string
	Users      []string
	PushType   []string
	MajorLevel string //DEBUG,INFO,WARN,ERROR,FATAL
	MinorLevel int
	Message    string
	Memo       string //备注(预留字段)
}

func main() {
	var err error = nil
	var engine *xorm.Engine = nil
	if engine, err = xorm.NewEngine("sqlite3", "./test.db"); err != nil {
		log.Println(err)
		return
	}

	engine.SetMapper(core.GonicMapper{}) //支持struct为驼峰式命名,表结构为下划线命名之间的转换,同时对于特定词支持更好.

	if location, err := time.LoadLocation("Asia/Shanghai"); err != nil {
		log.Println(err)
		return
	} else {
		engine.TZLocation = location
	}

	if err = engine.CreateTables(new(PushMessage), new(KeyValue)); err != nil { //应该是:只要存在这个tablename,就跳过它.
		log.Println(err)
	}
	if err = engine.Sync2(new(PushMessage), new(KeyValue)); err != nil { //同步数据库结构
		log.Println(err)
	}

	insertMsgSlice := make([]PushMessage, 0)
	for i := 1; i < 9; i++ {
		pushMsg := PushMessage{}
		pushMsg.Users = []string{"u1", "u2", "u3"}
		pushMsg.MajorLevel = "INFO"
		pushMsg.MinorLevel = i + 100
		pushMsg.Message = fmt.Sprintf("自动生成了第%v个消息内容", i)
		insertMsgSlice = append(insertMsgSlice, pushMsg)
	}
	if affected, err := engine.Insert(insertMsgSlice); err != nil {
		log.Println(affected, err)
	}

	if affected, err := engine.Insert(KeyValue{"pushid", "6"}); err != nil {
		log.Println(affected, err)
	}

	findMsgSlice1 := make([]PushMessage, 0)
	if err = engine.Find(&findMsgSlice1); err != nil {
		log.Println(err)
	}

	findMsgSlice2 := make([]PushMessage, 0)
	if err = engine.SQL("SELECT * FROM push_message WHERE id > (SELECT value FROM key_value WHERE key='pushid' LIMIT 1)").Find(&findMsgSlice2); err != nil {
		log.Println(err)
	}

	log.Println("program will exit...")
}
