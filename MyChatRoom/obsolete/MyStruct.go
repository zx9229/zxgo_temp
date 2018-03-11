package main_obsolete

import (
	"fmt"
	"time"
)

type KeyValue struct {
	Key   string `xorm:"pk"`
	Value string
}

//原始聊天消息
type ChatMessageRaw struct {
	MyTn       string    `xorm:"-"` //(我的表名)这个Field将不进行字段映射
	Id         int64     `xorm:"pk autoincr unique notnull"`
	Dttm       time.Time `xorm:"updated"`
	SendUser   string
	RecvUsers  []string
	RecvGroups []string
	Message    string
	Memo       string //备注(预留字段)
}

func (self *ChatMessageRaw) TableName() string {
	if len(self.MyTn) <= 0 {
		panic(fmt.Sprintf("字段的值=%v,请正确赋值", self.MyTn))
	}
	return self.MyTn
}

//整合后的聊天消息
type ChatMessage struct {
	MyTn     string    `xorm:"-"` //这个Field将不进行字段映射
	Id       int64     `xorm:"pk autoincr unique notnull"`
	Dttm     time.Time `xorm:"updated"`
	SendUser string
	Message  string
	Memo     string //备注(预留字段)
}

func (self *ChatMessage) TableName() string {
	if len(self.MyTn) <= 0 {
		panic(fmt.Sprintf("字段的值=%v,请正确赋值", self.MyTn))
	}
	return self.MyTn
}

type GroupData struct {
	Name       string    `xorm:"pk"` //组名
	Superuser  string    //这个组的超级管理员是谁
	Members    []string  //这个组的成员都有谁
	UpdateTime time.Time //更新时间
	Memo       string    //备注(预留字段)
}

type UserData struct {
	Name       string           `xorm:"pk"` //用户名
	Password   string           //密码
	GroupPos   map[string]int64 //组里的消息,接收到哪个位置了.
	FriendPos  map[string]int64 //好友的消息,接收到哪个位置了.
	UpdateTime time.Time        //更新时间
	Memo       string           //备注(预留字段)
}
