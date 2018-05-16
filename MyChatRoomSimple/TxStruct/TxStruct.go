package TxStruct //通信结构体.
import (
	"encoding/json"
	"fmt"
	"reflect"
)

type BaseTxData struct {
	Type string
}

type RspMessage struct {
	Type         string
	TransmitId   int64 //传输ID(客户端带过来一个序号,服务器会返回同样的数字)
	Code         int
	Message      string
	SucceedId    map[int64]string
	FailedId     map[int64]string
	SucceedGroup map[int64]string
	FailedGroup  map[int64]string
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

func InitMap() map[string]reflect.Type {
	slice_ := make([]interface{}, 0)
	slice_ = append(slice_, RspMessage{})
	slice_ = append(slice_, ChatMessage{})

	cacheData := map[string]reflect.Type{}
	for _, element := range slice_ {
		curType := reflect.ValueOf(element).Type()
		cacheData[curType.Name()] = curType
	}

	return cacheData
}

func ThisIsExample() {
	cacheTypeData := InitMap()
	testMsg := RspMessage{}
	testMsg.Type = reflect.ValueOf(testMsg).Type().Name()
	testMsg.TransmitId = 1
	testMsg.Code = 2
	testMsg.Message = "testting..."
	if jsonByte, err := json.Marshal(testMsg); err != nil {
		panic(err)
	} else {
		var baseData *BaseTxData = &BaseTxData{}
		if err := json.Unmarshal(jsonByte, baseData); err != nil {
			panic(err)
		}
		if curType, ok := cacheTypeData[baseData.Type]; !ok {
			panic("")
		} else {
			curValue := reflect.New(curType).Interface()
			if err = json.Unmarshal(jsonByte, curValue); err != nil {
				panic(err)
			}
			fmt.Println(curValue)
			if baseData.Type == "RspMessage" {
				fmt.Println("这一段,仍然是硬编码,我还不知道要怎么处理...")
			}
		}
	}
}
