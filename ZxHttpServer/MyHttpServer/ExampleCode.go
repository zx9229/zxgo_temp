package MyHttpServer

import (
	"errors"
	"fmt"
	"log"
	"reflect"

	"golang.org/x/net/websocket"
)

var (
	// ErrNotImplemented not implemented
	ErrNotImplemented = errors.New("Function not implemented")
)

type DefaultDataParser struct {
}

func (self *DefaultDataParser) RegisterHandler(curType reflect.Type, curFun DataParserHandler) bool {
	return false
}

func (self *DefaultDataParser) ParseByteSlice(ws *websocket.Conn, jsonByte []byte) (objData interface{}, cbOk bool, err error) {
	objData = nil
	cbOk = false
	err = ErrNotImplemented
	return
}

type DefaultDataBusiness struct {
}

func (self *DefaultDataBusiness) Handle_WebSocket_Connected(ws *websocket.Conn) {
	log.Println(fmt.Sprintf("收到连接:ws=[%p],RemoteAddr=%v", ws, ws.Request().RemoteAddr))
}

func (self *DefaultDataBusiness) Handle_WebSocket_Disconnected(ws *websocket.Conn) {
	log.Println(fmt.Sprintf("断开连接:ws=[%p]", ws))
}

func (self *DefaultDataBusiness) Handle_WebSocket_Receive(ws *websocket.Conn, bytes []byte) {
	if false {
		log.Println(fmt.Sprintf("收到消息:ws=[%p],%v", ws, string(bytes)))
	}
}

func (self *DefaultDataBusiness) Handle_WebSocket_Operation_Error(ws *websocket.Conn, operation string, err error) {
	log.Println(fmt.Sprintf("操作失败:ws=[%p],%v=>%v", ws, operation, err))
}

func (self *DefaultDataBusiness) Handle_Parse_Fail(ws *websocket.Conn, bytes []byte, obj interface{}, cbOk bool, err error) {
	log.Println(fmt.Sprintf("解析失败:ws=[%p],%v,%v", ws, string(bytes), err))

	var sendMessage string = "数据处理失败!"
	if err = websocket.Message.Send(ws, sendMessage); err != nil {
		self.Handle_WebSocket_Operation_Error(ws, "Send", err)
	}
}

func (self *DefaultDataBusiness) GetRegisterHandlerMap() map[reflect.Type]DataParserHandler {
	return make(map[reflect.Type]DataParserHandler)
}

func ThisIsAnExample() {
	var port int = 8080
	listenAddr := fmt.Sprintf("localhost:%d", port)
	myWebServer := New_MyHttpServer(listenAddr, &DefaultDataParser{}, &DefaultDataBusiness{})
	myWebServer.Init()
	myWebServer.Run()
	fmt.Println("will exit...")
}
