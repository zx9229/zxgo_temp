package TxStruct

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type Handler func(i interface{})

type innerPair struct {
	refType reflect.Type //类型.
	handler Handler      //回调函数.
}

type OrmData struct {
	mapStr2Data map[string]*innerPair
}

func New_OrmData() *OrmData {
	newData := new(OrmData)
	for k, v := range initMap() {
		newData.mapStr2Data[k] = &innerPair{v, nil}
	}
	return newData
}

func initMap() map[string]reflect.Type {
	slice_ := make([]interface{}, 0)
	slice_ = append(slice_, RspMessage{})
	slice_ = append(slice_, ChatMessage{})

	cacheData := map[string]reflect.Type{}
	for _, element := range slice_ {
		curType := reflect.ValueOf(element).Type()
		cacheData[curType.Name()] = curType
	}

	return cacheData
}

func (self *OrmData) AddHandler(curType reflect.Type, cbFun Handler) {
	for _, vData := range self.mapStr2Data {
		if vData.refType == curType {
			vData.handler = cbFun
			return
		}
	}
}

func (self *OrmData) test(jsonStr string) error {
	var err error
	var jsonByte []byte = []byte(jsonStr)

	var baseData *BaseTxData = &BaseTxData{}
	if err = json.Unmarshal(jsonByte, baseData); err != nil {
		return err
	}

	var ok bool
	var matchedData *innerPair
	if matchedData, ok = self.mapStr2Data[baseData.Type]; !ok {
		return errors.New("找不到对应的类型")
	}

	curValue := reflect.New(matchedData.refType).Interface()
	if err = json.Unmarshal(jsonByte, curValue); err != nil {
		return err
	}

	if matchedData.handler == nil {
		return errors.New("找不到对应的函数")
	}

	matchedData.handler(curValue)

	return err
}

func cbRspMessage(i interface{}) {
	fmt.Println(i)
}

func ThisIsExample() {

	xxObj := New_OrmData()
	xxObj.AddHandler(reflect.ValueOf(RspMessage{}).Type(), cbRspMessage)

	testMsg := RspMessage{}
	testMsg.Type = reflect.ValueOf(testMsg).Type().Name()
	testMsg.TransmitId = 1
	testMsg.Code = 2
	testMsg.Message = "testting..."
	if jsonByte, err := json.Marshal(testMsg); err != nil {
		panic(err)
	} else {
		if err := xxObj.test(string(jsonByte)); err != nil {
			panic(err)
		}
	}
}
