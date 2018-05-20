package TxStruct

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"golang.org/x/net/websocket"
)

type TxParserHandler func(ws *websocket.Conn, i interface{})

type TxParser struct {
	mapStr2Data map[string]*onePair
}

type onePair struct {
	refType reflect.Type    //类型.
	handler TxParserHandler //回调函数.
}

func New_TxParser() *TxParser {
	newData := new(TxParser)
	newData.mapStr2Data = make(map[string]*onePair)

	for k, v := range CalcMapStr2Type() {
		newData.mapStr2Data[k] = &onePair{v, nil}
	}

	return newData
}

func CalcMapStr2Type() map[string]reflect.Type {
	slice_ := make([]interface{}, 0)
	slice_ = append(slice_, ChatMessage{})
	slice_ = append(slice_, ChatMessageRsp{})
	slice_ = append(slice_, PushMessage{})
	slice_ = append(slice_, PushMessageRsp{})

	cacheData := map[string]reflect.Type{}
	for _, element := range slice_ {
		curType := reflect.ValueOf(element).Type()
		cacheData[curType.Name()] = curType
	}

	return cacheData
}

func (self *TxParser) RegisterHandler(curType reflect.Type, curFun TxParserHandler) bool {
	for _, vData := range self.mapStr2Data {
		if vData.refType == curType {
			vData.handler = curFun
			return true
		}
	}
	return false
}

func (self *TxParser) ParseString(ws *websocket.Conn, jsonStr string) (objData interface{}, cbOk bool, err error) {
	return self.ParseByteSlice(ws, []byte(jsonStr))
}

// objData:反序列化jsonByte后,得到的对象; cbOk:成功调用对应的回调函数; err:错误的详细情况.
func (self *TxParser) ParseByteSlice(ws *websocket.Conn, jsonByte []byte) (objData interface{}, cbOk bool, err error) {
	objData = nil
	cbOk = false
	err = nil

	var baseData *TxBaseData = &TxBaseData{}
	if err = json.Unmarshal(jsonByte, baseData); err != nil {
		return
	}

	var ok bool
	var matchedData *onePair
	if matchedData, ok = self.mapStr2Data[baseData.Type]; !ok {
		err = errors.New("根据字符串找不到对应的类型")
		return
	}

	objData = reflect.New(matchedData.refType).Interface()
	if err = json.Unmarshal(jsonByte, objData); err != nil {
		return
	}

	if matchedData.handler == nil {
		err = errors.New("没有注册对应的回调函数")
		return
	}

	matchedData.handler(ws, objData)
	cbOk = true

	return
}

func ThisIsExample() {

	parser := New_TxParser()
	parser.RegisterHandler(reflect.ValueOf(ChatMessage{}).Type(), cb_ChatMessage)

	someObj := ChatMessage{}
	someObj.FillField_Type()
	someObj.Message = "data_for_test"
	if jsonByte, err := json.Marshal(someObj); err != nil {
		panic(err)
	} else {
		if _, _, err := parser.ParseByteSlice(nil, jsonByte); err != nil {
			panic(err)
		}
	}
}
func cb_ChatMessage(ws *websocket.Conn, i interface{}) {
	fmt.Println(i)
}
