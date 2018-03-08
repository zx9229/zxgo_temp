//golang（5）：编写WebSocket服务，客户端和html5调用
//http://blog.csdn.net/freewebsys/article/details/46882777
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/net/websocket"
)

func NowStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func HandleRead(ws *websocket.Conn) {
	var err error = nil
	var recvRawMessage []byte = nil

	defer func() {
		if err = ws.Close(); err != nil {
			log.Println(fmt.Sprintf("%v,ws=%p,调用Close失败,err=%v", NowStr(), ws, err))
		}
	}()

	for {
		recvRawMessage = nil
		if err = websocket.Message.Receive(ws, &recvRawMessage); err != nil {
			log.Println(fmt.Sprintf("%v,ws=%p,调用Receive失败,err=%v", NowStr(), ws, err))
			return
		} else {
			message := string(recvRawMessage)
			log.Println(fmt.Sprintf("%v,%v", NowStr(), message))
		}
	}
}

func main() {
	var port int = 8080
	var url string = fmt.Sprintf("ws://localhost:%d/websocket", port)
	var origin string = fmt.Sprintf("https://localhost:%d/", port)

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Println(fmt.Sprintf("%v,调用Dial出错,err=%v", NowStr(), err.Error()))
		return
	}

	go HandleRead(ws)

	br := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("登录:  login|用户名|密码")
		fmt.Println("聊天:   chat|用户名|内容")
		fmt.Printf("(%v)请输入命令:\n", NowStr())

		line, isPrefix, err := br.ReadLine()
		if isPrefix || err != nil {
			log.Println(fmt.Sprintf("%v,调用ReadLine异常,isPrefix=%v,err=%v", NowStr(), isPrefix, err))
		} else {
			inputData := string(line)
			if inputData == "quit" || inputData == "exit" {
				break
			} else if inputData == "" {
				//nothing.
			} else {
				if err = websocket.Message.Send(ws, inputData); err != nil {
					log.Println(fmt.Sprintf("%v,调用Send失败,err=%v", NowStr(), err))
				}
			}
		}
	}
	return
}
