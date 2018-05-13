package main

import (
	"fmt"
	"time"
)

//pk(主键)里是可以插入NULL的,所以,如果不想出现NULL值,应当是"pk notnull"
//unique和pk有什么区别,暂时我还不知道.

//默认,别名是ID的字符串化的值,如果要修改别名,别名需要是"非数字开头"
//我要在组里说话,消息发送到服务器,服务器写数据到数据库,通知所有的socket[某张表有改变],
//对于每一个socket,我和这张表有关联吗,如果有,根据已有的pos读取剩余数据,发送给自己的对端,更新pos,
//可能要对ID作部分保留,比如,ID=1的用户作为广播用户,等.
//可能还要有一个通知表,通知表里面,如果ID=0,表示向全体user/group广播这条消息.

type KeyValue struct {
	Key        string    `xorm:"notnull pk"` //键
	Value      string    //值
	UpdateTime time.Time `xorm:"updated"` //更新时间
}

type UserData struct {
	Id         int64           `xorm:"notnull pk autoincr"` //类似QQ的唯一ID(不可修改,全局唯一)
	Alias      string          `xorm:"notnull unique"`      //别名(可以修改,全局唯一)
	Password   string          //密码
	GroupPos   map[int64]int64 //组里的消息,接收到哪个位置了.
	FriendPos  map[int64]int64 //好友的消息,接收到哪个位置了.
	NoticePos  int64           //  通知消息,接收到哪个位置了.
	CreateTime time.Time       `xorm:"created"` //创建时间
	UpdateTime time.Time       `xorm:"updated"` //更新时间
	Memo       string          //备注(预留字段)
}

type GroupData struct {
	Id         int64     `xorm:"notnull pk autoincr"` //类似QQ的唯一ID(不可修改,全局唯一)
	Alias      string    `xorm:"notnull unique"`      //别名(可以修改,全局唯一)
	SuperId    int64     //这个组的超级管理员ID号(预留字段)
	AdminId    []int64   //这个组的普通管理员ID号(预留字段)
	OtherMemId []int64   //这个组的其他成员的ID号
	CreateTime time.Time `xorm:"created"` //创建时间
	UpdateTime time.Time `xorm:"updated"` //更新时间
	Memo       string    //备注(预留字段)
}

//原始通知消息(服务器刚刚接收到的最原始消息)
type PushMessageRaw struct {
	Id          int64     `xorm:"notnull pk autoincr"`
	Dttm        time.Time `xorm:"updated"`
	SenderId    int64     //发送者的ID号
	SenderAlias string    //发送者的别名
	RecverId    []int64   //接收者的ID号列表
	RecverAlias []string  //接收者的别名列表
	GroupId     []int64   //接收组的ID号列表
	GroupAlias  []string  //接收组的别名列表
	Message     string    //聊天消息内容
	Memo        string    //备注(预留字段)
}

//整理后的通知消息(最终呈现给用户的通知消息)
type PushMessage struct {
	Id          int64     `xorm:"notnull pk autoincr"`
	Dttm        time.Time `xorm:"updated"`
	IdRaw       int64     //此行记录对应的原始ID号
	SenderId    int64     //发送者的ID号
	SenderAlias string    //发送者的别名
	RecverId    []int64   //接收者的ID号列表
	RecverAlias []string  //接收者的别名列表
	GroupId     []int64   //接收组的ID号列表
	GroupAlias  []string  //接收组的别名列表
	Message     string    //聊天消息内容
	Memo        string    //备注(预留字段)
}

//原始聊天消息(服务器刚刚接收到的最原始消息)
type ChatMessageRaw struct {
	Id          int64     `xorm:"notnull pk autoincr"`
	Dttm        time.Time `xorm:"updated"`
	SenderId    int64     //发送者的ID号
	SenderAlias string    //发送者的别名
	RecverId    []int64   //接收者的ID号列表
	RecverAlias []string  //接收者的别名列表
	GroupId     []int64   //接收组的ID号列表
	GroupAlias  []string  //接收组的别名列表
	Message     string    //聊天消息内容
	Memo        string    //备注(预留字段)
}

//整理后的聊天消息(最终呈现给用户的聊天消息)
type ChatMessage struct {
	MyTn        string    `xorm:"-"` //(我的表名)这个Field将不进行字段映射
	Id          int64     `xorm:"notnull pk autoincr unique"`
	IdRaw       int64     //此行记录对应的原始ID号
	Dttm        time.Time `xorm:"updated"`
	SenderId    int64     //发送者的ID号
	SenderAlias string    //发送者的别名
	Message     string    //聊天消息内容
	Memo        string    //备注(预留字段)
}

func (self *ChatMessage) TableName() string {
	if len(self.MyTn) <= 0 {
		panic(fmt.Sprintf("字段的值=%v,请正确赋值", self.MyTn))
	}
	return self.MyTn
}

func innerNewUserData() *UserData {
	newData := &UserData{}
	newData.Id = 0
	newData.Alias = ""
	newData.Password = ""
	newData.GroupPos = make(map[int64]int64)
	newData.FriendPos = make(map[int64]int64)
	newData.NoticePos = 0
	newData.CreateTime = time.Now()
	newData.UpdateTime = time.Now()
	newData.Memo = ""
	return newData
}
