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

	var err error
	if true {
		err = myWebServer.Run()
	} else {
		//go run C:\go\src\crypto\tls\generate_cert.go --host localhost
		certFile := "cert.pem"
		keyFile := "key.pem"
		err = myWebServer.RunTLS(certFile, keyFile)
	}
	fmt.Println(err)
	fmt.Println("will exit...")
}
