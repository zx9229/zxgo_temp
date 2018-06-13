package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
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

	dataCenter := New_DataCenter()
	if err := dataCenter.Init(cfgData.DB_DriverName, cfgData.DB_DataSourceName, cfgData.DB_LocationName); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	listenAddr := fmt.Sprintf("%s:%d", cfgData.Host, cfgData.Port)
	simpleHttpServer := New_SimpleHttpServer(listenAddr)
	simpleHttpServer.GetHttpServeMux().HandleFunc("/", dataCenter.Handler_ROOT)
	err := simpleHttpServer.Run()
	log.Println(err)
}
