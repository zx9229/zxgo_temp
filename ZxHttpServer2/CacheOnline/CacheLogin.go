package CacheOnline

import (
	"encoding/json"
	"fmt"

	"github.com/zx9229/zxgo_temp/ZxHttpServer2/CacheData"
	"github.com/zx9229/zxgo_temp/ZxHttpServer2/ChatStruct"
	"github.com/zx9229/zxgo_temp/ZxHttpServer2/TxConnection"
)

type CacheOnline struct {
	allConnection map[*TxConnection.TxConnection]bool
	mapUser       map[int64]map[int]*TxConnection.TxConnection           //计算出来的缓存数据(            int64=>用户名;int=>设备类型)
	mapGroup      map[int64]map[int64]map[int]*TxConnection.TxConnection //计算出来的缓存数据(int64=>组ID;int64=>用户名;int=>设备类型)
}

func New_CacheOnline() *CacheOnline {
	curData := new(CacheOnline)
	curData.allConnection = make(map[*TxConnection.TxConnection]bool)
	curData.mapUser = make(map[int64]map[int]*TxConnection.TxConnection)
	curData.mapGroup = make(map[int64]map[int64]map[int]*TxConnection.TxConnection)
	return curData
}

func (self *CacheOnline) Flush(jsonStr string) error {
	var err error

	tmpObj := new(CacheData.InnerCacheData)
	if err = json.Unmarshal([]byte(jsonStr), tmpObj); err != nil {
		return err
	}

	for conn := range self.allConnection {
		if conn.UD == nil {
			continue
		}
		if newUd, ok := tmpObj.AllUser[conn.UD.Id]; ok {
			conn.UD = newUd
		} else {
			conn.UD = nil //当删除了一个/多个用户之后,对应的连接,就变成没有登录的连接了.
		}
	}

	assistFun := func(mapCache map[int64]map[int]*TxConnection.TxConnection, conn *TxConnection.TxConnection) {
		var temp map[int]*TxConnection.TxConnection
		var isOk bool
		if temp, isOk = mapCache[conn.UD.Id]; !isOk {
			temp = make(map[int]*TxConnection.TxConnection)
			mapCache[conn.UD.Id] = temp
		}
		temp[conn.DeviceType] = conn
	}

	//重新刷新user的cache.
	for k := range self.mapUser { //clear
		delete(self.mapUser, k)
	}
	for conn := range self.allConnection {
		if conn.UD == nil {
			continue
		}
		assistFun(self.mapUser, conn)
	}

	//重新刷新group的cache.
	for k := range self.mapGroup { //clear
		delete(self.mapGroup, k)
	}

	for conn := range self.allConnection {
		if conn.UD == nil {
			continue
		}
		for gId := range conn.UD.Groups {
			var temp map[int64]map[int]*TxConnection.TxConnection
			var isOk bool
			if temp, isOk = self.mapGroup[gId]; !isOk {
				temp = make(map[int64]map[int]*TxConnection.TxConnection)
				self.mapGroup[gId] = temp
			}
			assistFun(temp, conn)
		}
	}

	return err
}

func (self *CacheOnline) Send(msgSlice []*ChatStruct.MessageData) {

	if msgSlice == nil {
		return
	}

	for _, msgData := range msgSlice {
		if isUser, ok := CacheData.TagName_ReceiverIsUserOrGroup(msgData.Tag); ok {
			if isUser {
				if temp, ok := self.mapUser[msgData.Receiver]; ok {
					for _, conn := range temp {
						conn.Send_Temp(msgData)
					}
				}
			} else {
				if mapMember, ok := self.mapGroup[msgData.Receiver]; ok {
					for _, temp := range mapMember {
						for _, conn := range temp {
							conn.Send_Temp(msgData)
						}
					}
				}
			}
		} else {
			panic(fmt.Sprintf("未知的Tag=%v", msgData.Tag))
		}
	}
}

func (self *CacheOnline) HandleLogin(conn *TxConnection.TxConnection) {

	if _, ok := self.allConnection[conn]; !ok {
		panic("逻辑异常")
	}

	if conn.UD == nil || conn.DeviceType <= 0 {
		panic("逻辑异常_2")
	}

	assistFun := func(mapCache map[int64]map[int]*TxConnection.TxConnection, conn *TxConnection.TxConnection) {
		var temp map[int]*TxConnection.TxConnection
		var isOk bool
		if temp, isOk = mapCache[conn.UD.Id]; !isOk {
			temp = make(map[int]*TxConnection.TxConnection)
			mapCache[conn.UD.Id] = temp
		}
		temp[conn.DeviceType] = conn
	}

	assistFun(self.mapUser, conn)

	for gId := range conn.UD.Groups {
		var temp map[int64]map[int]*TxConnection.TxConnection
		var isOk bool
		if temp, isOk = self.mapGroup[gId]; !isOk {
			temp = make(map[int64]map[int]*TxConnection.TxConnection)
			self.mapGroup[gId] = temp
		}
		assistFun(temp, conn)
	}
}

func (self *CacheOnline) HandleLogout(conn *TxConnection.TxConnection) {

	if _, ok := self.allConnection[conn]; !ok {
		panic("逻辑异常")
	}

	if conn.UD == nil || conn.DeviceType <= 0 {
		panic("代码没有设计好, 需要TxConnection调用完这个函数之后,再执行清理操作")
	}

	delete(self.mapUser[conn.UD.Id], conn.DeviceType)

	for gId := range conn.UD.Groups {
		delete(self.mapGroup[gId][conn.UD.Id], conn.DeviceType)
	}
}

func (self *CacheOnline) HandleConnected(conn *TxConnection.TxConnection) {

	if _, ok := self.allConnection[conn]; ok {
		panic("逻辑异常")
	}

	self.allConnection[conn] = true
}

func (self *CacheOnline) HandleDisconnected(conn *TxConnection.TxConnection) {

	if _, ok := self.allConnection[conn]; !ok {
		panic("逻辑异常")
	}

	if conn.UD != nil {
		self.HandleLogout(conn)
	}

	conn.UD = nil
	conn.DeviceType = 0

	delete(self.allConnection, conn)
}
