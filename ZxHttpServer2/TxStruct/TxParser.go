package TxStruct

import (
	"encoding/json"
	"errors"
	"reflect"
)

type TxParser struct {
	mapStr2Type map[string]reflect.Type
}

func New_TxParser() *TxParser {
	newData := new(TxParser)
	newData.mapStr2Type = CalcMapStr2Type()
	return newData
}

func CalcMapStr2Type() map[string]reflect.Type {
	slice_ := make([]interface{}, 0)
	slice_ = append(slice_, LoginReq{})
	slice_ = append(slice_, LoginRsp{})

	cacheData := map[string]reflect.Type{}
	for _, element := range slice_ {
		curType := reflect.ValueOf(element).Type()
		cacheData[curType.Name()] = curType
	}

	return cacheData
}

func (self *TxParser) ParseString(jsonStr string) (objData interface{}, objType reflect.Type, err error) {
	return self.ParseByteSlice([]byte(jsonStr))
}

// objData:反序列化jsonByte后,得到的对象; objType:对象的类型; err:错误的详细情况.
func (self *TxParser) ParseByteSlice(jsonByte []byte) (objData interface{}, objType reflect.Type, err error) {
	objData = nil
	objType = nil
	err = nil

	var baseData *TxBaseData = &TxBaseData{}
	if err = json.Unmarshal(jsonByte, baseData); err != nil {
		return
	}

	var ok bool
	if objType, ok = self.mapStr2Type[baseData.Type]; !ok {
		err = errors.New("根据字符串找不到对应的类型")
		return
	}

	objData = reflect.New(objType).Interface()
	if err = json.Unmarshal(jsonByte, objData); err != nil {
		return
	}

	return
}
