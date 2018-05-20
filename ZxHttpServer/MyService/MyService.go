package MyService

import (
	"net/http"

	"github.com/zx9229/zxgo_temp/ZxHttpServer/BusinessHttp"
	"github.com/zx9229/zxgo_temp/ZxHttpServer/BusinessWebSocket"
	"github.com/zx9229/zxgo_temp/ZxHttpServer/SimpleHttpServer"
	"golang.org/x/net/websocket"
)

type MyService struct {
	httpServer       *SimpleHttpServer.SimpleHttpServer
	serviceHttp      *BusinessHttp.BusinessHttp
	serviceWebSocket *BusinessWebSocket.BusinessWebSocket
}

func New_MyService(listenAddr string, driverName, dataSourceName, locationName string) *MyService {
	curData := new(MyService)
	curData.httpServer = SimpleHttpServer.New_SimpleHttpServer(listenAddr)
	curData.serviceHttp = BusinessHttp.New_BusinessHttp()
	curData.serviceWebSocket = BusinessWebSocket.New_BusinessWebSocket(driverName, dataSourceName, locationName)
	return curData
}

func (self *MyService) Init() {
	self.httpServer.GetHttpServeMux().HandleFunc("/", self.serviceHttp.Handler_ROOT)
	self.httpServer.GetHttpServeMux().HandleFunc("/TxStruct", self.serviceHttp.Handler_TxStruct)
	self.httpServer.GetHttpServeMux().Handle("/websocket", websocket.Handler(self.serviceWebSocket.Handler_websocket))
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
