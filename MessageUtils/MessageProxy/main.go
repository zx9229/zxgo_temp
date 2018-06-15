package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/zx9229/zxgo_temp/MessageUtils/TxStruct"
)

//服务器必须要有紧急通道,如果大量的数据一直提交不上来的话,需要通过紧急通道告诉服务器
//因为人员肯定会去看服务器里的信息,所以紧急信息可以被人员看到,从而被人们所知.
//建议: body里面写上紧急信息的内容,服务器专门有一个紧急信息表,key是int64,value是string,
//这样,只要服务器能收到数据,就肯定能被存下来.

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
