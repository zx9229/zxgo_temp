package TxStruct //通信结构体.
import (
	"reflect"
)

//创建某用户(用户一旦创建,不可删除,只能禁用[为了数据唯一性]).
type AddUserReq struct {
	Type         string
	TransmitId   int64
	OperatorId   int64  //操作员,一般是超管/用户本人.
	UserAlias    string //用户的别名.
	UserPassword string //用户的密码.
}

func (self *AddUserReq) FillField_Type() {
	self.Type = reflect.ValueOf(*self).Type().Name()
}

type AddUserRsp struct {
	Type         string //类型的名字
	TransmitId   int64  //传输ID(客户端带过来一个序号,服务器会返回同样的数字)
	Code         int    //返回值
	Message      string //返回的详细信息
	OperatorId   int64
	UserId       int64
	UserAlias    string
	UserPassword string
}

func (self *AddUserRsp) FillField_Type() {
	self.Type = reflect.ValueOf(*self).Type().Name()
}

func (self *AddUserRsp) FillField_FromReq(reqObj *AddUserReq) {
	self.FillField_Type()
	self.TransmitId = reqObj.TransmitId
	self.OperatorId = reqObj.OperatorId
	self.UserAlias = reqObj.UserAlias
	self.UserPassword = reqObj.UserPassword
}
