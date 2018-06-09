package main

import (
	"errors"
	"fmt"
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
	RecvId     []int64
	GroupId    []int64
	Message    string
	Memo       string
	UpdateTime time.Time
}

type MessageData struct {
	Id         int64
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
	Id         int64     `xorm:"notnull pk autoincr"` //类似QQ的唯一ID(不可修改,全局唯一)
	Alias      string    `xorm:"notnull unique"`      //别名(可以修改,全局唯一)
	Password   string    //密码
	Friends    []int64   //它和哪些人是好友.
	Groups     []int64   //它是哪些组的成员.
	CreateTime time.Time `xorm:"created"` //创建时间
	UpdateTime time.Time `xorm:"updated"` //更新时间
	Memo       string    //备注(预留字段)
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

type InnerCacheData struct { //内存中的缓存数据.
	AllUser   map[int64]*UserData        //所有的用户信息.
	AllGroup  map[int64]*GroupData       //所有的组信息.
	AllTagIdx map[string]int64           //所有的tag最新序号.
	PosCache  map[int64]map[string]int64 //用户上传过来的"哪个tag接收到哪个位置了"的信息.
}

const ROOT_USER_ID int64 = 1

func new_InnerCacheData() *InnerCacheData {
	curData := new(InnerCacheData)
	//
	curData.AllUser = make(map[int64]*UserData)
	curData.AllGroup = make(map[int64]*GroupData)
	curData.AllTagIdx = make(map[string]int64)
	//
	if true {
		rootUser := new(UserData)
		rootUser.Id = ROOT_USER_ID
		rootUser.Alias = "root"
		rootUser.Password = "toor"
		curData.AllUser[rootUser.Id] = rootUser
		//
		curData.AllTagIdx[TagName_User_Push(ROOT_USER_ID)] = 0
	}
	//
	return curData
}

func (self *InnerCacheData) checkFriend(uId1 int64, uId2 int64) error {
	var err error
	var ok bool
	var ud1 *UserData
	var ud2 *UserData

	for _ = range "1" {

		if ud1, ok = self.AllUser[uId1]; !ok {
			err = fmt.Errorf("找不到userId=%v的用户", uId1)
			break
		}

		if ud2, ok = self.AllUser[uId2]; !ok {
			err = fmt.Errorf("找不到userId=%v的用户", uId2)
			break
		}

		if !SliceInt64_isIn(ud1.Friends, ud2.Id) {
			err = fmt.Errorf("userId=%v不是userId=%v的好友", ud2.Id, ud1.Id)
			break
		}

		if !SliceInt64_isIn(ud2.Friends, ud1.Id) {
			err = fmt.Errorf("userId=%v不是userId=%v的好友", ud1.Id, ud2.Id)
			break
		}

		tableName := TagName_User_Chat(ud1.Id, ud2.Id)

		if _, ok = self.AllTagIdx[tableName]; !ok {
			err = fmt.Errorf("找不到%v", tableName)
			break
		}
	}
	return err
}

func (self *InnerCacheData) checkGroupMember(uId int64, gId int64) error {
	var err error
	var ok bool
	var ud *UserData
	var gd *GroupData

	for _ = range "1" {

		if ud, ok = self.AllUser[uId]; !ok {
			err = fmt.Errorf("找不到userId=%v的用户", uId)
			break
		}

		if gd, ok = self.AllGroup[gId]; !ok {
			err = fmt.Errorf("找不到groupId=%v的组", gId)
			break
		}

		if !SliceInt64_isIn(ud.Groups, gId) {
			err = fmt.Errorf("userId=%v不是groupId=%v的成员_1", ud.Id, gd.Id)
			break
		}

		if gd.SuperId != ud.Id && !SliceInt64_isIn(gd.AdminId, ud.Id) && !SliceInt64_isIn(gd.OtherMemId, ud.Id) {
			err = fmt.Errorf("userId=%v不是groupId=%v的成员_2", ud.Id, gd.Id)
			break
		}

		tableName := TagName_Group_Chat(gd.Id)

		if _, ok = self.AllTagIdx[tableName]; !ok {
			err = fmt.Errorf("找不到%v", tableName)
			break
		}
	}
	return err
}

func (self *InnerCacheData) checkUser(uId int64) error {
	var err error
	var ok bool
	var ud *UserData

	for _ = range "1" {

		if ud, ok = self.AllUser[uId]; !ok {
			err = fmt.Errorf("找不到userId=%v的用户", uId)
			break
		}

		if uId <= 0 || uId != ud.Id {
			err = fmt.Errorf("数据异常,uId=%v,udId=%v", uId, ud.Id)
			break
		}

		for _, fId := range ud.Friends {
			if err = self.checkFriend(ud.Id, fId); err != nil {
				break
			}
		}

		for _, gId := range ud.Groups {
			if err = self.checkGroupMember(ud.Id, gId); err != nil {
				break
			}
		}
	}
	return err
}

func (self *InnerCacheData) checkGroup(gId int64) error {
	var err error
	var ok bool
	var gd *GroupData

	for _ = range "1" {

		if gd, ok = self.AllGroup[gId]; !ok {
			err = fmt.Errorf("找不到groupId=%v的组", gId)
			break
		}

		if gId <= 0 || gId != gd.Id {
			err = fmt.Errorf("数据异常,gId=%v,gdId=%v", gId, gd.Id)
			break
		}

		allMember := make(map[int64]bool)
		allMember[gd.SuperId] = true
		for _, aId := range gd.AdminId {
			allMember[aId] = true
		}
		for _, mId := range gd.OtherMemId {
			allMember[mId] = true
		}
		if len(allMember) != 1+len(gd.AdminId)+len(gd.OtherMemId) {
			err = fmt.Errorf("数据异常:groupId=%v的成员数据有重复", gId)
			break
		}

		ok = true
		for uId, _ := range allMember {
			if uId <= 0 {
				err = fmt.Errorf("数据异常:groupId=%v的userId=%v异常", gId, uId)
				ok = false
				break
			}
			if err = self.checkGroupMember(uId, gId); err != nil {
				ok = false
				break
			}
		}
		if !ok {
			break
		}

	}
	return err
}

func (self *InnerCacheData) checkTag() error {
	var err error

	userPushNum := 0
	groupChatNum := 0
	groupPushNum := 0

	for tagName, tagIdx := range self.AllTagIdx {
		if tagIdx < 0 {
			err = errors.New("序号有问题")
			return err
		}
		if data, ok := ParseTagName(tagName); !ok {
			err = fmt.Errorf("未知的%v", tagName)
			return err
		} else {
			if data.IsUserChat {
				if err = self.checkFriend(data.ChatUserId1, data.ChatUserId2); err != nil {
					return err
				}
			} else if data.IsUserPush {
				if err = self.checkUser(data.PushUserId); err != nil {
					return err
				}
				userPushNum += 1
			} else if data.IsGroupChat {
				if err = self.checkGroup(data.ChatGroupId); err != nil {
					return err
				}
				groupChatNum += 1
			} else if data.IsGroupPush {
				if err = self.checkGroup(data.PushGroupId); err != nil {
					return err
				}
				groupPushNum += 1
			} else {
				err = errors.New("进入未知逻辑")
				return err
			}
		}
	}

	if len(self.AllUser) != userPushNum {
		err = errors.New("tag个数和user个数不符")
		return err
	}
	if groupChatNum != groupPushNum || groupChatNum != len(self.AllGroup) {
		err = errors.New("tag个数和group个数不符")
		return err
	}
	return nil
}

func (self *InnerCacheData) check() error {
	//用户的别名唯一且不为空.
	//组的别名唯一且不为空.
	var err error
	allUserAlias := make(map[string]bool)
	for _, ud := range self.AllUser {
		if len(ud.Alias) == 0 {
			err = errors.New("数据异常:存在用户的别名为空")
			return err
		}
		allUserAlias[ud.Alias] = true
	}
	if len(allUserAlias) != len(self.AllUser) {
		err = errors.New("数据异常:用户的别名有重复")
		return err
	}

	allGroupAlias := make(map[string]bool)
	for _, gd := range self.AllGroup {
		if len(gd.Alias) == 0 {
			err = errors.New("数据异常:存在组的别名为空")
			return err
		}
		allGroupAlias[gd.Alias] = true
	}
	if len(allGroupAlias) != len(self.AllGroup) {
		err = errors.New("数据异常:组的别名有重复")
		return err
	}

	if _, ok := self.AllUser[ROOT_USER_ID]; !ok {
		err = fmt.Errorf("数据异常,没有userId=%v的用户", ROOT_USER_ID)
		return err
	}

	for _, ud := range self.AllUser {
		if err = self.checkUser(ud.Id); err != nil {
			return err
		}
	}

	for _, gd := range self.AllGroup {
		if err = self.checkGroup(gd.Id); err != nil {
			return err
		}
	}

	if err = self.checkTag(); err != nil {
		return err
	}

	return nil
}
