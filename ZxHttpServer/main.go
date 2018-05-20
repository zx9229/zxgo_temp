package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/zx9229/zxgo_temp/ZxHttpServer/MyService"
)

type ConfigData struct {
	Host              string
	Port              int
	DB_DriverName     string
	DB_DataSourceName string
	DB_LocationName   string
}

func main() {
	var cfgData ConfigData = ConfigData{}
	var config_filename string = "./config.json"
	if content, err := ioutil.ReadFile(config_filename); err != nil && err != io.EOF {
		log.Println(fmt.Sprintf("读取配置文件出错: %v", err))
		os.Exit(1)
	} else {
		if err := json.Unmarshal(content, &cfgData); err != nil {
			log.Println(fmt.Sprintf("解析配置文件出错: %v", err))
			os.Exit(1)
		}
	}

	listenAddr := fmt.Sprintf("%s:%d", cfgData.Host, cfgData.Port)
	myWebService := MyService.New_MyService(listenAddr, cfgData.DB_DriverName, cfgData.DB_DataSourceName, cfgData.DB_LocationName)
	myWebService.Init()

	var err error
	if true {
		err = myWebService.Run()
	} else {
		//go run C:\go\src\crypto\tls\generate_cert.go --host localhost
		certFile := "cert.pem"
		keyFile := "key.pem"
		err = myWebService.RunTLS(certFile, keyFile)
	}
	log.Println(err)
	log.Println("will exit...")
}
