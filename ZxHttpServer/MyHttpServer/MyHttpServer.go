package MyHttpServer

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	"golang.org/x/net/websocket"
)

// [DataParser]的回调函数.
type DataParserHandler func(ws *websocket.Conn, i interface{})

// 数据解析器.
type DataParser interface {
	// 往[DataParser]里面注册回调函数,[bool]表示注册成功了吗.
	RegisterHandler(curType reflect.Type, curFun DataParserHandler) bool
	// 解析[jsonByte]并执行回调函数.[objData]转换成的对象,[cbOk]回调函数执行成功了吗,[err]具体的错误信息.
	ParseByteSlice(ws *websocket.Conn, jsonByte []byte) (objData interface{}, cbOk bool, err error)
}

// 业务处理器.
type DataBusiness interface {
	// websocket连接成功时的回调.
	Handle_WebSocket_Connected(ws *websocket.Conn)
	// websocket连接断开时的回调.
	Handle_WebSocket_Disconnected(ws *websocket.Conn)
	// 执行[websocket.Message.Receive]接收到消息时的回调,[bytes]是接收到的消息.
	Handle_WebSocket_Receive(ws *websocket.Conn, bytes []byte)
	// 执行[*websocket.Conn]的函数失败时的回调,[operation]是执行的函数名,[err]是具体的错误信息.
	Handle_WebSocket_Operation_Error(ws *websocket.Conn, operation string, err error)
	// 执行[DataParser.ParseByteSlice]失败时的回调.
	Handle_Parse_Fail(ws *websocket.Conn, bytes []byte, obj interface{}, cbOk bool, err error)
	// 要注册到[DataParser]的所有的函数,由此函数返回出来,用于后续的注册操作.
	GetRegisterHandlerMap() map[reflect.Type]DataParserHandler
}

type MyHttpServer struct {
	httpServer *http.Server
	parser     DataParser
	business   DataBusiness
}

func New_MyHttpServer(listenAddr string, parserObj DataParser, businessObj DataBusiness) *MyHttpServer {
	var newData *MyHttpServer = new(MyHttpServer)
	//
	newData.httpServer = new(http.Server)
	newData.httpServer.Addr = listenAddr
	newData.httpServer.Handler = http.NewServeMux()
	//
	newData.parser = parserObj
	//
	newData.business = businessObj
	//
	return newData
}

func (self *MyHttpServer) GetHttpServeMux() *http.ServeMux {
	return self.httpServer.Handler.(*http.ServeMux)
}

func (self *MyHttpServer) Init() {
	self.GetHttpServeMux().HandleFunc("/", self.Root_http_handler)
	self.GetHttpServeMux().Handle("/websocket", websocket.Handler(self.Root_websocket_handler))
	//
	for curType, curFun := range self.business.GetRegisterHandlerMap() {
		if self.parser.RegisterHandler(curType, curFun) == false {
			panic(fmt.Sprintf("注册函数失败,%v,%v", curType, curFun))
		}
	}
	//
}

func (self *MyHttpServer) Run() {
	self.httpServer.ListenAndServe()
}

func (self *MyHttpServer) Root_http_handler(http.ResponseWriter, *http.Request) {
	fmt.Println("test_Root_http")
}

func (self *MyHttpServer) Root_websocket_handler(ws *websocket.Conn) {
	var err error = nil
	var recvRawMessage []byte = nil

	defer func() {
		self.business.Handle_WebSocket_Disconnected(ws)
		if err = ws.Close(); err != nil {
			log.Println(fmt.Sprintf("ws=%p,调用Close失败,err=%v", ws, err))
		}
	}()
	self.business.Handle_WebSocket_Connected(ws)

	for {
		recvRawMessage = nil
		if err = websocket.Message.Receive(ws, &recvRawMessage); err != nil {
			self.business.Handle_WebSocket_Operation_Error(ws, "Receive", err)
			return
		}
		self.business.Handle_WebSocket_Receive(ws, recvRawMessage)

		//如果解析成功,会调用(TmpData)里面注册的对应的回调函数.
		if obj, cbOk, err2 := self.parser.ParseByteSlice(ws, recvRawMessage); err2 != nil {
			err = err2
			self.business.Handle_Parse_Fail(ws, recvRawMessage, obj, cbOk, err2)
		}
	}
}
