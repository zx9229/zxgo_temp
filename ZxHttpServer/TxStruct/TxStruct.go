package TxStruct //通信结构体.
import "reflect"

type TxBaseData struct {
	Type string
}

//聊天消息
//将RecverId和RecverAlias取并集,这个聊天消息会发给这个并集里的成员;GroupId和GroupAlias同理.
type ChatMessage struct {
	Type        string   //reflect.ValueOf(ChatMessage{}).Type().Name()
	TransmitId  int64    //传输ID(客户端带过来一个序号,服务器会返回同样的数字)
	SenderId    int64    //发送者的ID号
	SenderAlias string   //发送者的别名
	RecverId    []int64  //接收者的ID号列表
	RecverAlias []string //接收者的别名列表
	GroupId     []int64  //接收组的ID号列表
	GroupAlias  []string //接收组的别名列表
	Message     string   //聊天消息内容
	Memo        string   //备注(预留字段)
}

func (self *ChatMessage) FillField_Type() {
	self.Type = reflect.ValueOf(*self).Type().Name()
}

// 聊天消息,响应
type ChatMessageRsp struct {
	Type         string           //类型的名字
	TransmitId   int64            //传输ID(客户端带过来一个序号,服务器会返回同样的数字)
	Code         int              //返回值
	Message      string           //返回的详细信息
	SucceedId    map[int64]string //[key]是[id],[value]是[alias]
	FailedId     map[int64]string //[key]是[id],[value]是[alias]
	SucceedGroup map[int64]string //[key]是[id],[value]是[alias]
	FailedGroup  map[int64]string //[key]是[id],[value]是[alias]
}

func (self *ChatMessageRsp) FillField_Type() {
	self.Type = reflect.ValueOf(*self).Type().Name()
}

//通知消息
type PushMessage struct {
	Type        string
	TransmitId  int64    //传输ID(客户端带过来一个序号,服务器会返回同样的数字)
	SenderId    int64    //发送者的ID号
	SenderAlias string   //发送者的别名
	RecverId    []int64  //接收者的ID号列表
	RecverAlias []string //接收者的别名列表
	GroupId     []int64  //接收组的ID号列表
	GroupAlias  []string //接收组的别名列表
	Message     string   //聊天消息内容
	Memo        string   //备注(预留字段)
}

func (self *PushMessage) FillField_Type() {
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

//创建某用户(用户一旦创建,不可删除,只能禁用[为了数据唯一性]).
type CreateUser struct {
	Type       string
	TransmitId int64
	OperatorId int64  //操作员,一般是超管/用户本人.
	Alias      string //用户的别名.
	Password   string //用户的密码.
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

type Login struct {
	Type       string
	TransmitId int64
	UserId     int64  //要登录用户的ID.(优先使用ID,ID无效[id<=0]则使用别名)
	UserAlias  string //要登录的用户的别名.(len(alias)>0)
	Password   string
}

type LoginRsp struct {
	Type       string
	TransmitId int64
	Code       int
	Message    string
	UserId     int64
	UserAlias  string
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
