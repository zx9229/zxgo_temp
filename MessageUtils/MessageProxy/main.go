package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type ReportReq struct {
	UserId  int64
	RefId   int64     //rowId
	RefTime time.Time //rowUpdateTime
	Status  int       // (三态) 0=>正常;1=>警告;其他值=>错误
	Message string
	Group1  string
	Group2  string
	Group3  string
	Group4  string
}

type ReportRsp struct {
	UserId  int64
	RefId   int64
	Id      int64 // 0=>没有入库;正数=>写入数据库
	Code    int   // 0=>处理成功;其他值=>处理失败
	Message string
}

func main() {
	httpPostForm()
}

func httpPostForm() {
	data := ReportReq{}
	data.UserId = 1
	data.RefId = 6
	data.Message = "qwert"
	bytes, err := json.Marshal(data)
	jsonStr := string(bytes)
	r := strings.NewReader(jsonStr)

	resp, err := http.Post("http://localhost:8080", "application/json", r)

	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}
	fmt.Println("=================")
	fmt.Println(string(body))

}

func test() {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Poor soul, here is what you got: " + err.Error())
	}
	for _, inter := range interfaces {
		fmt.Println(inter.Name, inter.HardwareAddr)
	}
}
