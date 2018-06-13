package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

type DataCenter struct {
	engine *xorm.Engine
}

func New_DataCenter() *DataCenter {
	curData := new(DataCenter)
	curData.engine = nil
	return curData

}

func (self *DataCenter) Init(driverName string, dataSourceName string, locationName string) error {
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
		beans = append(beans, new(ReportReq))
		beans = append(beans, new(ReportRsp))
		beans = append(beans, new(ReportData))

		if err = self.engine.CreateTables(beans...); err != nil { //应该是:只要存在这个tablename,就跳过它.
			break
		}

		if err = self.engine.Sync2(beans...); err != nil { //同步数据库结构
			break
		}
	}
	return err
}

func (self *DataCenter) Handler_ROOT(w http.ResponseWriter, r *http.Request) {
	var err error
	var byteSlice []byte
	var dataRsp *ReportRsp = new(ReportRsp)
	//curl -d "{\"a\":123}" http://localhost:8080
	for _ = range "1" {
		if r.Method != "POST" {
			break
		}

		defer r.Body.Close()
		if byteSlice, err = ioutil.ReadAll(r.Body); err != nil {
			break
		}

		dataReq := new(ReportReq)
		if err = json.Unmarshal(byteSlice, dataReq); err != nil {
			break
		}

		//TODO:校验通过.
		if dataRsp.Id, err = InsertOne(self.engine, ReportReq_2_ReportData(dataReq)); err != nil {
			break
		}
		//
		dataRsp.UserId = dataReq.UserId
		dataRsp.RefId = dataReq.RefId
		dataRsp.Code = 0
		dataRsp.Message = "SUCCESS"
	}

	if dataRsp.Id <= 0 {
		dataRsp.Code = 2
		dataRsp.Message = err.Error()
	}

	if bytes, err := json.Marshal(dataRsp); err != nil {
		panic(err)
	} else {
		fmt.Fprintf(w, string(bytes))
	}
}
