package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/go-xorm/xorm"
)

func calcMyTablename(engine *xorm.Engine, bean interface{}, index int) string {
	//我参考的代码 func (engine *Engine) tbName(v reflect.Value) string {
	if index <= 0 {
		panic(fmt.Sprintf("index=%d", index))
	}
	var v reflect.Value = reflect.Indirect(reflect.ValueOf(bean))
	var tbName string = engine.TableMapper.Obj2Table(reflect.Indirect(v).Type().Name())
	return fmt.Sprintf("%v_%d", tbName, index)
}

func CreateTablesAndSync2(engine *xorm.Engine) error {
	var err error = nil

	cmr1 := new(ChatMessageRaw)
	cmr1.MyTn = calcMyTablename(engine, cmr1, 1)

	cm1 := &ChatMessage{MyTn: calcMyTablename(engine, ChatMessage{}, 1)}

	for i := 0; i < 1; i++ {
		if err = engine.CreateTables(cmr1, cm1, KeyValue{}); err != nil { //应该是:只要存在这个tablename,就跳过它.
			break
		}
		if err = engine.Sync2(cmr1, cm1, new(KeyValue)); err != nil { //同步数据库结构
			break
		}
	}

	return err
}

func RenameTable(engine *xorm.Engine, oldTablename string) error {
	//sqlite3测试通过.
	newTablename := fmt.Sprintf("bak_%s_%s", oldTablename, time.Now().Format("20060102150405"))
	sqlStr := fmt.Sprintf("ALTER TABLE %v RENAME TO %v", oldTablename, newTablename)
	_, err := engine.Exec(sqlStr)
	return err
}

func xxxx(engine *xorm.Engine) error {
	//本系统内的表名的规定: TABLENAME_1, TABLENAME_1 存储满了之后, 就存储 TABLENAME_2, 以此类推.
	tableSlice, err := engine.DBMetas()
	if err != nil {
		return err
	}
	fmt.Println(tableSlice)
	panic("未实现")
}
