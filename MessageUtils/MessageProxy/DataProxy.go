package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/zx9229/zxgo/zxxorm"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
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

		if 0 < len(locationName) {
			if location, err2 := time.LoadLocation(locationName); err2 != nil {
				err = err2
				break
			} else {
				self.engine.TZLocation = location
			}
		}

		beans := make([]interface{}, 0)
		beans = append(beans, &ConfigInfoField{})
		beans = append(beans, &ExeInfoField{})
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

func (self *DataProxy) LoadConfigInfo() (cfgInfo *ConfigInfo, err error) {
	slice_ := make([]ConfigInfoField, 0)
	if err = self.engine.Find(&slice_); err != nil {
		return
	}
	cfgInfo = new(ConfigInfo)
	cfgInfo.From(slice_)
	return
}

func (self *DataProxy) SaveConfigInfo(cfgInfo *ConfigInfo) error {
	var err error
	slice_ := cfgInfo.To()
	for _, kv := range slice_ {
		if err = zxxorm.Upsert(self.engine, kv); err != nil {
			return err
		}
	}
	return err
}

func (self *DataProxy) SaveExeInfo(exeInfo *ExeInfo) error {
	var err error
	slice_ := exeInfo.To()
	for _, kv := range slice_ {
		if err = zxxorm.Upsert(self.engine, kv); err != nil {
			return err
		}
	}
	return err
}

func (self *DataProxy) FlushExeInfo() error {
	var err error

	exeInfo := new(ExeInfo)
	exeInfo.Pid = os.Getpid()
	exeInfo.Pname = filepath.Base(os.Args[0])
	exeInfo.Workdir, _ = os.Getwd()
	exeInfo.Exe = filepath.Join(exeInfo.Workdir, exeInfo.Pname)

	if err = self.SaveExeInfo(exeInfo); err != nil {
		return err
	}

	return err
}
