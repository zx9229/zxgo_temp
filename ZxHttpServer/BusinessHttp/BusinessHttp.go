package BusinessHttp

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"reflect"

	"github.com/zx9229/zxgo_temp/ZxHttpServer/TxStruct"
)

func panicIfNil(obj interface{}) {
	//一般在执行类的成员函数时使用,
	//不用:成员函数里面未使用数据成员;(此时,该成员函数可以当做静态成员函数)
	//使用:成员函数里面使用了数据成员;这样可以更快的发现问题.
	if obj == nil {
		panic("object is nil and will call panic!")
	}
}

type BusinessHttp struct {
	str2type map[string]reflect.Type
}

func New_BusinessHttp() *BusinessHttp {
	curData := new(BusinessHttp)
	curData.str2type = TxStruct.CalcMapStr2Type()
	return curData
}

func (self *BusinessHttp) Handler_ROOT(w http.ResponseWriter, r *http.Request) {
	panicIfNil(self) //这里直接使用了[数据成员],所以需要[panicIfNil(self)]操作.

	sliceTypeName := make([]string, 0)
	for typeName := range self.str2type {
		sliceTypeName = append(sliceTypeName, typeName)
	}

	if t, err := template.ParseFiles("template/WebSocket.html"); err != nil {
		fmt.Fprintf(w, "%v", err)
	} else {
		if err = t.Execute(w, sliceTypeName); err != nil {
			fmt.Fprintf(w, "%v", err)
		}
	}
}

func (self *BusinessHttp) Handler_TxStruct(w http.ResponseWriter, r *http.Request) {
	//这里仅仅调用了其它的[成员函数],没有直接使用[数据成员],所以不需要[panicIfNil(self)]操作.
	var typeName string
	if queryForm, err := url.ParseQuery(r.URL.RawQuery); err != nil {
		fmt.Fprintf(w, "解析GET参数报错%v", err)
		return
	} else {
		if typeNameSlice, ok := queryForm["Type"]; !ok {
			fmt.Fprintf(w, "解析GET参数没有Type")
			return
		} else {
			typeName = typeNameSlice[0]
		}
	}

	if jsonStr, err := self.fromTypeNameToJsonStr(typeName); err != nil {
		fmt.Fprintf(w, "内部逻辑出错:%v", err)
	} else {
		fmt.Fprintf(w, "%s", jsonStr)
	}
}

func (self *BusinessHttp) fromTypeNameToJsonStr(typeName string) (jsonStr string, err error) {
	panicIfNil(self)

	var curType reflect.Type
	if tmpType, ok := self.str2type[typeName]; !ok {
		err = fmt.Errorf("找不到名字为[%s]的类型", typeName)
		return
	} else {
		curType = tmpType
	}

	var jsonByte []byte
	//构造一个"最小可用的通信json串"
	if jsonByte, err = json.Marshal(&TxStruct.TxBaseData{Type: typeName}); err != nil {
		return
	}

	objData := reflect.New(curType).Interface()
	//将"最小可用的通信json串"转成"目标结对象",此时,如果目标对象里面含有slice或map等数据成员的话,它们的值是nil
	if err = json.Unmarshal(jsonByte, objData); err != nil {
		return
	}

	//将"目标结对象"转成"目标对象的json串"
	if jsonByte, err = json.Marshal(objData); err != nil {
		return
	}

	jsonStr = string(jsonByte)
	return
}
