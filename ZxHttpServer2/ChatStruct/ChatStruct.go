package ChatStruct

import (
	"time"
)

type MessageRaw struct {
	Id         int64
	SenderId   int64
	RecvId     []int64
	RecvAlias  []string
	GroupId    []int64
	GroupAlias []string
	Message    string
	Memo       string
	UpdateTime time.Time
}

type MessageCache struct {
	Id         int64
	SenderId   int64
	RecvId     map[int64]bool
	GroupId    map[int64]bool
	Message    string
	Memo       string
	UpdateTime time.Time
}

type MessageData struct {
	Id         int64
	IdCache    int64  //MessageCache中的ID.
	Tag        string //c_id1_id2_(2人聊天的表),g_id_(group里聊天),pc_id_(推送给个人的消息),pg_id_(推送给组的消息)
	TagIdx     int64  //在这个标签里,本消息的序号.
	Sender     int64  //发送者的ID
	Receiver   int64  //接收者的ID/接收组的ID
	Message    string //聊天消息内容
	Memo       string //备注(预留字段)
	UpdateTime time.Time
}

type KeyValue struct {
	Key        string    `xorm:"notnull pk"` //键
	Value      string    //值
	UpdateTime time.Time `xorm:"updated"` //更新时间
}

type UserData struct {
	Id         int64          `xorm:"notnull pk autoincr"` //类似QQ的唯一ID(不可修改,全局唯一)
	Alias      string         `xorm:"notnull unique"`      //别名(可以修改,全局唯一)
	Password   string         //密码
	Friends    map[int64]bool //它和哪些人是好友.
	Groups     map[int64]bool //它是哪些组的成员.
	CreateTime time.Time      `xorm:"created"` //创建时间
	UpdateTime time.Time      `xorm:"updated"` //更新时间
	Memo       string         //备注(预留字段)
}

type GroupData struct {
	Id         int64          `xorm:"notnull pk autoincr"` //类似QQ的唯一ID(不可修改,全局唯一)
	Alias      string         `xorm:"notnull unique"`      //别名(可以修改,全局唯一)
	SuperId    int64          //这个组的超级管理员ID号(预留字段)
	AdminId    map[int64]bool //这个组的普通管理员ID号(预留字段)
	OtherMemId map[int64]bool //这个组的其他成员的ID号
	CreateTime time.Time      `xorm:"created"` //创建时间
	UpdateTime time.Time      `xorm:"updated"` //更新时间
	Memo       string         //备注(预留字段)
}

func New_UserData() *UserData { //因为这个结构体里面有好些map,所以不建议直接new,建议调用它进行创建.
	newData := &UserData{}
	//
	newData.Id = 0
	newData.Alias = ""
	newData.Password = ""
	newData.Friends = make(map[int64]bool)
	newData.Groups = make(map[int64]bool)
	newData.CreateTime = time.Now()
	newData.UpdateTime = time.Now()
	newData.Memo = ""
	//
	return newData
}
