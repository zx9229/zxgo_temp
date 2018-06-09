package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type InnerCacheData struct { //内存中的缓存数据.
	AllUser       map[int64]*UserData  //所有的用户信息.
	LastUserId    int64                //最后一个创建的用户ID(允许删除用户,删除之后,这个ID就不能再用了,所以需要一个字段维护数据).
	AllGroup      map[int64]*GroupData //所有的组信息.
	LastGroupId   int64                //最后一个创建的组ID.
	TagIdxUsable  map[string]int64     //tag的最新序号.
	TagIdxUseless map[string]int64     //解除好友关系的tag移到这里来.
}

const ROOT_USER_ID int64 = 1

func new_InnerCacheData() *InnerCacheData {
	curData := new(InnerCacheData)
	//
	curData.AllUser = make(map[int64]*UserData)
	curData.AllGroup = make(map[int64]*GroupData)
	curData.TagIdxUsable = make(map[string]int64)
	curData.TagIdxUseless = make(map[string]int64)
	//
	if true {
		rootUser := New_UserData()
		rootUser.Id = ROOT_USER_ID
		rootUser.Alias = "root"
		rootUser.Password = "toor"
		//
		curData.AllUser[rootUser.Id] = rootUser
		curData.LastUserId = rootUser.Id
		curData.TagIdxUsable[TagName_User_Push(ROOT_USER_ID)] = 0
	}
	//
	return curData
}

func (self *InnerCacheData) clone() (cloneObj *InnerCacheData, err error) {
	var jsonByte []byte
	if jsonByte, err = json.Marshal(self); err != nil {
		return
	}
	cloneObj = new(InnerCacheData)
	if err = json.Unmarshal(jsonByte, cloneObj); err != nil {
		cloneObj = nil
		return
	}
	return
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

		if _, ok = self.TagIdxUsable[tableName]; !ok {
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

		if _, ok = self.TagIdxUsable[tableName]; !ok {
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
	//受限于这个设计,A和B成为好友了,然后互发了3条消息,此时u_A_B_=3,
	//然后A和B解除好友了,此时不删除u_A_B_,
	//当他们再次添加好友时,数据从3开始继续计数,
	//否则,当他们再次添加好友时,u_A_B_的计数器会混乱.

	var err error

	userPushNum := 0
	groupChatNum := 0
	groupPushNum := 0

	for tagName, tagIdx := range self.TagIdxUsable {
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

	for tagName, tagIdx := range self.TagIdxUseless {
		if tagIdx < 0 {
			err = errors.New("序号有问题")
			return err
		}
		if data, ok := ParseTagName(tagName); !ok {
			err = fmt.Errorf("未知的%v", tagName)
			return err
		} else {
			if data.IsUserChat {
				if err = self.checkUser(data.ChatUserId1); err != nil {
					return err
				}
				if err = self.checkUser(data.ChatUserId2); err != nil {
					return err
				}
			} else {
				err = errors.New("数据异常")
				return err
			}
		}
	}

	return nil
}

func (self *InnerCacheData) checkUserAlias() error {
	//用户的别名唯一且不为空.
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

	return err
}

func (self *InnerCacheData) checkGroupAlias() error {
	//组的别名唯一且不为空.
	var err error
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

	return err
}

func (self *InnerCacheData) checkLast_U_G_Id() error {
	var err error
	var maxId int64

	maxId = 0
	for _, ud := range self.AllUser {
		if maxId < ud.Id {
			maxId = ud.Id
		}
	}
	if self.LastUserId < maxId {
		err = errors.New("数据异常")
		return err
	}

	maxId = 0
	for _, gd := range self.AllGroup {
		if maxId < gd.Id {
			maxId = gd.Id
		}
	}
	if self.LastGroupId < maxId {
		err = errors.New("数据异常")
		return err
	}

	return err
}

func (self *InnerCacheData) check() error {
	var err error

	if err = self.checkUserAlias(); err != nil {
		return err
	}

	if err = self.checkGroupAlias(); err != nil {
		return err
	}

	if err = self.checkLast_U_G_Id(); err != nil {
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

func (self *InnerCacheData) findUserId(uId int64) (ud *UserData, err error) {
	if _ud, isOk := self.AllUser[uId]; !isOk {
		err = fmt.Errorf("找不到userId=%v的用户", uId)
	} else {
		ud = _ud
	}
	return
}

func (self *InnerCacheData) findUserAlias(uAlias string) (ud *UserData, err error) {
	for _, _ud := range self.AllUser {
		if _ud.Alias == uAlias {
			ud = _ud
			return
		}
	}
	err = fmt.Errorf("找不到userAlias=%v的用户", uAlias)
	return
}
