package main

import (
	"errors"

	"github.com/zx9229/zxgo_temp/MessageUtils/TxStruct"
)

func (self *DataCenter) check_AddAgentReq(dataReq *TxStruct.AddAgentReq) error {
	var err error
	for _ = range "1" {
		if dataReq.Id <= 0 {
			err = errors.New("dataReq.Id <= 0")
			break
		}
		if dataReq.Id < int64(len(self.infoSlice)) {
			if self.infoSlice[dataReq.Id-1] != nil {
				err = errors.New("already exitst.")
				break
			}
		}
	}
	return err
}
