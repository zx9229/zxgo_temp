package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type CacheData struct {
	inner *InnerCacheData
}

func New_CacheData() *CacheData {
	curData := new(CacheData)
	curData.inner = new_InnerCacheData()
	return curData
}

func (self *CacheData) FromJson(jsonStr string) error {
	tmpInner := new(InnerCacheData)
	if err := json.Unmarshal([]byte(jsonStr), tmpInner); err != nil {
		return err
	}
	self.inner = tmpInner
	return nil
}

func (self *CacheData) ToJson() (jsonStr string, err error) {
	var jsonByte []byte
	if jsonByte, err = json.Marshal(self.inner); err != nil {
		return
	}
	jsonStr = string(jsonByte)
	return
}

func (self *CacheData) Check() error {
	return self.inner.check()
}

func (self *CacheData) AddUser(alias string, password string) (newUserId int64, err error) {
	if len(alias) == 0 {
		err = errors.New("要创建的用户的alias字符串为空")
		return
	}

	if ud, _ := self.inner.findUserAlias(alias); ud != nil {
		err = fmt.Errorf("已经存在userAlias=%v的用户", alias)
		return
	}

	var cloneInner *InnerCacheData
	if cloneInner, err = self.inner.clone(); err != nil {
		err = errors.New("执行代码出错,数据未修改")
		return
	}

	cloneInner.LastUserId += 1
	newUd := New_UserData()
	newUd.Id = cloneInner.LastUserId
	newUd.Alias = alias
	newUd.Password = password

	cloneInner.AllUser[newUd.Id] = newUd
	cloneInner.TagIdxUsable[TagName_User_Push(newUd.Id)] = 0

	if err = cloneInner.check(); err != nil {
		return
	}

	self.inner = cloneInner

	newUserId = newUd.Id

	return
}

func (self *CacheData) AddFriend(fId1 int64, fId2 int64) error {
	var err error
	var ud1 *UserData
	var ud2 *UserData

	if ud1, err = self.inner.findUserId(fId1); err != nil {
		return err
	}

	if ud2, err = self.inner.findUserId(fId2); err != nil {
		return err
	}

	if SliceInt64_isIn(ud1.Friends, ud2.Id) {
		err = errors.New("已经是好友了")
		return err
	}

	if err = self.inner.check(); err != nil { //真正操作前,先检查一下数据,这样,出问题的时候,可以知道,是"已经出问题了"还是"本次操作出现了问题".
		return err
	}

	var cloneInner *InnerCacheData
	if cloneInner, err = self.inner.clone(); err != nil {
		err = errors.New("执行代码出错,数据未修改")
		return err
	}

	ud1 = nil
	ud2 = nil
	ud1 = cloneInner.AllUser[fId1]
	ud2 = cloneInner.AllUser[fId2]
	ud1.Friends = append(ud1.Friends, ud2.Id)
	ud2.Friends = append(ud2.Friends, ud1.Id)
	tagName := TagName_User_Chat(ud1.Id, ud2.Id)
	if idx, ok := cloneInner.TagIdxUseless[tagName]; ok {
		delete(cloneInner.TagIdxUseless, tagName)
		cloneInner.TagIdxUsable[tagName] = idx
	} else {
		cloneInner.TagIdxUsable[tagName] = 0
	}
	if err = cloneInner.check(); err != nil {
		return err
	}

	self.inner = cloneInner

	return err
}
