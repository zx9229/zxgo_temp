package main

import (
	"time"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
)

type DataProxy struct {
	engine *xorm.Engine
}

func new_DataProxy() *DataProxy {
	curData := new(DataProxy)
	curData.engine = nil
	return curData

}

func (self *DataProxy) Init(driverName string, dataSourceName string, locationName string) error {
	var err error

	for _ = range "1" {
		if self.engine, err = xorm.NewEngine(driverName, dataSourceName); err != nil {
			break
		}

		self.engine.SetMapper(core.GonicMapper{}) //支持struct为驼峰式命名,表结构为下划线命名之间的转换,同时对于特定词支持更好.

		if location, err2 := time.LoadLocation(locationName); err2 != nil {
			err = err2
			break
		} else {
			self.engine.TZLocation = location
		}

		beans := make([]interface{}, 0)
		beans = append(beans, new(KeyValue))
		beans = append(beans, new(ReportReqRsp))

		if err = self.engine.CreateTables(beans...); err != nil { //应该是:只要存在这个tablename,就跳过它.
			break
		}

		if err = self.engine.Sync2(beans...); err != nil { //同步数据库结构
			break
		}
	}
	return err
}

func (self *DataProxy) QueryData() (slice_ []ReportReqRsp, err error) {
	slice_ = make([]ReportReqRsp, 0)
	if err = self.engine.In("is_handled", 0).Find(&slice_); err != nil { //TODO:字符串可能会变的.
		slice_ = nil
	}
	return
}
