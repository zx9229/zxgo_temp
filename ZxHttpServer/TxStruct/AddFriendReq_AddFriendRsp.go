package TxStruct //通信结构体.
import (
	"reflect"
)

type AddFriendReq struct {
	Type       string
	TransmitId int64
	UserId     int64
	FriendId   int64
}

func (self *AddFriendReq) FillField_Type() {
	self.Type = reflect.ValueOf(*self).Type().Name()
}

type AddFriendRsp struct {
	Type       string
	TransmitId int64
	Code       int
	Message    string
	UserId     int64
	FriendId   int64
}

func (self *AddFriendRsp) FillField_Type() {
	self.Type = reflect.ValueOf(*self).Type().Name()
}

func (self *AddFriendRsp) FillField_FromReq(reqObj *AddFriendReq) {
	self.TransmitId = reqObj.TransmitId
	self.UserId = reqObj.UserId
	self.FriendId = reqObj.FriendId
}
