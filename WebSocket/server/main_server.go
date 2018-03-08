//golang websocket的例子
//https://studygolang.com/articles/3392
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

func NowStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

type BusinessData struct {
	mutex         *sync.Mutex
	activeClients map[*websocket.Conn]string //登录表
	userPwds      map[string]string          //用户名密码表
	sep           string                     //字段分隔符
}

func NewBusinessData() *BusinessData {
	user_passwords := map[string]string{"a1": "a1pwd", "b2": "b2pwd", "c3": "c3pwd"}
	return &BusinessData{new(sync.Mutex), map[*websocket.Conn]string{}, user_passwords, "|"}
}

func (self *BusinessData) Sep() string {
	return self.sep
}

func (self *BusinessData) Count() int {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return len(self.activeClients)
}

func (self *BusinessData) Enter(ws *websocket.Conn) {
	{ //花括号里面的defer是不能触发的.
		self.mutex.Lock()
		self.activeClients[ws] = ""
		self.mutex.Unlock()
	}
	log.Println(fmt.Sprintf("%v,ws=%p,连接成功,当前共有%d个连接", NowStr(), ws, self.Count()))
}

func (self *BusinessData) Exit(ws *websocket.Conn) {
	{
		self.mutex.Lock()
		delete(self.activeClients, ws)
		self.mutex.Unlock()
	}
	log.Println(fmt.Sprintf("%v,ws=%p,断开成功,当前还有%d个连接", NowStr(), ws, self.Count()))
}

func (self *BusinessData) Login(ws *websocket.Conn, message string) string {
	//login|用户名|密码
	fieldSlice := strings.Split(message, self.sep)
	if len(fieldSlice) != 3 || fieldSlice[0] != "login" {
		return "协议错误!"
	}

	self.mutex.Lock()
	defer self.mutex.Unlock()

	if userName, ok := self.activeClients[ws]; !ok || userName != "" {
		return "此连接已经登录,无法再次登录!"
	}

	if pwd, ok := self.userPwds[fieldSlice[1]]; !ok || pwd != fieldSlice[2] {
		return "用户名或密码错误!"
	}

	self.activeClients[ws] = fieldSlice[1]
	return "登录成功."
}

func (self *BusinessData) Chat(ws *websocket.Conn, message string) string {
	//Chat|要发给哪个用户|内容
	fieldSlice := strings.Split(message, self.sep)
	if len(fieldSlice) != 3 || fieldSlice[0] != "chat" {
		return "协议错误!"
	}

	self.mutex.Lock()
	defer self.mutex.Unlock()

	var succcessCnt int = 0
	var failureCnt int = 0
	for cs, user := range self.activeClients {
		if fieldSlice[1] == user {
			if err := websocket.Message.Send(cs, fieldSlice[2]); err != nil {
				log.Println(fmt.Sprintf("%v,cs=%p,调用Send失败,err=%v", NowStr(), cs, err))
				failureCnt += 1
			} else {
				succcessCnt += 1
			}
		}
	}
	return fmt.Sprintf("聊天结束,成功%v个,失败%v个.", succcessCnt, failureCnt)
}

var GloablBusiness *BusinessData = NewBusinessData()

func Root(http.ResponseWriter, *http.Request) {

}

func Root_websocket(ws *websocket.Conn) {
	var err error = nil
	var recvRawMessage []byte = nil

	defer func() {
		GloablBusiness.Exit(ws)
		if err = ws.Close(); err != nil {
			log.Println(fmt.Sprintf("%v,ws=%p,调用Close失败,err=%v", NowStr(), ws, err))
		}
	}()
	GloablBusiness.Enter(ws)

	log.Println(fmt.Sprintf("%v,ws=%p,RemoteAddr=%v", NowStr(), ws, ws.Request().RemoteAddr))

	for {
		recvRawMessage = nil
		if err = websocket.Message.Receive(ws, &recvRawMessage); err != nil {
			log.Println(fmt.Printf("%v,ws=%p,调用Receive失败,err=%v", NowStr(), ws, err))
			return
		}

		var sendMessage string = "数据无法识别!"
		message := string(recvRawMessage)
		if strings.HasPrefix(message, "login"+GloablBusiness.Sep()) {
			sendMessage = GloablBusiness.Login(ws, message)
		} else if strings.HasPrefix(message, "chat"+GloablBusiness.Sep()) {
			sendMessage = GloablBusiness.Chat(ws, message)
		}

		if err = websocket.Message.Send(ws, sendMessage); err != nil {
			log.Println(fmt.Sprintf("%v,ws=%p,调用Send失败,err=%v", NowStr(), ws, err))
		}
	}
}

func main() {
	var port int = 8080
	listenAddr := fmt.Sprintf("localhost:%d", port)

	myDefaultServeMux := http.NewServeMux()
	myDefaultServeMux.HandleFunc("/", Root)
	myDefaultServeMux.Handle("/websocket", websocket.Handler(Root_websocket))

	httpServer := &http.Server{Addr: listenAddr, Handler: myDefaultServeMux}

	log.Println(fmt.Sprintf("%v,即将在瞬间开启服务...", NowStr()))
	var err error = nil
	if true {
		err = httpServer.ListenAndServe()
	} else {
		//go run C:\go\src\crypto\tls\generate_cert.go --host localhost
		certFile := "cert.pem"
		keyFile := "key.pem"
		err = httpServer.ListenAndServeTLS(certFile, keyFile)
	}
	log.Println(err)
	log.Println(fmt.Sprintf("%v,程序即将退出...", NowStr()))
}
