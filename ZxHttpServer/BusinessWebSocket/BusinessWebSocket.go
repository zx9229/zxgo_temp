package BusinessWebSocket

import (
	"fmt"
	"io"
	"log"
	"reflect"

	"github.com/zx9229/zxgo_temp/ZxHttpServer/ChatRoom"
	"github.com/zx9229/zxgo_temp/ZxHttpServer/TxStruct"
	"golang.org/x/net/websocket"
)

type BusinessWebSocket struct {
	parser   *TxStruct.TxParser
	chatRoom *ChatRoom.ChatRoom
}

func New_BusinessWebSocket(driverName, dataSourceName, locationName string) *BusinessWebSocket {
	curData := new(BusinessWebSocket)
	//
	curData.parser = TxStruct.New_TxParser()
	for curType, curFun := range curData.GetRegisterHandlerMap() {
		if curData.parser.RegisterHandler(curType, curFun) == false {
			panic(fmt.Sprintf("注册回调函数失败, %v, %v", curType, curFun))
		}
	}
	//
	curData.chatRoom = ChatRoom.New_ChatRoom()
	curData.chatRoom.Init(driverName, dataSourceName, locationName)
	return curData
}

func (self *BusinessWebSocket) _websocket_Message_Send(ws *websocket.Conn, v interface{}) {
	if err := websocket.Message.Send(ws, v); err != nil {
		self.Handle_WebSocket_Operation_Error(ws, "Send", err)
	}
}

func (self *BusinessWebSocket) _websocket_Message_Receive(ws *websocket.Conn, v interface{}) error {
	var err error
	if err = websocket.Message.Receive(ws, v); err != nil {
		if err != io.EOF {
			self.Handle_WebSocket_Operation_Error(ws, "Receive", err)
		}
	}
	return err
}

func (self *BusinessWebSocket) Handler_websocket(ws *websocket.Conn) {
	var err error = nil
	var recvRawMessage []byte = nil

	defer func() {
		self.Handle_WebSocket_Disconnected(ws)
		if err = ws.Close(); err != nil {
			log.Println(fmt.Sprintf("ws=%p,调用Close失败,err=%v", ws, err))
		}
	}()
	self.Handle_WebSocket_Connected(ws)

	for {
		recvRawMessage = nil
		if err = self._websocket_Message_Receive(ws, &recvRawMessage); err != nil {
			return
		}
		self.Handle_WebSocket_Receive(ws, recvRawMessage)

		//如果解析成功,会调用(self.parser)里面注册的对应的回调函数.
		if obj, cbOk, err2 := self.parser.ParseByteSlice(ws, recvRawMessage); err2 != nil {
			err = err2
			self.Handle_Parse_Fail(ws, recvRawMessage, obj, cbOk, err2)
		}
	}
}

func (self *BusinessWebSocket) Handle_WebSocket_Connected(ws *websocket.Conn) {
	log.Println(fmt.Sprintf("收到连接:ws=[%p],RemoteAddr=%v", ws, ws.Request().RemoteAddr))
}

func (self *BusinessWebSocket) Handle_WebSocket_Disconnected(ws *websocket.Conn) {
	log.Println(fmt.Sprintf("断开连接:ws=[%p]", ws))
}

func (self *BusinessWebSocket) Handle_WebSocket_Receive(ws *websocket.Conn, bytes []byte) {
	//log.Println(fmt.Sprintf("收到消息:ws=[%p],%v", ws, string(bytes)))
}

func (self *BusinessWebSocket) Handle_WebSocket_Operation_Error(ws *websocket.Conn, operation string, err error) {
	log.Println(fmt.Sprintf("操作失败:ws=[%p],%v=>%v", ws, operation, err))
}

func (self *BusinessWebSocket) Handle_Parse_Fail(ws *websocket.Conn, bytes []byte, obj interface{}, cbOk bool, err error) {
	log.Println(fmt.Sprintf("解析失败:ws=[%p],%v,%v", ws, string(bytes), err))
	if true {
		var sendMessage string = "数据处理失败!"
		self._websocket_Message_Send(ws, sendMessage)
	}
}

func (self *BusinessWebSocket) Handle_Parse_OK_ChatMessage(ws *websocket.Conn, objData interface{}) {
	log.Println(fmt.Sprintf("解析成功:ws=[%p],%v", ws, objData))
	if true {
		var sendMessage string = "解析数据成功!" + (objData.(*TxStruct.ChatMessage)).Type
		self._websocket_Message_Send(ws, sendMessage)
	}
}
func (self *BusinessWebSocket) Handle_Parse_OK_PushMessage(ws *websocket.Conn, objData interface{}) {
	log.Println(fmt.Sprintf("解析成功:ws=[%p],%v", ws, objData))
	if true {
		var sendMessage string = "解析数据成功!" + (objData.(*TxStruct.PushMessage)).Type
		self._websocket_Message_Send(ws, sendMessage)
	}
}

func (self *BusinessWebSocket) GetRegisterHandlerMap() map[reflect.Type]TxStruct.TxParserHandler {
	mapData := make(map[reflect.Type]TxStruct.TxParserHandler)
	//
	mapData[reflect.ValueOf(TxStruct.ChatMessage{}).Type()] = self.Handle_Parse_OK_ChatMessage
	mapData[reflect.ValueOf(TxStruct.PushMessage{}).Type()] = self.Handle_Parse_OK_PushMessage
	//
	return mapData
}
