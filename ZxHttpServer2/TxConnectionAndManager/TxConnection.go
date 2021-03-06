package TxConnectionAndManager

import (
	"fmt"
	"io"
	"log"
	"reflect"

	"github.com/zx9229/zxgo_temp/ZxHttpServer2/ChatStruct"
	"github.com/zx9229/zxgo_temp/ZxHttpServer2/TxStruct"
	"golang.org/x/net/websocket"
)

// 连接(Connection)和连接管理器(ConnectionManager)不太可能独立开来. 因为:
// 很显然, 管理器要维护各个连接, 所以避免不了的,              管理器要存储连接的指针, 调用某连接的函数.
// 如果一个连接断开了, 它要通知管理器, 让管理器删除自己, 所以, 连接要存储管理器的指针, 调用管理器的函数.
// (如果不使用std::function等回调函数绑定) 在C++里, 它俩是要搞成友元类的, 否则搞不定这个事情.
// 在golang里面与C++里面类似.

type TxConnection struct {
	ws         *websocket.Conn
	handles    map[reflect.Type]func(i interface{})
	mngr       *TxConnectionManager
	DeviceType int //(登录时,使用的)设备类型(手机/PC/网页/等).
	UD         *ChatStruct.UserData
}

func New_TxConnection(ws *websocket.Conn, manager *TxConnectionManager) *TxConnection {
	curData := new(TxConnection)
	//
	curData.ws = ws
	curData.handles = curData.CalcHandlerMap()
	curData.mngr = manager
	curData.DeviceType = 0
	curData.UD = nil
	//
	manager.HandleConnected(curData)
	go curData.Handler_websocket()
	//
	return curData
}

func (self *TxConnection) Handler_websocket() {
	var err error = nil
	var recvRawMessage []byte = nil
	var objData interface{} = nil
	var objType reflect.Type = nil
	var handler func(i interface{}) = nil
	var isOk bool = false

	defer func() {
		self.Handle_WebSocket_Disconnected()
		if err = self.ws.Close(); err != nil {
			log.Println(fmt.Sprintf("ws=%p,调用Close失败,err=%v", self.ws, err))
		}
	}()
	self.Handle_WebSocket_Connected()

	for {
		recvRawMessage = nil
		if err = self._websocket_Message_Receive(&recvRawMessage); err != nil {
			return
		}
		self.Handle_WebSocket_Receive(recvRawMessage)

		if objData, objType, err = self.mngr.parser.ParseByteSlice(recvRawMessage); err != nil {
			self.Handle_Parse_Fail(recvRawMessage, err)
			continue
		}

		if handler, isOk = self.handles[objType]; !isOk {
			log.Println(fmt.Sprintf("ws=%p,找不到对应的处理函数", self.ws))
			continue
		}

		handler(objData)
	}
}

func (self *TxConnection) CalcHandlerMap() map[reflect.Type]func(i interface{}) {
	mapData := make(map[reflect.Type]func(i interface{}))
	//
	mapData[reflect.ValueOf(TxStruct.LoginReq{}).Type()] = self.Handle_Parse_OK_LoginReq
	//
	return mapData
}

func (self *TxConnection) Send_Temp(v interface{}) {
	self._websocket_Message_Send(v)
}

func (self *TxConnection) _websocket_Message_Send(v interface{}) {
	if err := websocket.Message.Send(self.ws, v); err != nil {
		self.Handle_WebSocket_Operation_Error("Send", err)
	}
}

func (self *TxConnection) _websocket_Message_Receive(v interface{}) error {
	var err error
	if err = websocket.Message.Receive(self.ws, v); err != nil {
		if err != io.EOF {
			self.Handle_WebSocket_Operation_Error("Receive", err)
		}
	}
	return err
}

func (self *TxConnection) Handle_WebSocket_Connected() {
	log.Println(fmt.Sprintf("收到连接:ws=[%p],RemoteAddr=%v", self.ws, self.ws.Request().RemoteAddr))
}

func (self *TxConnection) Handle_WebSocket_Disconnected() {
	self.mngr.HandleDisconnected(self)
	log.Println(fmt.Sprintf("断开连接:ws=[%p]", self.ws))
}

func (self *TxConnection) Handle_WebSocket_Receive(bytes []byte) {
	//log.Println(fmt.Sprintf("收到消息:ws=[%p],%v", ws, string(bytes)))
}

func (self *TxConnection) Handle_WebSocket_Operation_Error(operation string, err error) {
	log.Println(fmt.Sprintf("操作失败:ws=[%p],%v=>%v", self.ws, operation, err))
}

func (self *TxConnection) Handle_Parse_Fail(bytes []byte, err error) {
	log.Println(fmt.Sprintf("解析失败:ws=[%p],%v,%v", self.ws, string(bytes), err))
	if true {
		var sendMessage string = "数据处理失败!"
		self._websocket_Message_Send(sendMessage)
	}
}

func (self *TxConnection) Handle_Parse_OK_LoginReq(v interface{}) {
	reqObj := v.(*TxStruct.LoginReq)
	rspObj := new(TxStruct.LoginRsp)
	rspObj.FillField_FromReq(reqObj)
	//var err error
	//if err = self.chatRoom.LoginReq(ws, reqObj.UserId, reqObj.UserAlias, reqObj.Password); err != nil {
	//	rspObj.Code = -1
	//	rspObj.Message = err.Error()
	//} else {
	//	rspObj.Code = 0
	//	rspObj.Message = "登录成功"
	//}

	for _ = range "1" {
		if self.UD != nil {
			break
		}
		//TODO:
		self.mngr.HandleLogin(self)
	}

	self._websocket_Message_Send(TxStruct.ToJsonStr(rspObj))
}
