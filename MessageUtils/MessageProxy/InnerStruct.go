package main

import (
	"errors"
	"strconv"
	"time"

	"github.com/zx9229/zxgo"
	"github.com/zx9229/zxgo_temp/MessageUtils/TxStruct"
)

func ProxyReqRsp_ToReq(src *TxStruct.ProxyReqRsp) *TxStruct.ReportReq {
	dst := new(TxStruct.ReportReq)
	dst.UserId = src.UserId
	dst.RefId = src.RefId
	dst.RefTime = src.RefTime
	dst.Status = src.Status
	dst.Message = src.Message
	dst.Group1 = src.Group1
	dst.Group2 = src.Group2
	dst.Group3 = src.Group3
	dst.Group4 = src.Group4
	return dst
}

func ProxyReqRsp_FillWithRsp(src *TxStruct.ProxyReqRsp, data *TxStruct.ReportRsp, doCheck bool) error {
	if doCheck && (src.UserId != data.UserId || src.RefId != data.RefId) {
		return errors.New("check fail")
	}
	src.RspId = data.Id
	src.RspCode = data.Code
	src.RspMessage = data.Message
	return nil
}

type ConfigInfoField struct {
	Key        string    `xorm:"notnull pk unique"`
	Value      string    //字段的值.
	UpdateTIme time.Time `xorm:"updated"`
}

type ConfigInfo struct {
	Host          string
	Port          int
	ScanInterval  int
	RetryInterval int
}

func (self *ConfigInfo) From(fields []ConfigInfoField) {
	allKv := make(map[string]string)
	for _, field := range fields {
		allKv[field.Key] = field.Value
	}
	zxgo.ModifyByMap(self, allKv, true)
}

func (self *ConfigInfo) To() []ConfigInfoField {
	mapData := make(map[string]string) //TODO:待优化成一个通用函数.
	mapData["Host"] = self.Host
	mapData["Port"] = strconv.Itoa(self.Port)
	mapData["ScanInterval"] = strconv.Itoa(self.ScanInterval)
	mapData["RetryInterval"] = strconv.Itoa(self.RetryInterval)
	//
	slice_ := make([]ConfigInfoField, 0)
	for k, v := range mapData {
		slice_ = append(slice_, ConfigInfoField{Key: k, Value: v})
	}
	//
	return slice_
}

type ExeInfoField struct {
	Key        string    `xorm:"notnull pk unique"`
	Value      string    //字段的值.
	UpdateTIme time.Time `xorm:"updated"`
}

type ExeInfo struct {
	Pid     int
	Pname   string
	Exe     string
	Workdir string
}

func (self *ExeInfo) From(fields []ExeInfoField) {
	allKv := make(map[string]string)
	for _, field := range fields {
		allKv[field.Key] = field.Value
	}
	zxgo.ModifyByMap(self, allKv, true)
}

func (self *ExeInfo) To() []ExeInfoField {
	mapData := make(map[string]string) //TODO:待优化成一个通用函数.
	mapData["Pid"] = strconv.Itoa(self.Pid)
	mapData["Pname"] = self.Pname
	mapData["Exe"] = self.Exe
	mapData["Workdir"] = self.Workdir
	//
	slice_ := make([]ExeInfoField, 0)
	for k, v := range mapData {
		slice_ = append(slice_, ExeInfoField{Key: k, Value: v})
	}
	//
	return slice_
}
