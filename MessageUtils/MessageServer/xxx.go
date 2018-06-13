package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/zx9229/zxgo/zxxorm"
)

func ReportReq_2_ReportData(req *ReportReq) *ReportData {
	data := new(ReportData)
	//
	data.Id = 0
	data.Time = time.Time{}
	//
	data.UserId = req.UserId
	data.RefId = req.RefId
	data.RefTime = req.RefTime
	data.Status = req.Status
	data.Message = req.Message
	data.Group1 = req.Group1
	data.Group2 = req.Group2
	data.Group3 = req.Group3
	data.Group4 = req.Group4
	//
	return data
}

type DataCenter struct {
	myDb *MyXormDb
}

func New_DataCenter() *DataCenter {
	curData := new(DataCenter)
	curData.myDb = new_MyXormDb()
	return curData

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
		session := self.myDb.engine.NewSession()
		defer func() {
			session.Close()
		}()
		var needRollback bool = true
		defer func() {
			if needRollback {
				session.Rollback()
			}
		}()
		if err = session.Begin(); err != nil {
			break
		}
		if err = zxxorm.SessionInsertOne(session, ReportReq_2_ReportData(dataReq)); err != nil {
			break
		}
		if dataRsp.Id, err = QueryAndGetId(session, dataReq.UserId, dataReq.RefId); err != nil {
			break
		}
		if err = session.Commit(); err != nil {
			break
		}
		needRollback = false
		//
		dataRsp.UserId = dataReq.UserId
		dataRsp.RefId = dataReq.RefId
		dataRsp.Code = 0
		dataRsp.Message = "SUCCESS"
	}

	if dataRsp.Id == 0 {
		dataRsp.Code = 2
		dataRsp.Message = err.Error()
	}

	if bytes, err := json.Marshal(dataRsp); err != nil {
		panic(err)
	} else {
		fmt.Fprintf(w, string(bytes))
	}
}

func main() {
	var listenAddr string = "localhost:8080"
	simpleHttpServer := New_SimpleHttpServer(listenAddr)
	dataCenter := New_DataCenter()
	if true {
		dataCenter.myDb.Init()
	}
	simpleHttpServer.GetHttpServeMux().HandleFunc("/", dataCenter.Handler_ROOT)
	simpleHttpServer.Run()
}
