package main

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

var tnPushMessageRaw string = "" //TODO:待初始化.
var tnPushMessage string = ""    //TODO:待初始化.
var tnChatMessageRaw string = "" //TODO:待初始化.

type MyChat struct {
	engine     *xorm.Engine
	rwMutex    *sync.Mutex
	allUser    map[int64]UserData
	allGroup   map[int64]GroupData
	allKV      map[string]KeyValue
	mapRowIdx  map[string]int64     //以Id递增的表,缓存了它的序号.
	chanPmr    chan *PushMessageRaw //存数据库专用
	chanPm     chan *PushMessage    //存数据库专用
	chanCmr    chan *ChatMessageRaw //存数据库专用
	chanCm     chan *ChatMessage    //存数据库专用
	mapSession map[*websocket.Conn]string
}

func NewMyChat() *MyChat {
	newData := new(MyChat)

	newData.engine = nil
	newData.rwMutex = new(sync.Mutex)
	newData.allUser = make(map[int64]UserData)
	newData.allGroup = make(map[int64]GroupData)
	newData.allKV = make(map[string]KeyValue)
	newData.mapRowIdx = make(map[string]int64)
	newData.chanPmr = make(chan *PushMessageRaw, 1024)
	newData.chanPm = make(chan *PushMessage, 1024)
	newData.chanCmr = make(chan *ChatMessageRaw, 1024)
	newData.chanCm = make(chan *ChatMessage, 1024)
	newData.mapSession = make(map[*websocket.Conn]string)

	return newData
}

func (self *MyChat) Engine() *xorm.Engine {
	return self.engine
}

func (self *MyChat) Init(driverName string, dataSourceName string, locationName string) error {
	//[非]线程安全.
	var err error = nil

	for _ = range "1" {
		var location *time.Location = nil
		if len(locationName) > 0 {
			if location, err = time.LoadLocation(locationName); err != nil {
				break
			}
		}

		if self.engine != nil {
			err = errors.New("数据库引擎engine已经被初始化过了!")
			break
		}
		if self.engine, err = xorm.NewEngine(driverName, dataSourceName); err != nil {
			break
		}

		if location != nil {
			self.engine.TZLocation = location
		}

		self.engine.SetMapper(core.GonicMapper{}) //支持struct为驼峰式命名,表结构为下划线命名之间的转换,同时对于特定词支持更好.

		if err = self.createTablesAndSync2(); err != nil {
			break
		}

		if err = self.loadDataFromDbWithLock(); err != nil {
			break
		}

		self.mapRowIdx = self.mapRowIdx //TODO:初始化.
	}

	return err
}

func (self *MyChat) AddUserWithLock(alias string, password string) error {
	var err error = nil

	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	for _ = range "1" {

		if _, err = self.findUserWithoutLock(nil, &alias); err == nil { //找到了这个用户别名
			err = errors.New("用户别名已存在")
			break
		} else {
			err = nil
		}

		if err = self.saveDataToDbWithoutLock(); err != nil {
			break
		}

		ud := UserData{Alias: alias, Password: password}

		if err = self.InsertOne(ud); err != nil {
			break
		}

		if err = self.loadDataFromDbWithoutLock(); err != nil {
			panic(err) //插入数据之后,程序无法刷新数据,此时状态已不可挽回.
		}
	}

	return err
}

func (self *MyChat) ModifyUserWithLock(id *int64, alias *string, newAlias *string, newPwd *string) error {
	var err error = nil

	if newAlias == nil && newPwd == nil {
		err = errors.New("newAlias和newPwd需要:至少存在一个有效数据!")
		return err
	}

	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	for _ = range "1" {
		if err = self.saveDataToDbWithoutLock(); err != nil {
			break
		}

		var ud UserData
		if ud, err = self.findUserWithoutLock(id, alias); err != nil {
			break
		}

		if newAlias != nil {
			ud.Alias = *newAlias
		}
		if newPwd != nil {
			ud.Password = *newPwd
		}

		if err = self.UpdateOnce(ud); err != nil {
			break
		}

		if err = self.loadDataFromDbWithoutLock(); err != nil {
			panic(err) //更新数据之后,程序无法刷新数据,此时状态已不可挽回.
		}
	}

	return err
}

