package main

import (
	"fmt"
	"time"

	"github.com/go-xorm/xorm"
	"github.com/zx9229/zxgo/zxxorm"
	"github.com/zx9229/zxgo_temp/MessageUtils/TxStruct"
)

func InsertOne(engine *xorm.Engine, data *TxStruct.ReportData) (id int64, err error) {
	session := engine.NewSession()
	defer session.Close()
	var needRollback bool = true
	defer func() {
		if needRollback {
			session.Rollback()
		}
	}()
	for _ = range "1" {
		if err = session.Begin(); err != nil {
			break
		}
		if err = zxxorm.SessionInsertOne(session, data); err != nil {
			break
		}
		//成功执行insert操作后,xorm内部自动对data的Id填值了,因为传进去的data是一个指针,所以把Id的值带出来了.
		if data.Id <= 0 {
			err = fmt.Errorf("should be a positive number, actually %v", data.Id)
			break
		}
		if err = session.Commit(); err != nil {
			break
		}
		needRollback = false
		id = data.Id
	}
	return
}

func ReportReq_2_ReportData(req *TxStruct.ReportReq) *TxStruct.ReportData {
	data := new(TxStruct.ReportData)
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

func AddAgentReq_2_AgentInfo(dataReq *TxStruct.AddAgentReq) *TxStruct.AgentInfo {
	data := new(TxStruct.AgentInfo)
	//
	data.Id = dataReq.Id
	data.Memo = dataReq.Memo
	//
	return data
}
