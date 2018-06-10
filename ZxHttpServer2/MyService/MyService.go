package MyService

import (
	"net/http"

	"github.com/zx9229/zxgo_temp/ZxHttpServer2/SimpleHttpServer"
	"github.com/zx9229/zxgo_temp/ZxHttpServer2/TxConnectionAndManager"

	"golang.org/x/net/websocket"
)

type MyService struct {
	httpServer *SimpleHttpServer.SimpleHttpServer
	xxx        *TxConnectionAndManager.TxConnectionManager
}

func New_MyService(listenAddr string) *MyService {
	curData := new(MyService)
	curData.httpServer = SimpleHttpServer.New_SimpleHttpServer(listenAddr)
	curData.xxx = TxConnectionAndManager.New_TxConnectionManager()
	return curData
}

func (self *MyService) Init() {
	self.httpServer.GetHttpServeMux().Handle("/websocket", websocket.Handler(self.Handler_websocket))
	self.httpServer.GetHttpServeMux().HandleFunc("/files/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})
}

func (self *MyService) Run() error {
	return self.httpServer.Run()
}

func (self *MyService) RunTLS(certFile string, keyFile string) error {
	return self.httpServer.RunTLS(certFile, keyFile)
}

func (self *MyService) Handler_websocket(ws *websocket.Conn) {
	TxConnectionAndManager.New_TxConnection(ws, nil, nil, self.xxx)
}
