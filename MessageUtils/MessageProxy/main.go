package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/zx9229/zxgo_temp/MessageUtils/TxStruct"
)

func main() {
	dataProxy := new_DataProxy()
	if err := dataProxy.Init("sqlite3", "test_proxy.db", "Asia/Shanghai"); err != nil {
		panic(err)
	}
	if slice_, err := dataProxy.QueryData(); err != nil {
		for _, item := range slice_ {
			for cnt := 0; xxx(&item, cnt); cnt++ {
				if 100 < cnt {
					panic(cnt)
				}
			}
		}
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
