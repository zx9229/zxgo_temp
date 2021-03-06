package TxStruct //通信结构体.
import (
	"encoding/json"
	"reflect"
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

//通知消息
type PushMessageReq struct {
	Type        string
	TransmitId  int64    //传输ID(客户端带过来一个序号,服务器会返回同样的数字)
	SenderId    int64    //发送者的ID号(优先使用ID,ID无效[id<=0]则使用别名)
	SenderAlias string   //发送者的别名
	RecverId    []int64  //接收者的ID号列表
	RecverAlias []string //接收者的别名列表
	GroupId     []int64  //接收组的ID号列表
	GroupAlias  []string //接收组的别名列表
	Message     string   //聊天消息内容
	Memo        string   //备注(预留字段)
}

func (self *PushMessageReq) FillField_Type() {
	self.Type = reflect.ValueOf(*self).Type().Name()
}

type PushMessageRsp struct {
	Type         string           //类型的名字
	TransmitId   int64            //传输ID(客户端带过来一个序号,服务器会返回同样的数字)
	Code         int              //返回值
	Message      string           //返回的详细信息
	SucceedId    map[int64]string //[key]是[id],[value]是[alias]
	FailedId     map[int64]string //[key]是[id],[value]是[alias]
	SucceedGroup map[int64]string //[key]是[id],[value]是[alias]
	FailedGroup  map[int64]string //[key]是[id],[value]是[alias]
}

func (self *PushMessageRsp) FillField_Type() {
	self.Type = reflect.ValueOf(*self).Type().Name()
}

//禁用某用户
type DisableUser struct {
	Type       string
	TransmitId int64
	OperatorId int64 //操作员,一般是超管/用户本人.
	UserId     int64 //要禁用的用户.
}

//启用某用户
type EnableUser struct {
	Type       string
	TransmitId int64
	OperatorId int64 //操作员,一般是超管.
	UserId     int64 //要启用的用户.
}

type Logout struct {
	Type       string
	TransmitId int64
	UserId     int64  //要登出用户的ID.
	UserAlias  string //要登出的用户的别名.
}

type LogoutRsp struct {
	Type       string
	TransmitId int64
	Code       int
	Message    string
	UserId     int64
	UserAlias  string
}

type CreateGroup struct {
	Type       string
	TransmitId int64
	UserId     int64 //操作员,创建成功后,会成为这个group的SuperId.
	GroupAlias string
}

type CreateGroupRsp struct {
	Type       string
	TransmitId int64
	Code       int
	Message    string
	GroupId    int64
	GroupAlias string
	SuperId    int64
}

type QueryUser struct {
	Type       string
	TransmitId int64
}
