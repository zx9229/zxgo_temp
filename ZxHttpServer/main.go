package main

import (
	"fmt"

	"github.com/zx9229/zxgo_temp/ZxHttpServer/BusinessCenter"
	"github.com/zx9229/zxgo_temp/ZxHttpServer/MyHttpServer"
	"github.com/zx9229/zxgo_temp/ZxHttpServer/TxStruct"
)

func main() {
	var port int = 8080
	listenAddr := fmt.Sprintf("localhost:%d", port)
	myWebServer := MyHttpServer.New_MyHttpServer(listenAddr, TxStruct.New_TxParser(), BusinessCenter.New_DataCenter())
	myWebServer.Init()
	myWebServer.Run()
	fmt.Println("will exit...")
}
