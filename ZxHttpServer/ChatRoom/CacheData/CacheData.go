package CacheData

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/go-xorm/xorm"
	"github.com/zx9229/zxgo_temp/ZxHttpServer/ChatRoom/MyStruct"
)

func calcTablenameXorm(engine *xorm.Engine, bean interface{}) string {
	//我参考的代码 func (engine *Engine) tbName(v reflect.Value) string {
	var v reflect.Value = reflect.Indirect(reflect.ValueOf(bean))
	tbName := engine.TableMapper.Obj2Table(reflect.Indirect(v).Type().Name())
	return tbName
}

func TableName_PushRow() string {
	return "t_push_row"
}

func TableName_ChatRow() string {
	return "t_chat_row"
}

func TableName_Group(groupId int64) string {
	return fmt.Sprintf("t_g_%v", groupId)
}

func TableName_Friend(userId1 int64, userId2 int64) string {
	var minVal, maxVal int64
	if userId1 < userId2 {
		minVal = userId1
		maxVal = userId2
	} else {
		minVal = userId2
		maxVal = userId1
	}
	return fmt.Sprintf("t_f_%v_%v", minVal, maxVal)
}

func myInSlice(dataItem int64, dataSlice []int64) bool {
	if dataSlice != nil {
		for _, element := range dataSlice {
			if dataItem == element {
				return true
			}
		}
	}
	return false
}

type InnerCacheData struct { //内存中的缓存数据.
	RootUser  *MyStruct.UserData
	AllUser   map[int64]*MyStruct.UserData  //所有的用户数据.
	AllGroup  map[int64]*MyStruct.GroupData //所有的组数据.
	MapRowIdx map[string]int64              //以Id递增的表,缓存了它的序号.
}

