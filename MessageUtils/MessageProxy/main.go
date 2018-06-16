package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
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

//这一组程序,写到最后,好像变成了一个"日志收集器",我是看nxlog介绍的时候发现的.
//服务器必须要有紧急通道,如果大量的数据一直提交不上来的话,需要通过紧急通道告诉服务器
//因为人员肯定会去看服务器里的信息,所以紧急信息可以被人员看到,从而被人们所知.
//建议: body里面写上紧急信息的内容,服务器专门有一个紧急信息表,
// CREATE TABLE tn(Id int64, Message string);  // Message里面可以是任何内容(比如json).
//这样,只要服务器能收到数据,就肯定能被存下来.
//如果服务器收不到数据,那就是网络不通/服务器挂了,这样的话,基础假设都不存在,那就不用想了.

func main() {
	helpPtr := flag.Bool("help", false, "show this help.")
	hostPtr := flag.String("host", "localhost", "set the server address")
	portPtr := flag.Int("port", 0, "set the server port")
	intervalPtr := flag.Int("interval", 500, "set the scan interval(ms)")
	flag.Parse()
	if *helpPtr {
		flag.Usage()
		return
	}

	var err error

	if err = PrepareWorkDir(); err != nil {
		panic(err)
	}

	dataProxy := new_DataProxy()
	if err = dataProxy.Init(TxStruct.DRIVER_NAME, TxStruct.DATA_SOURCE_NAME); err != nil {
		panic(err)
	}

	var cfg *ConfigInfo
	if cfg, err = PrepareConfig(dataProxy, *hostPtr, *portPtr, *intervalPtr); err != nil {
		panic(err)
	}

	if err = dataProxy.FlushExeInfo(); err != nil {
		panic(err)
	}

	for {
		dataProxy.FlushExeInfo()
		if slice_, err := dataProxy.QueryProxyReqRsp(); err == nil {
			for _, item := range slice_ {
				if doReportAction(cfg.Host, cfg.Port, &item) {
					continue
				}
				if err = dataProxy.UpdateData(&item); err != nil {
					panic(err)
				}
			}
		}
		time.Sleep(time.Duration(*intervalPtr) * time.Millisecond)
	}
}

func PrepareWorkDir() error {
	var err error
	var path string
	if path, err = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
		return err
	}
	if err = os.Chdir(path); err != nil {
		return err
	}
	if _, err = os.Getwd(); err != nil {
		return err
	}
	if os.Getpid() <= 0 {
		err = errors.New("Getpid ERROR")
		return err
	}
	return err
}

func CheckConfigPassed(cfg *ConfigInfo) bool {
	if cfg == nil {
		return false
	}
	if !(0 < cfg.Port && cfg.Port < 65536) {
		return false
	}
	return true
}

func PrepareConfig(dataProxy *DataProxy, host string, port int, interval int) (cfg *ConfigInfo, err error) {
	if 0 < flag.NFlag() {
		cfg = new(ConfigInfo)
		cfg.Host = host
		cfg.Port = port
		cfg.Interval = interval
	} else {
		if cfg, err = dataProxy.LoadConfigInfo(); err != nil {
			return
		}
	}
	if !CheckConfigPassed(cfg) {
		err = errors.New("not CheckConfigPassed")
	}
	if 0 < flag.NFlag() {
		if err = dataProxy.SaveConfigInfo(cfg); err != nil {
			return
		}
	}
	return
}

func doReportAction(host string, port int, reqRsp *TxStruct.ProxyReqRsp) bool {
	//返回值(bool)=>是否还需要重新处理它(true=>需要重新处理).
	var needRetry bool = false

	var err error
	var byteSlice []byte

	url := fmt.Sprintf("http://%s:%d/ReportReq", host, port)

	for _ = range "1" {
		var reqData *TxStruct.ReportReq = ProxyReqRsp_ToReq(reqRsp)
		if byteSlice, err = json.Marshal(reqData); err != nil {
			reqRsp.IsPending = false
			reqRsp.RspId = -1
			reqRsp.RspCode = 1
			reqRsp.Message = fmt.Sprintf("[Proxy]转换成ReportReq失败,err=%v", err)
			//
			needRetry = false
			break
		}

		var resp *http.Response
		if resp, err = http.Post(url, "application/json", strings.NewReader(string(byteSlice))); err != nil {
			needRetry = true
			break
		}

		defer resp.Body.Close()

		if byteSlice, err = ioutil.ReadAll(resp.Body); err != nil {
			reqRsp.IsPending = false
			reqRsp.RspId = -1
			reqRsp.RspCode = 1
			reqRsp.Message = fmt.Sprintf("[Proxy]ReadAll失败,err=%v", err)
			//
			needRetry = false
			break
		}

		rspData := new(TxStruct.ReportRsp)
		if err = json.Unmarshal(byteSlice, rspData); err != nil {
			reqRsp.IsPending = false
			reqRsp.RspId = -1
			reqRsp.RspCode = 1
			reqRsp.Message = fmt.Sprintf("[Proxy]转换成ReportRsp失败,err=%v", err)
			//
			needRetry = false
			break
		}

		if err = ProxyReqRsp_FillWithRsp(reqRsp, rspData, false); err != nil {
			reqRsp.IsPending = false
			reqRsp.RspId = -1
			reqRsp.RspCode = 1
			reqRsp.Message = fmt.Sprintf("[Proxy]转换成ReportRsp失败,err=%v", err)
			//
			needRetry = false
			break
		}
		needRetry = false
	}
	return needRetry
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
