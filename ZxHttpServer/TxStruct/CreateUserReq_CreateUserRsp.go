package TxStruct //通信结构体.
import (
	"encoding/json"
	"reflect"
)

//创建某用户(用户一旦创建,不可删除,只能禁用[为了数据唯一性]).
type CreateUserReq struct {
	Type         string
	TransmitId   int64
	OperatorId   int64  //操作员,一般是超管/用户本人.
	UserAlias    string //用户的别名.
	UserPassword string //用户的密码.
}

func (self *CreateUserReq) FillField_Type() {
	self.Type = reflect.ValueOf(*self).Type().Name()
}

type CreateUserRsp struct {
	Type         string //类型的名字
	TransmitId   int64  //传输ID(客户端带过来一个序号,服务器会返回同样的数字)
	Code         int    //返回值
	Message      string //返回的详细信息
	OperatorId   int64
	UserId       int64
	UserAlias    string
	UserPassword string
}

func (self *CreateUserRsp) FillField_Type() {
	self.Type = reflect.ValueOf(*self).Type().Name()
}

func (self *CreateUserRsp) FillField_FromReq(reqObj *CreateUserReq) {
	self.FillField_Type()
	self.TransmitId = reqObj.TransmitId
	self.OperatorId = reqObj.OperatorId
	self.UserAlias = reqObj.UserAlias
	self.UserPassword = reqObj.UserPassword
}

func (self *CreateUserRsp) ToJsonStr() string {
	if jsonByte, err := json.Marshal(self); err != nil {
		panic(err)
	} else {
		return string(jsonByte)
	}
}