func (self *MyChat) AddGroupWithLock(alias string, superId int64) error {
	var err error = nil

	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	for _ = range "1" {
		if _, err = self.findGroupWithoutLock(nil, &alias); err == nil { //找到了这个组的别名
			err = errors.New("组的别名已存在")
			break
		} else {
			err = nil
		}

		if err = self.saveDataToDbWithoutLock(); err != nil {
			break
		}

		//判断superId是否存在
		if _, err = self.findUserWithoutLock(&superId, nil); err != nil {
			break
		}

		gd := GroupData{Alias: alias, SuperId: superId}

		if err = self.InsertOne(gd); err != nil {
			break
		}

		if err = self.loadDataFromDbWithoutLock(); err != nil {
			panic(err) //插入数据之后,程序无法刷新数据,此时状态已不可挽回.
		}
	}

	return err
}

func (self *MyChat) ModifyGroupBasicData(id *int64, alias *string, newAlias *string, newSuperId *int64, newAdminId []int64) error {
	var err error = nil

	if newAlias == nil && newSuperId == nil {
		err = errors.New("newAlias和newPwd需要:至少存在一个有效数据!")
		return err
	}

	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	for _ = range "1" {
		if err = self.saveDataToDbWithoutLock(); err != nil {
			break
		}

		if newSuperId != nil { //判断newSuperId是否存在
			if _, err = self.findUserWithoutLock(newSuperId, nil); err != nil {
				break
			}
		}

		if newAdminId != nil { //判断newAdminId是否存在
			for _, adminId := range newAdminId {
				if _, err = self.findUserWithoutLock(&adminId, nil); err != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}

		var gd GroupData
		if gd, err = self.findGroupWithoutLock(id, alias); err != nil {
			break
		}

		if newAlias != nil {
			gd.Alias = *newAlias
		}
		if newSuperId != nil {
			gd.SuperId = *newSuperId
		}
		if newAdminId != nil {
			gd.AdminId = newAdminId
		}

		if err = self.UpdateOnce(gd); err != nil {
			break
		}

		if err = self.loadDataFromDbWithoutLock(); err != nil {
			panic(err) //更新数据之后,程序无法刷新数据,此时状态已不可挽回.
		}
	}

	return err
}

func (self *MyChat) SetGroupMembersWithLock(gId *int64, gAlias *string, memid []int64, memAlias []string) error {
	var err error = nil

	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	for _ = range "1" {
		if err = self.saveDataToDbWithoutLock(); err != nil {
			break
		}

		var gd GroupData
		if gd, err = self.findGroupWithoutLock(gId, gAlias); err != nil {
			break
		}

		var memberSlice []int64 = nil
		if memberSlice, _, err = self.findAndMergeUserWithoutLock(memid, memAlias); err != nil {
			break
		}

		var tName string
		if tName, err = calcChatTableOfGroup(gd.Id); err != nil {
			break
		}

		var lastData ChatMessage //TODO:如果这个组,刚刚创建完,尚未有任何人发言的话,是查不到数据的.
		if lastData, err = self.queryLastChatMessage(tName); err != nil {
			break
		}

		for _, ud := range self.allUser {
			var isOk bool = false
			if myInSlice(ud.Id, memberSlice) {
				if _, isOk = ud.GroupPos[gd.Id]; !isOk {
					ud.GroupPos[gd.Id] = lastData.Id
				} //如果找到了,说明这个用户原来就在这个组里面,此时不能修改pos的.
			} else {
				if _, isOk = ud.GroupPos[gd.Id]; isOk {
					delete(ud.GroupPos, gd.Id)
				}
			}
		}

		if err = self.saveDataToDbWithoutLock(); err != nil {
			//如果操作(保存)失败了,尝试进行回滚,如果回滚也失败了,那就悲催了.
			if err1 := self.loadDataFromDbWithoutLock(); err1 != nil {
				panic(err1)
			}
			break
		}
	}

	return err
}

func (self *MyChat) AddGroupMemberWithLock(gId *int64, gAlias *string, uId *int64, uAlias *string) error {
	var err error = nil

	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	for _ = range "1" {
		if err = self.saveDataToDbWithoutLock(); err != nil {
			break
		}

		var gd GroupData
		if gd, err = self.findGroupWithoutLock(gId, gAlias); err != nil {
			break
		}

		var ud UserData
		if ud, err = self.findUserWithoutLock(uId, uAlias); err != nil {
			break
		}

		var tName string
		if tName, err = calcChatTableOfGroup(gd.Id); err != nil {
			break
		}

		var lastData ChatMessage //TODO:如果这个组,刚刚创建完,尚未有任何人发言的话,是查不到数据的.
		if lastData, err = self.queryLastChatMessage(tName); err != nil {
			break
		}

		var isOk bool = false
		if _, isOk = ud.GroupPos[gd.Id]; !isOk {
			ud.GroupPos[gd.Id] = lastData.Id
			self.allUser[ud.Id] = ud
		} else {
			err = errors.New("此组已经存在此用户,无法再次添加")
			break
		}

		if err = self.saveDataToDbWithoutLock(); err != nil {
			//如果操作(保存)失败了,尝试进行回滚,如果回滚也失败了,那就悲催了.
			if err1 := self.loadDataFromDbWithoutLock(); err1 != nil {
				panic(err1)
			}
			return err
		}
	}

	return err
}

func (self *MyChat) DelGroupMemberWithLock(gId *int64, gAlias *string, uId *int64, uAlias *string) error {
	var err error = nil

	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	for _ = range "1" {
		if err = self.saveDataToDbWithoutLock(); err != nil {
			break
		}

		var gd GroupData
		if gd, err = self.findGroupWithoutLock(gId, gAlias); err != nil {
			break
		}

		var ud UserData
		if ud, err = self.findUserWithoutLock(uId, uAlias); err != nil {
			break
		}

		var isOk bool = false
		if _, isOk = ud.GroupPos[gd.Id]; isOk {
			delete(ud.GroupPos, gd.Id)
		} else {
			err = errors.New("此组不存在此用户,无法删除")
			break
		}

		if err = self.saveDataToDbWithoutLock(); err != nil {
			//如果操作(保存)失败了,尝试进行回滚,如果回滚也失败了,那就悲催了.
			if err1 := self.loadDataFromDbWithoutLock(); err1 != nil {
				panic(err1)
			}
			break
		}
	}

	return err
}

func (self *MyChat) RecvPushMessageRaw(pushDataRaw *PushMessageRaw) error {
	//带出去了Id字段的值,再没有修改其他字段.
	//传进来参数的时候,请传过来new(PushMessageRaw)得到的指针,如果(&PushMessageRaw)我不知道有没有问题.
	var err error = nil
	var pushData *PushMessage = nil

	for _ = range "1" {
		var newUserI []int64
		var newUserA []string
		if myInSlice(0, pushDataRaw.RecverId) {
			newUserI = []int64{0}
			newUserA = nil
		} else {
			if newUserI, newUserA, err = self.findAndMergeUserWithoutLock(pushDataRaw.RecverId, pushDataRaw.RecverAlias); err != nil {
				break
			}
		}

		var newGroupI []int64
		var newGroupA []string
		if myInSlice(0, pushDataRaw.GroupId) {
			newGroupI = []int64{0}
			newGroupA = nil
		} else {
			if newGroupI, newGroupA, err = self.findAndMergeGroupWithoutLock(pushDataRaw.GroupId, pushDataRaw.GroupAlias); err != nil {
				break
			}
		}

		pushData = toPushMessage(pushDataRaw)
		pushData.RecverId = newUserI
		pushData.RecverAlias = newUserA
		pushData.GroupId = newGroupI
		pushData.GroupAlias = newGroupA
	}

	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	pushDataRaw.Id = self.innerIncrIdxWithoutLock(tnPushMessageRaw)
	if pushData != nil {
		pushData.Id = self.innerIncrIdxWithoutLock(tnPushMessage)
		pushData.IdRaw = pushDataRaw.Id
		self.chanPm <- pushData //TODO:如果写chan失败了怎么办?
	}
	self.chanPmr <- pushDataRaw //TODO:如果写chan失败了怎么办?

	self.mapSession = self.mapSession //TODO:发送给对应的socket.

	return err
}

func (self *MyChat) tryGetDataFromChan() (spmr []*PushMessageRaw, spm []*PushMessage, scmr []*ChatMessageRaw, scm []*ChatMessage) {
	var ok bool = true
	spmr = make([]*PushMessageRaw, 0)
	var pmr *PushMessageRaw
	spm = make([]*PushMessage, 0)
	var pm *PushMessage
	scmr = make([]*ChatMessageRaw, 0)
	var cmr *ChatMessageRaw
	scm = make([]*ChatMessage, 0)
	var cm *ChatMessage

	for ok {
		select {
		case pmr, ok = <-self.chanPmr:
		default:
		}
		if ok {
			spmr = append(spmr, pmr)
		}
	}

	for ok {
		select {
		case pm, ok = <-self.chanPm:
		default:
		}
		if ok {
			spm = append(spm, pm)
		}
	}

	for ok {
		select {
		case cmr, ok = <-self.chanCmr:
		default:
		}
		if ok {
			scmr = append(scmr, cmr)
		}
	}

	for ok {
		select {
		case cm, ok = <-self.chanCm:
		default:
		}
		if ok {
			scm = append(scm, cm)
		}
	}

	return
}

func (self *MyChat) innerIncrIdxWithoutLock(tablename string) int64 {
	//(非)[线程安全]
	idx, ok := self.mapRowIdx[tablename]
	if !ok {
		panic("逻辑错误")
	}
	idx += 1
	self.mapRowIdx[tablename] = idx
	return idx
}

func calcMyTablename(engine *xorm.Engine, bean interface{}) string {
	//(无所谓)[线程安全]
	//我参考的代码 func (engine *Engine) tbName(v reflect.Value) string {
	var v reflect.Value = reflect.Indirect(reflect.ValueOf(bean))
	var tbName string = engine.TableMapper.Obj2Table(reflect.Indirect(v).Type().Name())
	return tbName
}

func (self *MyChat) handlePushChatChan() {
	var err error = nil
	var ok bool = true
	var tablenamePmr string = calcMyTablename(self.engine, PushMessageRaw{})
	var tablenamePm string = calcMyTablename(self.engine, PushMessage{})
	var tablenameCmr string = calcMyTablename(self.engine, ChatMessageRaw{})
	var tablenamecm string = calcMyTablename(self.engine, ChatMessage{})
	var pmr *PushMessageRaw
	var pm *PushMessage
	var cmr *ChatMessageRaw
	var cm *ChatMessage

	for {
		pmr = nil
		pm = nil
		cmr = nil
		cm = nil

		select {
		case pmr, ok = <-self.chanPmr:
		case pm, ok = <-self.chanPm:
		case cmr, ok = <-self.chanCmr:
		case cm, ok = <-self.chanCm:
		}

		if !ok {
			panic(fmt.Sprintf("逻辑错误:%v", ok))
		}

		spmr, spm, scmr, scm := self.tryGetDataFromChan()
		if pmr != nil {
			spmr = append(spmr, pmr)
		}
		if pm != nil {
			spm = append(spm, pm)
		}
		if cmr != nil {
			scmr = append(scmr, cmr)
		}
		if cm != nil {
			scm = append(scm, cm)
		}

		tablenames := make(map[string]bool)
		for _, cm = range scm {
			tablenames[cm.MyTn] = true
		}
		if len(scmr) > 0 {
			tablenames[tablenameCmr] = true
		}
		if len(spmr) > 0 {
			tablenames[tablenamePmr] = true
		}
		if len(spm) > 0 {
			tablenames[tablenamePm] = true
		}

		if err = self.Insert(spmr); err != nil {
			log.Println(err)
		}
		if err = self.Insert(pmr); err != nil {
			log.Println(err)
		}
		if err = self.Insert(cmr); err != nil {
			log.Println(err)
		}
		if err = self.Insert(cm); err != nil {
			log.Println(err)
		}

		//TODO:对于每一个session,发送tablenames的所有key.
	}

	var pushDataRaw []*PushMessageRaw = make([]*PushMessageRaw, 0)
	var pushData []*PushMessage = make([]*PushMessage, 0)
	var chatDataRaw []*ChatMessageRaw = make([]*ChatMessageRaw, 0)
	var chatData []*ChatMessage = make([]*ChatMessage, 0)

	select {}

}

func (self *MyChat) HandlePushMessage(uId *int64, uAlias *string) error {
	var err error = nil

	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	for _ = range "1" {
		var ud UserData
		if ud, err = self.findUserWithoutLock(uId, uAlias); err != nil {
			break
		}

		var messageSlice []PushMessage = nil
		if err = self.engine.Where("id>?", ud.NoticePos).Find(&messageSlice); err != nil {
			break
		}

		if len(messageSlice) <= 0 {
			break
		}

		var message PushMessage
		for _, message = range messageSlice {
			//userId=0,表示向所有的user发送数据,groupId=0,表示向所有的group发送数据.
			if myInSlice(0, message.RecverId) || (myInSlice(0, message.GroupId) && len(ud.GroupPos) > 0) ||
				myInSlice(ud.Id, message.RecverId) || myExistKeyInSlice(ud.GroupPos, message.GroupId) {
				log.Println("通过socket发送:", message)
				//TODO:通过socket发送数据,如果发送失败,就break.
			}
		}
		ud.NoticePos = message.Id
		self.allUser[ud.Id] = ud

		if err = self.saveDataToDbWithoutLock(); err != nil {
			break
		}
	}

	return err
}

func toPushMessage(dataRaw *PushMessageRaw) *PushMessage {
	data := new(PushMessage)
	data.IdRaw = dataRaw.Id
	data.SenderId = dataRaw.SenderId
	data.SenderAlias = dataRaw.SenderAlias
	data.RecverId = dataRaw.RecverId
	data.RecverAlias = dataRaw.RecverAlias
	data.GroupId = dataRaw.GroupId
	data.GroupAlias = dataRaw.GroupAlias
	data.Message = dataRaw.Message
	data.Memo = dataRaw.Memo
	return data
}

func calcChatTableOfUser(uId1 int64, uId2 int64) (name string, err error) {
	if uId1 == uId2 || uId1 <= 0 || uId2 <= 0 {
		err = errors.New(fmt.Sprintf("入参非法,uId1=%v,uId2=%v", uId1, uId2))
		return
	}
	if uId1 > uId2 {
		uId1, uId2 = uId2, uId1
	}
	name = fmt.Sprintf("f_%v_%v", uId1, uId2)
	return
}

func calcChatTableOfGroup(gId int64) (name string, err error) {
	if gId <= 0 {
		err = errors.New(fmt.Sprintf("入参非法,gid=%v", gId))
		return
	}
	name = fmt.Sprintf("g_%v", gId)
	return
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

func myExistKeyInSlice(dataMap map[int64]int64, dataSlice []int64) bool {
	//map里面,存在key,key在slice里.
	if dataMap == nil || dataSlice == nil {
		return false
	}
	for k := range dataMap {
		if myInSlice(k, dataSlice) {
			return true
		}
	}
	return false
}

func (self *MyChat) calcChatTablenameWithLock() (tablenames []string, err error) {
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()

	tablenames = make([]string, 0)
	var tName string

	for udKey, ud := range self.allUser {
		if udKey != ud.Id {
			err = errors.New(fmt.Sprintf("数据异常,udKey=%v,udId=%v", udKey, ud.Id))
			tablenames = nil
			return
		}

		if ud.FriendPos == nil {
			continue
		}

		for friendId, _ := range ud.FriendPos {
			if tName, err = calcChatTableOfUser(ud.Id, friendId); err != nil {
				tablenames = nil
				return
			}
			tablenames = append(tablenames, tName)
		}
	}

	for gdKey, gd := range self.allGroup {
		if gdKey != gd.Id {
			err = errors.New(fmt.Sprintf("数据异常,gdKey=%v,gdId=%v", gdKey, gd.Id))
			tablenames = nil
			return
		}
		if tName, err = calcChatTableOfGroup(gd.Id); err != nil {
			return
		}
		tablenames = append(tablenames, tName)
	}

	return
}

func (self *MyChat) createTablesAndSync2() error {
	var err error = nil

	for _ = range "1" {
		var tablenameSlice []string = nil
		if tablenameSlice, err = self.calcChatTablenameWithLock(); err != nil {
			break
		}

		beans := make([]interface{}, 0)
		beans = append(beans, new(KeyValue))
		beans = append(beans, new(UserData))
		beans = append(beans, new(GroupData))
		beans = append(beans, new(PushMessageRaw))
		beans = append(beans, new(PushMessage))
		beans = append(beans, new(ChatMessageRaw))
		for _, tablename := range tablenameSlice {
			cm := new(ChatMessage)
			cm.MyTn = tablename
			beans = append(beans, cm)
		}

		if err = self.engine.CreateTables(beans...); err != nil { //应该是:只要存在这个tablename,就跳过它.
			break
		}

		if err = self.engine.Sync2(beans...); err != nil { //同步数据库结构
			break
		}
	}

	return err
}

func (self *MyChat) loadDataFromDbWithLock() error {
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	return self.loadDataFromDbWithoutLock()
}

func (self *MyChat) loadDataFromDbWithoutLock() error {
	var err error

	for _ = range "1" {
		keyValueSlice := make([]KeyValue, 0)
		if err = self.engine.Find(&keyValueSlice); err != nil {
			break
		}
		for k := range self.allKV { //clear
			delete(self.allKV, k)
		}
		for _, kv := range keyValueSlice {
			self.allKV[kv.Key] = kv
		}

		userDataSlice := make([]UserData, 0)
		if err = self.engine.Find(&userDataSlice); err != nil {
			break
		}
		for k := range self.allUser {
			delete(self.allUser, k)
		}
		for _, ud := range userDataSlice {
			self.allUser[ud.Id] = ud
		}

		groupDataSlice := make([]GroupData, 0)
		if err = self.engine.Find(&groupDataSlice); err != nil {
			break
		}
		for k := range self.allGroup {
			delete(self.allGroup, k)
		}
		for _, gd := range groupDataSlice {
			self.allGroup[gd.Id] = gd
		}
	}

	return err
}

func (self *MyChat) saveDataToDbWithLock() error {
	self.rwMutex.Lock()
	defer self.rwMutex.Unlock()
	return self.saveDataToDbWithoutLock()
}

func (self *MyChat) saveDataToDbWithoutLock() error {
	var err error = nil
	//TODO:如果数据库里面没有这一条记录,我执行update的话,会出现什么行为,这个尚未测试.

	for _ = range "1" {
		var errWhenUpdate bool = false
		for _, kv := range self.allKV {
			if err = self.UpdateOnce(&kv); err != nil {
				errWhenUpdate = true
				break
			}
		}
		if errWhenUpdate {
			break
		}

		for _, ud := range self.allUser {
			if err = self.UpdateOnce(&ud); err != nil {
				errWhenUpdate = true
				break
			}
		}
		if errWhenUpdate {
			break
		}

		for _, gd := range self.allGroup {
			if err = self.UpdateOnce(&gd); err != nil {
				errWhenUpdate = true
				break
			}
		}
		if errWhenUpdate {
			break
		}
	}

	return err
}

func (self *MyChat) findUserWithoutLock(uId *int64, uAlias *string) (ud UserData, err error) {
	if (uId == nil && uAlias == nil) || (uId != nil && uAlias != nil) {
		err = errors.New("uId和uAlias需要:有且仅有一个有效数据!")
		return
	}

	if uId != nil {
		var isOk bool = false
		if ud, isOk = self.allUser[(*uId)]; !isOk {
			err = errors.New(fmt.Sprintf("找不到uId=%v的用户", *uId))
			return
		} else {
			return
		}
	} else if uAlias != nil {
		for _, _ud := range self.allUser {
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

	panic("程序进入了无法到达的逻辑!")
	return
}

func (self *MyChat) findGroupWithoutLock(gId *int64, gAlias *string) (gd GroupData, err error) {
	if (gId == nil && gAlias == nil) || (gId != nil && gAlias != nil) {
		err = errors.New("gId和gAlias需要:有且仅有一个有效数据!")
		return
	}

	if gId != nil {
		var isOk bool = false
		if gd, isOk = self.allGroup[(*gId)]; !isOk {
			err = errors.New(fmt.Sprintf("找不到gId=%v的用户", *gId))
			return
		} else {
			return
		}
	} else if gAlias != nil {
		for _, _gd := range self.allGroup {
			if _gd.Alias == *gAlias {
				gd = _gd
				return
			}
		}
		err = errors.New(fmt.Sprintf("找不到gAlias=%v的用户", *gAlias))
		return
	} else {
		panic("程序进入了无法到达的逻辑!")
	}

	panic("程序进入了无法到达的逻辑!")
	return
}

func (self *MyChat) findAndMergeUserWithoutLock(uId []int64, uAlias []string) (nuI []int64, nuA []string, err error) {
	err = nil

	calcUser := map[int64]UserData{}
	var ud UserData
	if uId != nil {
		for _, _id := range uId {
			if ud, err = self.findUserWithoutLock(&_id, nil); err != nil {
				break
			} else {
				calcUser[ud.Id] = ud
			}
		}
	}
	if uAlias != nil {
		for _, _alias := range uAlias {
			if ud, err = self.findUserWithoutLock(nil, &_alias); err != nil {
				break
			} else {
				calcUser[ud.Id] = ud
			}
		}
	}

	if err == nil {
		nuI = make([]int64, 0)  //newUserId
		nuA = make([]string, 0) //newUserAlias

		for _, ud := range calcUser {
			nuI = append(nuI, ud.Id)
			nuA = append(nuA, ud.Alias)
		}
	}

	return
}

func (self *MyChat) findAndMergeGroupWithoutLock(gId []int64, gAlias []string) (ngI []int64, ngA []string, err error) {
	calcGroup := map[int64]GroupData{}
	var gd GroupData
	if gId != nil {
		for _, _id := range gId {
			if gd, err = self.findGroupWithoutLock(&_id, nil); err != nil {
				break
			} else {
				calcGroup[gd.Id] = gd
			}
		}
	}
	if gAlias != nil {
		for _, _alias := range gAlias {
			if gd, err = self.findGroupWithoutLock(nil, &_alias); err != nil {
				break
			} else {
				calcGroup[gd.Id] = gd
			}
		}
	}

	if err == nil {
		ngI = make([]int64, 0)  //newGroupId
		ngA = make([]string, 0) //newGroupAlias

		for _, gd = range calcGroup {
			ngI = append(ngI, gd.Id)
			ngA = append(ngA, gd.Alias)
		}
	}

	return
}

func (self *MyChat) queryLastPushMessageRaw() (data PushMessageRaw, err error) {
	var isOk bool = false
	// SELECT * FROM tablename ORDER BY id DESC LIMIT 1
	isOk, err = self.engine.Desc("id").Get(&data)
	if (isOk && err != nil) || (!isOk && err == nil) {
		panic(fmt.Sprintf("xorm的逻辑异常,isOk=%v,err=%v", isOk, err))
	}
	return
}

func (self *MyChat) queryLastPushMessage() (data PushMessage, err error) {
	var isOk bool = false
	// SELECT * FROM tablename ORDER BY id DESC LIMIT 1
	isOk, err = self.engine.Desc("id").Get(&data)
	if (isOk && err != nil) || (!isOk && err == nil) {
		panic(fmt.Sprintf("xorm的逻辑异常,isOk=%v,err=%v", isOk, err))
	}
	return
}

func (self *MyChat) queryLastChatMessageRaw() (data ChatMessageRaw, err error) {
	var isOk bool = false
	// SELECT * FROM tablename ORDER BY id DESC LIMIT 1
	isOk, err = self.engine.Desc("id").Get(&data)
	if (isOk && err != nil) || (!isOk && err == nil) {
		panic(fmt.Sprintf("xorm的逻辑异常,isOk=%v,err=%v", isOk, err))
	}
	return
}

func (self *MyChat) queryLastChatMessage(tablename string) (data ChatMessage, err error) {
	var isOk bool = false
	// SELECT * FROM tablename ORDER BY id DESC LIMIT 1
	isOk, err = self.engine.Table(tablename).Desc("id").Get(&data)
	if (isOk && err != nil) || (!isOk && err == nil) {
		panic(fmt.Sprintf("xorm的逻辑异常,tablename=%v,isOk=%v,err=%v", tablename, isOk, err))
	}
	return
}

func (self *MyChat) InsertOne(bean interface{}) error {
	affected, err := self.engine.InsertOne(bean)
	if (affected <= 0 && err == nil) || (affected > 0 && err != nil) {
		panic(fmt.Sprintf("xorm的逻辑异常,InsertOne,affected=%v,err=%v", affected, err))
	}
	return err
}

func (self *MyChat) Insert(beans ...interface{}) error {
	affected, err := self.engine.Insert(beans...)
	if (affected <= 0 && err == nil) || (affected > 0 && err != nil) {
		panic(fmt.Sprintf("xorm的逻辑异常,InsertOne,affected=%v,err=%v", affected, err))
	}
	return err
}

func (self *MyChat) UpdateOnce(bean interface{}, condiBeans ...interface{}) error {
	affected, err := self.engine.Update(bean, condiBeans...)
	if (affected <= 0 && err == nil) || (affected > 0 && err != nil) {
		panic(fmt.Sprintf("xorm的逻辑异常,Update,affected=%v,err=%v", affected, err))
	}
	return err
}
