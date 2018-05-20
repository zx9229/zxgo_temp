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
	AllUser   map[int64]*MyStruct.UserData  //所有的用户数据.
	AllGroup  map[int64]*MyStruct.GroupData //所有的组数据.
	MapRowIdx map[string]int64              //以Id递增的表,缓存了它的序号.
}

func new_InnerCacheData() *InnerCacheData {
	curData := new(InnerCacheData)
	//
	curData.AllUser = make(map[int64]*MyStruct.UserData)
	curData.AllGroup = make(map[int64]*MyStruct.GroupData)
	curData.MapRowIdx = make(map[string]int64)
	curData.MapRowIdx[TableName_PushRow()] = 0
	curData.MapRowIdx[TableName_ChatRow()] = 0
	//
	return curData
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
		err = errors.New(fmt.Sprintf("找不到id=%v的用户", uId1))
		return err
	}
	if ud2, ok = self.AllUser[uId2]; !ok {
		err = errors.New(fmt.Sprintf("找不到id=%v的用户", uId2))
		return err
	}
	curTablename := TableName_Friend(uId1, uId2)
	if curLastRowIdx, ok = self.MapRowIdx[curTablename]; !ok {
		err = errors.New(fmt.Sprintf("找不到%v的数据库表", curTablename))
		return err
	}
	if curU1RecvPosition, ok = ud1.FriendPos[uId2]; !ok {
		err = errors.New(fmt.Sprintf("userId=%v里面找不到userId=%v的position", uId1, uId2))
		return err
	}
	if curU2RecvPosition, ok = ud2.FriendPos[uId1]; !ok {
		err = errors.New(fmt.Sprintf("userId=%v里面找不到userId=%v的position", uId2, uId1))
		return err
	}
	if curLastRowIdx < 0 || curU1RecvPosition < 0 || curU2RecvPosition < 0 || curLastRowIdx < curU1RecvPosition || curLastRowIdx < curU2RecvPosition {
		err = errors.New(fmt.Sprintf("数据异常:%v有%v条数据,userId=%v接收到了第%v条,userId=%v接收到了第%v条", curTablename, curLastRowIdx, uId1, curU1RecvPosition, uId2, curU2RecvPosition))
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
		err = errors.New(fmt.Sprintf("找不到groupId=%v的组", gId))
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
		err = errors.New(fmt.Sprintf("数据异常:groupId=%v的成员数据有重复", gId))
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
		err = errors.New(fmt.Sprintf("找不到userId=%v的用户", uId))
		return err
	}
	if gd, ok = self.AllGroup[gId]; !ok {
		err = errors.New(fmt.Sprintf("找不到groupId=%v的组", gId))
		return err
	}
	curTablename := TableName_Group(gId)
	if curLastRowIdx, ok = self.MapRowIdx[curTablename]; !ok {
		err = errors.New(fmt.Sprintf("找不到名为%v的数据库表", curTablename))
		return err
	}
	if uId != gd.SuperId && !myInSlice(uId, gd.AdminId) && !myInSlice(uId, gd.OtherMemId) {
		err = errors.New(fmt.Sprintf("userId=%v的用户不在groupId=%v里面", uId, gId))
		return err
	}
	if curRecvPosition, ok = ud.GroupPos[gId]; !ok {
		err = errors.New(fmt.Sprintf("userId=%v里面找不到groupId=%v的position", uId, gId))
		return err
	}
	if curLastRowIdx < 0 || curRecvPosition < 0 || curLastRowIdx < curRecvPosition {
		err = errors.New(fmt.Sprintf("数据异常:groupId=%v有%v条数据,userId=%v接收到了第%v条", gId, curLastRowIdx, uId, curRecvPosition))
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
		err = errors.New(fmt.Sprintf("找不到%v的数据库表", TableName_PushRow()))
		return err
	}

	AllUserAlias := make(map[string]bool)
	for _, ud := range self.AllUser {
		AllUserAlias[ud.Alias] = true
	}
	if len(AllUserAlias) != len(self.AllUser) {
		err = errors.New(fmt.Sprintf("数据异常:用户的别名有重复"))
		return err
	}

	for udKey, ud := range self.AllUser {
		if udKey != ud.Id {
			err = errors.New(fmt.Sprintf("数据异常,udKey=%v,udId=%v", udKey, ud.Id))
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
			err = errors.New(fmt.Sprintf("数据异常:%v有%v条数据,userId=%v接收到了第%v条", TableName_PushRow(), curLastRowIndexNotice, ud.Id, ud.NoticePos))
			return err
		}
	}

	AllGroupAlias := make(map[string]bool)
	for _, gd := range self.AllGroup {
		AllGroupAlias[gd.Alias] = true
	}
	if len(AllGroupAlias) != len(self.AllGroup) {
		err = errors.New(fmt.Sprintf("数据异常:组的别名有重复"))
		return err
	}

	for gdKey, gd := range self.AllGroup {
		if gdKey != gd.Id {
			err = errors.New(fmt.Sprintf("数据异常,gdKey=%v,gdId=%v", gdKey, gd.Id))
			return err
		}
		if err = self.checkGroup(gd.Id); err != nil {
			return err
		}
	}
	for tbName, rowIdx := range self.MapRowIdx {
		if len(tbName) <= 0 || rowIdx < 0 {
			err = errors.New(fmt.Sprintf("数据异常,tbName=%v,rowIdx=%v", tbName, rowIdx))
			return err
		}
	}
	err = nil
	return err
}

