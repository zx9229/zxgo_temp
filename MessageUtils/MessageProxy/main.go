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
	"regexp"
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

type ArgData struct {
	helpPtr          *bool
	hostPtr          *string
	portPtr          *int
	scanIntervalPtr  *int //扫描间隔(相邻的2次扫描SQLITE的间隔).
	retryIntervalPtr *int //重试间隔(发往服务器失败的时候,重试的间隔).
}

func main() {
	argData := new(ArgData)
	argData.helpPtr = flag.Bool("help", false, "show this help.")
	argData.hostPtr = flag.String("host", "localhost", "set the server address")
	argData.portPtr = flag.Int("port", 0, "set the server port")
	argData.scanIntervalPtr = flag.Int("scan", 500, "set the scan interval(ms)")
	argData.retryIntervalPtr = flag.Int("retry", 5000, "set the retry interval(ms)")
	flag.Parse()
	if *argData.helpPtr {
		flag.Usage()
		return
	}

	var err error

	if err = PrepareWorkDir(); err != nil {
		fmt.Fprintln(os.Stderr, "[PrepareWorkDir]", err)
		os.Exit(1)
	}

	dataProxy := new_DataProxy()
	if err = dataProxy.Init(TxStruct.DRIVER_NAME, TxStruct.DATA_SOURCE_NAME); err != nil {
		fmt.Fprintln(os.Stderr, "[Init]", err)
		os.Exit(1)
	}

	var cfg *ConfigInfo
	if cfg, err = PrepareConfig(dataProxy, argData); err != nil {
		fmt.Fprintln(os.Stderr, "[PrepareConfig]", err)
		os.Exit(1)
	}

	if err = dataProxy.FlushExeInfo(); err != nil {
		fmt.Fprintln(os.Stderr, "[FlushExeInfo]", err)
		os.Exit(1)
	}

	for {
		dataProxy.FlushExeInfo()
		if slice_, err := dataProxy.QueryProxyReqRsp(); err == nil {
			for _, item := range slice_ {
				for !ReportDataFinish(cfg.Host, cfg.Port, &item) {
					time.Sleep(time.Duration(*argData.retryIntervalPtr) * time.Millisecond)
				}
				if err = dataProxy.UpdateProxyReqRsp(&item); err != nil {
					panic(err)
				}
			}
		}
		time.Sleep(time.Duration(*argData.scanIntervalPtr) * time.Millisecond)
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
		err = errors.New("Getpid anomalous")
		return err
	}
	return err
}

func CheckConfig(cfg *ConfigInfo) error {
	var err error
	for _ = range "1" {
		if cfg == nil {
			err = errors.New("data is nil")
			break
		}
		if cfg.Host != "localhost" {
			pattern := `^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$`
			if ok, err2 := regexp.MatchString(pattern, cfg.Host); !ok || err2 != nil {
				err = fmt.Errorf("host(%v) is anomalous", cfg.Host)
				break
			}
		}
		if !(0 < cfg.Port && cfg.Port < 65536) {
			err = fmt.Errorf("port(%v) is anomalous", cfg.Port)
			break
		}
		if cfg.ScanInterval <= 0 {
			err = fmt.Errorf("ScanInterval(%v) is anomalous", cfg.ScanInterval)
			break
		}
		if cfg.RetryInterval <= 0 {
			err = fmt.Errorf("RetryInterval(%v) is anomalous", cfg.RetryInterval)
			break
		}
	}
	return err
}

func PrepareConfig(dataProxy *DataProxy, argData *ArgData) (cfg *ConfigInfo, err error) {
	//如果程序以带参的方式启动,就检查参数,保存,启动
	//否则,就读数据库里的配置,检查参数,启动.

	if 0 < flag.NFlag() {
		cfg = ConvertToConfigInfo(argData)
	} else {
		if cfg, err = dataProxy.LoadConfigInfo(); err != nil {
			cfg = nil
			return
		}
	}

	if err = CheckConfig(cfg); err != nil {
		cfg = nil
		return
	}

	if 0 < flag.NFlag() {
		if err = dataProxy.SaveConfigInfo(cfg); err != nil {
			return
		}
	}

	return
}

func ReportDataFinish(host string, port int, reqRsp *TxStruct.ProxyReqRsp) bool {
	//返回值(bool)=>是否还需要重新处理它(true=>需要重新处理).
	var isFinish bool = false

	var err error
	var byteSlice []byte

	url := fmt.Sprintf("http://%v:%v/ReportReq", host, port)

	for _ = range "1" {
		var reqData *TxStruct.ReportReq = ProxyReqRsp_ToReq(reqRsp)
		if byteSlice, err = json.Marshal(reqData); err != nil {
			reqRsp.IsPending = false
			reqRsp.RspId = -1
			reqRsp.RspCode = 1
			reqRsp.Message = fmt.Sprintf("[Proxy]转换成ReportReq失败,err=%v", err)
			//
			isFinish = true
			break
		}

		var resp *http.Response
		if resp, err = http.Post(url, "application/json", strings.NewReader(string(byteSlice))); err != nil {
			isFinish = false
			break
		}

		defer resp.Body.Close()

		if byteSlice, err = ioutil.ReadAll(resp.Body); err != nil {
			reqRsp.IsPending = false
			reqRsp.RspId = -1
			reqRsp.RspCode = 1
			reqRsp.Message = fmt.Sprintf("[Proxy]ReadAll失败,err=%v", err)
			//
			isFinish = true
			break
		}

		rspData := new(TxStruct.ReportRsp)
		if err = json.Unmarshal(byteSlice, rspData); err != nil {
			reqRsp.IsPending = false
			reqRsp.RspId = -1
			reqRsp.RspCode = 1
			reqRsp.Message = fmt.Sprintf("[Proxy]转换成ReportRsp失败,err=%v", err)
			//
			isFinish = true
			break
		}

		if err = ProxyReqRsp_FillWithRsp(reqRsp, rspData, false); err != nil {
			reqRsp.IsPending = false
			reqRsp.RspId = -1
			reqRsp.RspCode = 1
			reqRsp.Message = fmt.Sprintf("[Proxy]转换成ReportRsp失败,err=%v", err)
			//
			isFinish = true
			break
		}
		reqRsp.IsPending = false
		isFinish = true
	}
	return isFinish
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

func ConvertToConfigInfo(argData *ArgData) *ConfigInfo {
	cfg := new(ConfigInfo)
	cfg.Host = *argData.hostPtr
	cfg.Port = *argData.portPtr
	cfg.ScanInterval = *argData.scanIntervalPtr
	cfg.RetryInterval = *argData.retryIntervalPtr
	return cfg
}
