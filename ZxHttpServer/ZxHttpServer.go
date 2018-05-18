package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/zx9229/zxgo_temp/ZxHttpServer/BusinessCenter"
	"github.com/zx9229/zxgo_temp/ZxHttpServer/TxStruct"
	"golang.org/x/net/websocket"
)

type ZxHttpServer struct {
	httpServer *http.Server
	parser     *TxStruct.TxParser
	business   *BusinessCenter.DataCenter
}

func New_ZxHttpServer(listenAddr string) *ZxHttpServer {
	var newData *ZxHttpServer = new(ZxHttpServer)
	//
	newData.httpServer = new(http.Server)
	newData.httpServer.Addr = listenAddr
	newData.httpServer.Handler = http.NewServeMux()
	//
	newData.parser = TxStruct.New_TxParser()
	//
	newData.business = BusinessCenter.New_DataCenter()
	//
	return newData
}

func (self *ZxHttpServer) GetHttpServeMux() *http.ServeMux {
	return self.httpServer.Handler.(*http.ServeMux)
}

func (self *ZxHttpServer) Init() {
	self.GetHttpServeMux().HandleFunc("/", self.test_Root_http)
	self.GetHttpServeMux().Handle("/websocket", websocket.Handler(self.test_Root_websocket))
	//
	for k, v := range self.business.GetRegisterMap() {
		if self.parser.RegisterHandler(k, v) == false {
			panic(fmt.Sprintf("注册函数失败,%v,%v", k, v))
		}
	}
	//
}

func (self *ZxHttpServer) Run() {
	self.httpServer.ListenAndServe()
}

func (self *ZxHttpServer) test_Root_http(http.ResponseWriter, *http.Request) {
	fmt.Println("test_Root_http")
}

func (self *ZxHttpServer) test_Root_websocket(ws *websocket.Conn) {
	var err error = nil
	var recvRawMessage []byte = nil

	defer func() {
		self.business.Handle_websocket_Disconnected(ws)
		if err = ws.Close(); err != nil {
			log.Println(fmt.Sprintf("ws=%p,调用Close失败,err=%v", ws, err))
		}
	}()
	self.business.Handle_websocket_Connected(ws)

	for {
		recvRawMessage = nil
		if err = websocket.Message.Receive(ws, &recvRawMessage); err != nil {
			self.business.Handle_websocket_Operation_Error(ws, "Receive", err)
			return
		}
		self.business.Handle_websocket_Receive(ws, recvRawMessage)

		//如果解析成功,会调用(TmpData)里面注册的对应的回调函数.
		if obj, cbOk, err2 := self.parser.ParseByteSlice(ws, recvRawMessage); err2 != nil {
			err = err2
			self.business.Handle_Parse_Fail(ws, recvRawMessage, obj, cbOk, err2)
		}
	}
}

func main() {
	objData := new(TxStruct.ChatMessage)
	objData.FillField_Type()
	bb, _ := json.Marshal(objData)
	fmt.Println(string(bb))

	var port int = 8080
	listenAddr := fmt.Sprintf("localhost:%d", port)
	zxWebServer := New_ZxHttpServer(listenAddr)
	zxWebServer.Init()
	zxWebServer.Run()
	fmt.Println(time.Now())
}
