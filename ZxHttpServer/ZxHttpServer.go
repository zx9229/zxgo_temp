package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type ZxHttpServer struct {
	httpServer *http.Server
}

func New_ZxHttpServer(listenAddr string) *ZxHttpServer {
	var newData *ZxHttpServer = new(ZxHttpServer)
	newData.httpServer = new(http.Server)
	newData.httpServer.Addr = listenAddr
	newData.httpServer.Handler = http.NewServeMux()
	return newData
}

func (self *ZxHttpServer) GetHttpServeMux() *http.ServeMux {
	return self.httpServer.Handler.(*http.ServeMux)
}

func (self *ZxHttpServer) Init() {
	self.GetHttpServeMux().HandleFunc("/", self.test_Root_http)
	self.GetHttpServeMux().Handle("/websocket", websocket.Handler(self.test_Root_websocket))
}

func (self *ZxHttpServer) Run() {
	self.httpServer.ListenAndServe()
}

func (self *ZxHttpServer) test_Root_http(http.ResponseWriter, *http.Request) {
	fmt.Println("test_Root_http")
}

func (self *ZxHttpServer) test_Root_websocket(ws *websocket.Conn) {
	fmt.Println(ws)
	var err error = nil
	var recvRawMessage []byte = nil

	defer func() {
		if err = ws.Close(); err != nil {
			log.Println(fmt.Sprintf("ws=%p,调用Close失败,err=%v", ws, err))
		}
	}()

	log.Println(fmt.Sprintf("ws=%p,RemoteAddr=%v", ws, ws.Request().RemoteAddr))

	for {
		recvRawMessage = nil
		if err = websocket.Message.Receive(ws, &recvRawMessage); err != nil {
			log.Println(fmt.Printf("ws=%p,调用Receive失败,err=%v", ws, err))
			return
		}

		var sendMessage string = "数据无法识别!"
		message := string(recvRawMessage)
		log.Println(message)

		sendMessage = fmt.Sprintf("我收到了[%v]", message)

		if err = websocket.Message.Send(ws, sendMessage); err != nil {
			log.Println(fmt.Sprintf("ws=%p,调用Send失败,err=%v", ws, err))
		}
	}
}

func main() {
	var port int = 8080
	listenAddr := fmt.Sprintf("localhost:%d", port)
	zxWebServer := New_ZxHttpServer(listenAddr)
	zxWebServer.Init()
	zxWebServer.Run()
	fmt.Println(time.Now())
}
