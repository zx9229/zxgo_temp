package BusinessCenter

import (
	"fmt"
	"log"
	"reflect"

	"github.com/zx9229/zxgo_temp/ZxHttpServer/TxStruct"
	"golang.org/x/net/websocket"
)

type DataCenter struct {
	xyz map[*websocket.Conn]string
}

func New_DataCenter() *DataCenter {
	newData := new(DataCenter)
	//
	newData.xyz = nil
	//
	return newData
}

func (self *DataCenter) Handle_websocket_Connected(ws *websocket.Conn) {
	log.Println(fmt.Sprintf("收到连接:ws=%p,RemoteAddr=%v", ws, ws.Request().RemoteAddr))
}

func (self *DataCenter) Handle_websocket_Disconnected(ws *websocket.Conn) {
	log.Println(fmt.Sprintf("断开连接:ws=%p", ws))
}

func (self *DataCenter) Handle_websocket_Receive(ws *websocket.Conn, bytes []byte) {
	//log.Println(fmt.Sprintf("收到消息:ws=%p,%v", ws, string(bytes)))
}

func (self *DataCenter) Handle_websocket_Operation_Error(ws *websocket.Conn, operation string, err error) {
	log.Println(fmt.Sprintf("操作失败:ws=%p,%v=>%v", ws, operation, err))
}

func (self *DataCenter) Handle_Parse_Fail(ws *websocket.Conn, bytes []byte, obj interface{}, cbOk bool, err error) {
	log.Println(fmt.Sprintf("解析失败:ws=%p,%v,%v", ws, string(bytes), err))
	if true {
		var sendMessage string = "数据处理失败!"
		if err = websocket.Message.Send(ws, sendMessage); err != nil {
			self.Handle_websocket_Operation_Error(ws, "Send", err)
		}
	}
}

func (self *DataCenter) Handle_Parse_OK_ChatMessage(ws *websocket.Conn, objData interface{}) {
	log.Println(fmt.Sprintf("解析成功:ws=%p,%v", ws, objData))
	if true {
		var sendMessage string = "解析数据成功!"
		if err := websocket.Message.Send(ws, sendMessage); err != nil {
			self.Handle_websocket_Operation_Error(ws, "Send", err)
		}
	}
}

func (self *DataCenter) GetRegisterMap() map[reflect.Type]TxStruct.Handler {
	mapData := make(map[reflect.Type]TxStruct.Handler)
	//
	mapData[reflect.ValueOf(TxStruct.ChatMessage{}).Type()] = self.Handle_Parse_OK_ChatMessage
	//
	return mapData
}