func new_InnerCacheData() *InnerCacheData {
	curData := new(InnerCacheData)
	//
	curData.RootUser = new(MyStruct.UserData)
	curData.RootUser.Id = 0
	curData.RootUser.Password = "123"
	curData.AllUser = make(map[int64]*MyStruct.UserData)
	curData.AllGroup = make(map[int64]*MyStruct.GroupData)
	curData.MapRowIdx = make(map[string]int64)
	curData.MapRowIdx[TableName_PushRow()] = 0
	curData.MapRowIdx[TableName_ChatRow()] = 0
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

func (self *InnerCacheData) checkFriends(uId1 int64, uId2 int64) error {
	var err error
	var ok bool = false
	var ud1 *MyStruct.UserData = nil
	var ud2 *MyStruct.UserData = nil
	var curLastRowIdx int64 = -1     //数据库表中,当前时刻,最新的那一行,的序号.
	var curU1RecvPosition int64 = -1 //数据库表中,当前时刻,用户已经接收了哪个序号的数据.
	var curU2RecvPosition int64 = -1 //数据库表中,当前时刻,用户已经接收了哪个序号的数据.

	if ud1, ok = self.AllUser[uId1]; !ok {
		err = fmt.Errorf("找不到userId=%v的用户", uId1)
		return err
	}
	if ud2, ok = self.AllUser[uId2]; !ok {
		err = fmt.Errorf("找不到userId=%v的用户", uId2)
		return err
	}
	curTablename := TableName_Friend(uId1, uId2)
	if curLastRowIdx, ok = self.MapRowIdx[curTablename]; !ok {
		err = fmt.Errorf("找不到%v的数据库表", curTablename)
		return err
	}
	if curU1RecvPosition, ok = ud1.FriendPos[uId2]; !ok {
		err = fmt.Errorf("userId=%v里面找不到userId=%v的position", uId1, uId2)
		return err
	}
	if curU2RecvPosition, ok = ud2.FriendPos[uId1]; !ok {
		err = fmt.Errorf("userId=%v里面找不到userId=%v的position", uId2, uId1)
		return err
	}
	if curLastRowIdx < 0 || curU1RecvPosition < 0 || curU2RecvPosition < 0 || curLastRowIdx < curU1RecvPosition || curLastRowIdx < curU2RecvPosition {
		err = fmt.Errorf("数据异常:表%v有%v条数据,userId=%v接收到了第%v条,userId=%v接收到了第%v条", curTablename, curLastRowIdx, uId1, curU1RecvPosition, uId2, curU2RecvPosition)
		return err
	}
	err = nil
	return err
}

func (self *InnerCacheData) checkGroup(gId int64) error {
	var err error
	var ok bool
	var gd *MyStruct.GroupData = nil

	if gd, ok = self.AllGroup[gId]; !ok {
		err = fmt.Errorf("找不到groupId=%v的组", gId)
		return err
	}

	allMembers := make(map[int64]bool)
	allMembers[gd.SuperId] = true
	for _, aId := range gd.AdminId {
		allMembers[aId] = true
	}
	for _, mId := range gd.OtherMemId {
		allMembers[mId] = true
	}

	if len(allMembers) != (len(gd.AdminId) + len(gd.OtherMemId) + 1) {
		err = fmt.Errorf("数据异常:groupId=%v的成员数据有重复", gId)
		return err
	}

	for uId, _ := range allMembers {
		if err = self.checkGroupMember(uId, gId); err != nil {
			return err
		}
	}

	err = nil
	return err
}

func (self *InnerCacheData) checkGroupMember(uId int64, gId int64) error {
	var err error
	var ok bool
	var ud *MyStruct.UserData = nil
	var gd *MyStruct.GroupData = nil
	var curLastRowIdx int64 = -1   //数据库表中,当前时刻,最新的那一行,的序号.
	var curRecvPosition int64 = -1 //数据库表中,当前时刻,用户已经接收了哪个序号的数据.

	if ud, ok = self.AllUser[uId]; !ok {
		err = fmt.Errorf("找不到userId=%v的用户", uId)
		return err
	}
	if gd, ok = self.AllGroup[gId]; !ok {
		err = fmt.Errorf("找不到groupId=%v的组", gId)
		return err
	}
	curTablename := TableName_Group(gId)
	if curLastRowIdx, ok = self.MapRowIdx[curTablename]; !ok {
		err = fmt.Errorf("找不到名为%v的数据库表", curTablename)
		return err
	}
	if uId != gd.SuperId && !myInSlice(uId, gd.AdminId) && !myInSlice(uId, gd.OtherMemId) {
		err = fmt.Errorf("userId=%v的用户不在groupId=%v里面", uId, gId)
		return err
	}
	if curRecvPosition, ok = ud.GroupPos[gId]; !ok {
		err = fmt.Errorf("userId=%v里面找不到groupId=%v的position", uId, gId)
		return err
	}
	if curLastRowIdx < 0 || curRecvPosition < 0 || curLastRowIdx < curRecvPosition {
		err = fmt.Errorf("数据异常:groupId=%v有%v条数据,userId=%v接收到了第%v条", gId, curLastRowIdx, uId, curRecvPosition)
		return err
	}
	err = nil
	return err
}

func (self *InnerCacheData) check() error {
	var err error

	var curLastRowIndexNotice int64
	var ok bool
	if curLastRowIndexNotice, ok = self.MapRowIdx[TableName_PushRow()]; !ok {
		err = fmt.Errorf("找不到%v的数据库表", TableName_PushRow())
		return err
	}

	AllUserAlias := make(map[string]bool)
	for _, ud := range self.AllUser {
		AllUserAlias[ud.Alias] = true
	}
	if len(AllUserAlias) != len(self.AllUser) {
		err = errors.New("数据异常:用户的别名有重复")
		return err
	}

	for udKey, ud := range self.AllUser {
		if udKey != ud.Id {
			err = fmt.Errorf("数据异常,udKey=%v,udId=%v", udKey, ud.Id)
			return err
		}
		for gId, _ := range ud.GroupPos {
			if err = self.checkGroupMember(udKey, gId); err != nil {
				return err
			}
		}
		for fId, _ := range ud.FriendPos {
			if err = self.checkFriends(ud.Id, fId); err != nil {
				return err
			}
		}
		if ud.NoticePos < 0 || curLastRowIndexNotice < 0 || curLastRowIndexNotice < ud.NoticePos {
			err = fmt.Errorf("数据异常:表%v有%v条数据,userId=%v接收到了第%v条", TableName_PushRow(), curLastRowIndexNotice, ud.Id, ud.NoticePos)
			return err
		}
	}

	AllGroupAlias := make(map[string]bool)
	for _, gd := range self.AllGroup {
		AllGroupAlias[gd.Alias] = true
	}
	if len(AllGroupAlias) != len(self.AllGroup) {
		err = errors.New("数据异常:组的别名有重复")
		return err
	}

	for gdKey, gd := range self.AllGroup {
		if gdKey != gd.Id {
			err = fmt.Errorf("数据异常:gdKey=%v,gdId=%v", gdKey, gd.Id)
			return err
		}
		if err = self.checkGroup(gd.Id); err != nil {
			return err
		}
	}
	for tbName, rowIdx := range self.MapRowIdx {
		if len(tbName) <= 0 || rowIdx < 0 {
			err = fmt.Errorf("数据异常:tbName=%v,rowIdx=%v", tbName, rowIdx)
			return err
		}
	}
	err = nil
	return err
}

func (self *InnerCacheData) findUserId(uId int64) (ud *MyStruct.UserData, err error) {
	if _ud, isOk := self.AllUser[uId]; !isOk {
		err = fmt.Errorf("找不到userId=%v的用户", uId)
	} else {
		ud = _ud
	}
	return
}

func (self *InnerCacheData) findUserAlias(uAlias string) (ud *MyStruct.UserData, err error) {
	for _, _ud := range self.AllUser {
		if _ud.Alias == uAlias {
			ud = _ud
			return
		}
	}
	err = fmt.Errorf("找不到userAlias=%v的用户", uAlias)
	return
}

func (self *InnerCacheData) calcMaxUserId() int64 {
	var maxId int64 = 0
	for _, ud := range self.AllUser {
		if maxId < ud.Id {
			maxId = ud.Id
		}
	}
	return maxId
}

func (self *CacheData) AddUser(alias string, password string) (newUserId int64, err error) {
	if len(alias) == 0 {
		err = errors.New("要创建的用户的alias字符串为空")
		return
	}

	if _, err = self.inner.findUserAlias(alias); err == nil {
		err = fmt.Errorf("已经存在userAlias=%v的用户", alias)
		return
	}

	newUd := MyStruct.New_UserData()
	newUd.Id = self.inner.calcMaxUserId() + 1
	newUd.Alias = alias
	newUd.Password = password

	var cloneInner *InnerCacheData
	if cloneInner, err = self.inner.clone(); err != nil {
		err = errors.New("执行代码出错,数据未修改")
		return
	}

	cloneInner.AllUser[newUd.Id] = newUd
	if err = cloneInner.check(); err != nil {
		return
	}

	self.inner = cloneInner

	newUserId = newUd.Id
	return
}

func (self *CacheData) AddFriends(fId1 int64, fId2 int64) error {
	var err error
	var ud1 *MyStruct.UserData = nil
	var ud2 *MyStruct.UserData = nil
	if ud1, err = self.inner.findUserId(fId1); err != nil {
		return err
	}
	if ud2, err = self.inner.findUserId(fId2); err != nil {
		return err
	}
	if err = self.inner.check(); err != nil { //真正操作前,先检查一下数据,这样,出问题的时候,可以知道,是"已经出问题了"还是"本次操作出现了问题".
		return err
	}
	if _, ok := ud1.FriendPos[ud2.Id]; ok {
		err = errors.New("已经是好友了")
		return err
	}

	var cloneInner *InnerCacheData
	if cloneInner, err = self.inner.clone(); err != nil {
		err = errors.New("执行代码出错,数据未修改")
		return err
	}

	cloneInner.AllUser[fId1].FriendPos[fId2] = 0
	cloneInner.AllUser[fId2].FriendPos[fId1] = 0
	cloneInner.MapRowIdx[TableName_Friend(fId1, fId2)] = 0
	if err = cloneInner.check(); err != nil {
		return err
	}

	self.inner = cloneInner

	return err
}

func (self *CacheData) CheckPassword(uId int64, uAlias string, password string) (userId int64, err error) {
	var ud *MyStruct.UserData
	if 0 < uId {
		if ud, err = self.inner.findUserId(uId); err != nil {
			return
		}
	} else {
		if ud, err = self.inner.findUserAlias(uAlias); err != nil {
			return
		}
	}

	if ud.Password == password {
		userId = ud.Id
		return
	} else {
		err = errors.New("密码错误")
		return
	}
}
