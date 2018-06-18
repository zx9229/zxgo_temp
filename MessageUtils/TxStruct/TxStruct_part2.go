package TxStruct

import (
	"encoding/json"
	"reflect"
)

////////////////////////////////////////////////////////////////////

func (self *AddAgentReq) GET_TN() string {
	return self.TN
}

func (self *AddAgentReq) CALC_TN(modifyTN bool) string {
	TypeName := reflect.ValueOf(*self).Type().Name()
	if modifyTN {
		self.TN = TypeName
	}
	return TypeName
}

func (self *AddAgentReq) TO_JSON(panicWhenError bool) string {
	if bytes, err := json.Marshal(self); err != nil {
		if panicWhenError {
			panic(err)
		}
		return ""
	} else {
		return string(bytes)
	}
}

////////////////////////////////////////////////////////////////////

func (self *AddAgentRsp) GET_TN() string {
	return self.TN
}

func (self *AddAgentRsp) CALC_TN(modifyTN bool) string {
	TypeName := reflect.ValueOf(*self).Type().Name()
	if modifyTN {
		self.TN = TypeName
	}
	return TypeName
}

func (self *AddAgentRsp) TO_JSON(panicWhenError bool) string {
	if bytes, err := json.Marshal(self); err != nil {
		if panicWhenError {
			panic(err)
		}
		return ""
	} else {
		return string(bytes)
	}
}

////////////////////////////////////////////////////////////////////
