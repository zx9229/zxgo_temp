package TxStruct //通信结构体.
import "reflect"

type LoginReq struct {
	Type       string
	TransmitId int64
	UserId     int64  //要登录用户的ID.(优先使用ID,ID无效[id<=0]则使用别名)
	UserAlias  string //要登录的用户的别名.(len(alias)>0)
	Password   string
}

func (self *LoginReq) FillField_Type() {
	self.Type = reflect.ValueOf(*self).Type().Name()
}

type LoginRsp struct {
	Type       string
	TransmitId int64
	Code       int
	Message    string
	UserId     int64
	UserAlias  string
}

func (self *LoginRsp) FillField_Type() {
	self.Type = reflect.ValueOf(*self).Type().Name()
}
func (self *LoginRsp) FillField_FromReq(reqObj *LoginReq) {
	self.FillField_Type()
	self.TransmitId = reqObj.TransmitId
	self.UserId = reqObj.UserId
	self.UserAlias = reqObj.UserAlias
}
