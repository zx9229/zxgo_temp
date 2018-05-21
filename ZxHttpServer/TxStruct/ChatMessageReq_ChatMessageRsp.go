package TxStruct //通信结构体.
import "reflect"

//聊天消息,请求
//将RecverId和RecverAlias取并集,这个聊天消息会发给这个并集里的成员;GroupId和GroupAlias同理.
type ChatMessageReq struct {
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

func (self *ChatMessageReq) FillField_Type() {
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
