package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
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
	newData.httpServer = new(http.Server)
	newData.httpServer.Addr = listenAddr
	newData.httpServer.Handler = http.NewServeMux()
	//
	newData.parser = TxStruct.New_TxParser()
	//
	newData.business = BusinessCenter.New_DataCenter()
	//
	newData.parser.RegisterHandler(reflect.ValueOf(TxStruct.ChatMessage{}).Type(), newData.business.Handle_Parse_OK_ChatMessage)
	return newData
}

func (self *ZxHttpServer) GetHttpServeMux() *http.ServeMux {
	return self.httpServer.Handler.(*http.ServeMux)
}

func (self *ZxHttpServer) Init() {
	self.GetHttpServeMux().HandleFunc("/", self.test_Root_http)
	self.GetHttpServeMux().Handle("/websocket", websocket.Handler(self.test_Root_websocket))
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
		self.business.Handle_websocket_Close(ws)
		if err = ws.Close(); err != nil {
			log.Println(fmt.Sprintf("ws=%p,调用Close失败,err=%v", ws, err))
		}
	}()
	self.business.Handle_websocket_Open(ws)

	for {
		recvRawMessage = nil
		if err = websocket.Message.Receive(ws, &recvRawMessage); err != nil {
			self.business.Handle_websocket_Operate_Fail(ws, "Receive", err)
			return
		}
		self.business.Handle_websocket_Receive(ws, recvRawMessage)

		//如果解析成功,会调用(TmpData)里面注册的对应的回调函数.
		if _, _, err = self.parser.ParseByteSlice(ws, recvRawMessage); err != nil {
			log.Println(err) //TODO:
			if false {
				var sendMessage string = "数据处理失败!"
				if err = websocket.Message.Send(ws, sendMessage); err != nil {
					log.Println(fmt.Sprintf("ws=%p,调用Send失败,err=%v", ws, err))
				}
			}
		}
	}
}

func main() {
	//ormData := TxStruct.New_OrmData()
	xx := new(TxStruct.ChatMessage)
	bb, _ := json.Marshal(xx)
	fmt.Println(string(bb))

	var port int = 8080
	listenAddr := fmt.Sprintf("localhost:%d", port)
	zxWebServer := New_ZxHttpServer(listenAddr)
	zxWebServer.Init()
	zxWebServer.Run()
	fmt.Println(time.Now())
}