func (self *InnerCacheData) findUser(uId *int64, uAlias *string) (ud *MyStruct.UserData, err error) {
	ud = nil
	err = nil
	if (uId == nil && uAlias == nil) || (uId != nil && uAlias != nil) {
		err = errors.New("uId和uAlias需要:有且仅有一个有效数据!")
		return
	}

	if uId != nil {
		var isOk bool = false
		if ud, isOk = self.AllUser[(*uId)]; !isOk {
			err = errors.New(fmt.Sprintf("找不到uId=%v的用户", *uId))
			return
		} else {
			return
		}
	} else if uAlias != nil {
		for _, _ud := range self.AllUser {
			if _ud.Alias == *uAlias {
				ud = _ud
				return
			}
		}
		err = errors.New(fmt.Sprintf("找不到uAlias=%v的用户", *uAlias))
		return
	} else {
		panic("程序进入了无法到达的逻辑!")
	}
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

func (self *InnerCacheData) addUser(alias string, password string) error {
	var err error
	if _, err = self.findUser(nil, &alias); err == nil {
		err = errors.New(fmt.Sprintf("已经存在alias=%v的用户", alias))
		return err
	}

	newUd := MyStruct.InnerNewUserData()
	newUd.Id = self.calcMaxUserId() + 1
	newUd.Alias = alias
	newUd.Password = password

	//TODO:真正修改之前,预修改&检查一下,通过之后,再真正的修改.
	self.AllUser[newUd.Id] = newUd
	if err = self.check(); err != nil {
		panic(err)
	}

	return err
}

func (self *InnerCacheData) AddFriends(fId1 int64, fId2 int64) error {
	var err error
	var ud1 *MyStruct.UserData = nil
	var ud2 *MyStruct.UserData = nil
	if ud1, err = self.findUser(&fId1, nil); err != nil {
		return err
	}
	if ud2, err = self.findUser(&fId2, nil); err != nil {
		return err
	}
	if err = self.check(); err != nil { //真正操作前,先检查一下数据,这样,出问题的时候,可以知道,是"已经出问题了"还是"本次操作出现了问题".
		return err
	}
	if _, ok := ud1.FriendPos[ud2.Id]; ok {
		err = errors.New("已经是好友了")
		return err
	} else {
		ud1.FriendPos[ud2.Id] = 0
		ud2.FriendPos[ud1.Id] = 0
		self.MapRowIdx[TableName_Friend(ud1.Id, ud2.Id)] = 0
		if err = self.check(); err != nil {
			panic(err)
		}
	}
	return err
}
