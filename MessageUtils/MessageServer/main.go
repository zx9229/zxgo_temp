package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/zx9229/zxgo_temp/MessageUtils/MessageServer/SimpleHttpServer"
)

type ConfigData struct {
	Host              string
	Port              int
	DB_DriverName     string
	DB_DataSourceName string
	DB_LocationName   string
}

func main() {
	var err error

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

	dataCenter := New_DataCenter()
	if err = dataCenter.Init(cfgData.DB_DriverName, cfgData.DB_DataSourceName, cfgData.DB_LocationName); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if dataCenter.infoSlice, err = dataCenter.calcCacheData(); err != nil {
		panic(err)
	}

	listenAddr := fmt.Sprintf("%s:%d", cfgData.Host, cfgData.Port)
	simpleHttpServer := SimpleHttpServer.New_SimpleHttpServer(listenAddr)
	simpleHttpServer.GetHttpServeMux().HandleFunc("/ReportReq", dataCenter.Handler_ReportReq)
	simpleHttpServer.GetHttpServeMux().HandleFunc("/AddAgentReq", dataCenter.Handler_AddAgentReq)
	err = simpleHttpServer.Run()
	log.Println(err)
}
