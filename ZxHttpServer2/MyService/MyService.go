package MyService

import (
	"net/http"

	"github.com/zx9229/zxgo_temp/ZxHttpServer2/CacheOnline"
	"github.com/zx9229/zxgo_temp/ZxHttpServer2/SimpleHttpServer"

	"golang.org/x/net/websocket"
)

type MyService struct {
	httpServer *SimpleHttpServer.SimpleHttpServer
	xxx        *CacheOnline.ConnectionManager
}

func New_MyService(listenAddr string) *MyService {
	curData := new(MyService)
	curData.httpServer = SimpleHttpServer.New_SimpleHttpServer(listenAddr)
	curData.xxx = CacheOnline.New_ConnectionManager()
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
	CacheOnline.New_TxConnection(ws, nil, nil, self.xxx)
}
