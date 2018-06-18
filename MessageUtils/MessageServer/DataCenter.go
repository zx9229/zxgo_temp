package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"unsafe"

	"github.com/zx9229/zxgo/zxxorm"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zx9229/zxgo_temp/MessageUtils/TxStruct"
)

type DataCenter struct {
	engine    *xorm.Engine          //
	infoSlice []*TxStruct.AgentInfo //不用map,无锁.
}

func New_DataCenter() *DataCenter {
	curData := new(DataCenter)
	curData.engine = nil
	curData.infoSlice = nil
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
		beans = append(beans, new(TxStruct.ReportReq))
		beans = append(beans, new(TxStruct.ReportRsp))
		beans = append(beans, new(TxStruct.ReportData))
		beans = append(beans, new(TxStruct.AgentInfo))

		if err = self.engine.CreateTables(beans...); err != nil { //应该是:只要存在这个tablename,就跳过它.
			break
		}

		if err = self.engine.Sync2(beans...); err != nil { //同步数据库结构
			break
		}
	}
	return err
}

func (self *DataCenter) Handler_ReportReq(w http.ResponseWriter, r *http.Request) {
	var err error
	var byteSlice []byte
	var dataRsp *TxStruct.ReportRsp = new(TxStruct.ReportRsp)
	//curl -d "{\"a\":123}" http://localhost:8080
	for _ = range "1" {
		if r.Method != "POST" {
			break
		}

		defer r.Body.Close()
		if byteSlice, err = ioutil.ReadAll(r.Body); err != nil {
			break
		}

		dataReq := new(TxStruct.ReportReq)
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

func (self *DataCenter) Handler_AddAgentReq(w http.ResponseWriter, r *http.Request) {
	var err error
	dataRsp := new(TxStruct.AddAgentRsp)
	//curl -d "{\"a\":123}" http://localhost:8080
	for _ = range "1" {
		dataReq := new(TxStruct.AddAgentReq)
		if err = ParseRequest(r, dataReq, false); err != nil {
			break
		}
		dataRsp.DataReq = dataReq
		if err = self.check_AddAgentReq(dataReq); err != nil {
			break
		}
		if err = zxxorm.EngineInsertOne(self.engine, AddAgentReq_2_AgentInfo(dataReq)); err != nil {
			break
		}
		dataRsp.Code = 0
		dataRsp.Message = "SUCCESS"
	}

	if err != nil {
		dataRsp.Code = -1
		dataRsp.Message = err.Error()
	}

	fmt.Fprintf(w, dataRsp.TO_JSON(true))
}

func (self *DataCenter) calcCacheData_1() (slice_ []*TxStruct.AgentInfo, err error) {
	dbDataSlice := make([]*TxStruct.AgentInfo, 0)
	if err = self.engine.Find(&dbDataSlice); err != nil {
		return
	}
	var maxId int64 = 0
	dbDataMap := make(map[int64]*TxStruct.AgentInfo)
	for _, item := range dbDataSlice {
		if item.Id <= 0 {
			panic(item.Id)
		}
		if maxId < item.Id {
			maxId = item.Id
		}
		dbDataMap[item.Id] = item
	}
	slice_ = make([]*TxStruct.AgentInfo, 0)
	for idx := int64(1); idx <= maxId; idx++ {
		if item, ok := dbDataMap[idx]; ok {
			slice_ = append(slice_, item)
		} else {
			slice_ = append(slice_, nil)
		}
	}
	return
}

func (self *DataCenter) calcCacheData_2(slice_ []*TxStruct.AgentInfo) error {
	var err error
	for _, item := range slice_ {
		if item == nil {
			continue
		}
		if item.LastRefId, err = self.calcLatestRefId(item.Id); err != nil {
			return err
		}
	}
	return err
}

func (self *DataCenter) calcCacheData() (slice_ []*TxStruct.AgentInfo, err error) {
	if slice_, err = self.calcCacheData_1(); err != nil {
		slice_ = nil
		return
	}
	if err = self.calcCacheData_2(slice_); err != nil {
		slice_ = nil
		return
	}
	return
}

func (self *DataCenter) calcLatestRefId(userId int64) (refId int64, err error) {
	data := new(TxStruct.ReportData)
	cn_UserId := zxxorm.GuessColName(self.engine, data, unsafe.Offsetof(data.UserId), true)
	cn_RefId := zxxorm.GuessColName(self.engine, data, unsafe.Offsetof(data.RefId), true)
	ok, err := self.engine.Where(fmt.Sprintf("%v = ?", cn_UserId), userId).Desc(cn_RefId).Limit(1).Get(data)
	if err != nil {
		return
	} else {
		if ok {
			refId = data.RefId
		} else {
			refId = 0
		}
	}
	return
}

func ParseRequest(r *http.Request, v TxStruct.TxInterface, checkTN bool) error {
	var err error
	var byteSlice []byte

	for _ = range "1" {
		if r == nil {
			err = errors.New("r == nil")
			break
		}

		defer r.Body.Close()

		if byteSlice, err = ioutil.ReadAll(r.Body); err != nil {
			break
		}

		if err = json.Unmarshal(byteSlice, v); err != nil {
			break
		}

		if checkTN {
			if v.CALC_TN(false) != v.GET_TN() {
				err = errors.New("checkTN fail")
				break
			}
		}
	}

	return err
}
