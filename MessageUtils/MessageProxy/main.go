package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/zx9229/zxgo_temp/MessageUtils/TxStruct"
)

//服务器必须要有紧急通道,如果大量的数据一直提交不上来的话,需要通过紧急通道告诉服务器
//因为人员肯定会去看服务器里的信息,所以紧急信息可以被人员看到,从而被人们所知.
//建议: body里面写上紧急信息的内容,服务器专门有一个紧急信息表,
// CREATE TABLE tn(Id int64, Message string);  // Message里面可以是任何内容(比如json).
//这样,只要服务器能收到数据,就肯定能被存下来.

func main() {
	var err error
	if true {
		var path string
		if path, err = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
			panic(err)
		}
		if err = os.Chdir(path); err != nil {
			panic(err)
		}
		if _, err = os.Getwd(); err != nil {
			panic(err)
		}
	}

	dataProxy := new_DataProxy()
	if err = dataProxy.Init("sqlite3", "test_proxy.db", "Asia/Shanghai"); err != nil {
		panic(err)
	}

	var cfgInfo *ConfigInfo
	if cfgInfo, err = dataProxy.LoadConfigInfo(); err != nil {
		panic(err)
	}

	if cfgInfo == nil || len(cfgInfo.Host) <= 0 || !(0 < cfgInfo.Port && cfgInfo.Port < 65536) {
		cfgInfo = new(ConfigInfo)
		cfgInfo.Host, cfgInfo.Port = ReadFromStdin()
		if err = dataProxy.SaveConfigInfo(cfgInfo); err != nil {
			panic(err)
		}
	}

	if err = dataProxy.FlushExeInfo(); err != nil {
		panic(err)
	}

	for {
		dataProxy.FlushExeInfo()
		if slice_, err := dataProxy.QueryData(); err != nil {
			for _, item := range slice_ {
				for cnt := 0; xxx(&item, cnt); cnt++ {
					if 100 < cnt {
						panic(cnt)
					}
				}
			}
		}
		time.Sleep(time.Millisecond * 500)
	}
}

func xxx(reqRsp *ReportReqRsp, alreadyTryCnt int) bool {
	//返回值(bool)=>是否还需要重新处理它(true=>需要重新处理).

	var err error
	var byteSlice []byte

	url := fmt.Sprintf("http://%s:%d/ReportReq", "localhost", 8080)

	var reqData *TxStruct.ReportReq = reqRsp.ToReq()
	byteSlice, err = json.Marshal(reqData)
	if err != nil {
		reqRsp.IsHandled = 1
		reqRsp.RspId = -1
		reqRsp.RspCode = 1
		reqRsp.Message = fmt.Sprintf("[Proxy]转换成ReportReq失败,err=%v", err)
		return false
	}

	r := strings.NewReader(string(byteSlice))
	resp, err := http.Post(url, "application/json", r)
	if err != nil {
		return true
	}

	defer resp.Body.Close()
	if byteSlice, err = ioutil.ReadAll(resp.Body); err != nil {
		if 3 < alreadyTryCnt {
			reqRsp.IsHandled = 1
			reqRsp.RspId = -1
			reqRsp.RspCode = 1
			reqRsp.Message = fmt.Sprintf("[Proxy]ReadAll失败,err=%v", err)
			return false
		} else {
			return true
		}
	}

	rspData := new(TxStruct.ReportRsp)
	if err = json.Unmarshal(byteSlice, rspData); err != nil {
		reqRsp.IsHandled = 1
		reqRsp.RspId = -1
		reqRsp.RspCode = 1
		reqRsp.Message = fmt.Sprintf("[Proxy]转换成ReportRsp失败,err=%v", err)
		return false
	}

	if err = reqRsp.FillWithRsp(rspData, false); err != nil {
		reqRsp.IsHandled = 1
		reqRsp.RspId = -1
		reqRsp.RspCode = 1
		reqRsp.Message = fmt.Sprintf("[Proxy]转换成ReportRsp失败,err=%v", err)
		return false
	}

	return false
}

func ReadFromStdin() (Host string, Port int) {
	reader := bufio.NewReader(os.Stdin)

	tmpReadLine := func() string {
		line, isPrefix, err := reader.ReadLine()
		if isPrefix || err != nil {
			panic(fmt.Sprintf("isPrefix=%v,err=%v", isPrefix, err))
		}
		return string(line)
	}

	fmt.Printf("请输入 Host: ")
	Host = tmpReadLine()

	for {
		var err error
		fmt.Printf("请输入 Port: ")
		if Port, err = strconv.Atoi(tmpReadLine()); err != nil {
			fmt.Println("解析失败, 请重新输入!")
		} else {
			if 1 <= Port && Port <= 65535 {
				break
			} else {
				fmt.Println("请输入[1~65535]的数字!")
			}
		}
	}

	return
}
