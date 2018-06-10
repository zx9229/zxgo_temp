package TxStruct //通信结构体.
import (
	"encoding/json"
)

func ToJsonStr(v interface{}) string {
	if jsonByte, err := json.Marshal(v); err != nil {
		panic(err)
	} else {
		return string(jsonByte)
	}
}

type TxBaseData struct {
	Type string
}
