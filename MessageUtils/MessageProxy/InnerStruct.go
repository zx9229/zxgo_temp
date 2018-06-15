package main

import (
	"errors"
	"time"

	"github.com/zx9229/zxgo_temp/MessageUtils/TxStruct"
)

type KeyValue struct {
	Key   string `xorm:"notnull pk unique"`
	Value string
}

type ReportReqRsp struct {
	UserId     int64     //(req)
	RefId      int64     `xorm:"notnull pk autoincr"` //(req)rowId
	RefTime    time.Time `xorm:"created"`             //这个Field将在Insert时自动赋值为当前时间
	Status     int       //(req) (三态) 0=>正常;1=>警告;其他值=>错误
	Message    string    //(req)
	Group1     string    //(req)
	Group2     string    //(req)
	Group3     string    //(req)
	Group4     string    //(req)
	IsHandled  int       //是否处理过了(0:尚未处理)
	RspId      int64     //(rsp)
	RspCode    int       //(rsp)
	RspMessage string    //(rsp)
	UpdateTime time.Time `xorm:"updated"` //这个Field将在Insert或Update时自动赋值为当前时间
}

func (self *ReportReqRsp) ToReq() *TxStruct.ReportReq {
	dst := new(TxStruct.ReportReq)
	dst.UserId = self.UserId
	dst.RefId = self.RefId
	dst.RefTime = self.RefTime
	dst.Status = self.Status
	dst.Message = self.Message
	dst.Group1 = self.Group1
	dst.Group2 = self.Group2
	dst.Group3 = self.Group3
	dst.Group4 = self.Group4
	return dst
}

func (self *ReportReqRsp) FillWithRsp(data *TxStruct.ReportRsp, doCheck bool) error {
	if doCheck && (self.UserId != data.UserId || self.RefId != data.RefId) {
		return errors.New("check fail")
	}
	self.RspId = data.Id
	self.RspCode = data.Code
	self.RspMessage = data.Message
	return nil
}
