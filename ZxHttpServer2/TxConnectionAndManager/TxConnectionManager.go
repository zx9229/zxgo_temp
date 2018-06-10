package TxConnectionAndManager

import (
	"encoding/json"
	"fmt"

	"github.com/zx9229/zxgo_temp/ZxHttpServer2/CacheData"
	"github.com/zx9229/zxgo_temp/ZxHttpServer2/ChatStruct"
	"github.com/zx9229/zxgo_temp/ZxHttpServer2/TxStruct"
)

type TxConnectionManager struct {
	parser        *TxStruct.TxParser   //管理器存储起来,让连接使用,因为所有的链接都会用它们.
	cacheData     *CacheData.CacheData //管理器存储起来,让连接使用,因为所有的链接都会用它们.
	allConnection map[*TxConnection]bool
	mapUser       map[int64]map[int]*TxConnection           //计算出来的缓存数据(            int64=>用户名;int=>设备类型)
	mapGroup      map[int64]map[int64]map[int]*TxConnection //计算出来的缓存数据(int64=>组ID;int64=>用户名;int=>设备类型)
}

func New_TxConnectionManager(parser *TxStruct.TxParser, cacheData *CacheData.CacheData) *TxConnectionManager {
	curData := new(TxConnectionManager)
	//
	curData.parser = parser
	curData.cacheData = cacheData
	curData.allConnection = make(map[*TxConnection]bool)
	curData.mapUser = make(map[int64]map[int]*TxConnection)
	curData.mapGroup = make(map[int64]map[int64]map[int]*TxConnection)
	//
	return curData
}

func (self *TxConnectionManager) Flush(jsonStr string) error {
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

	assistFun := func(mapCache map[int64]map[int]*TxConnection, conn *TxConnection) {
		var temp map[int]*TxConnection
		var isOk bool
		if temp, isOk = mapCache[conn.UD.Id]; !isOk {
			temp = make(map[int]*TxConnection)
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
			var temp map[int64]map[int]*TxConnection
			var isOk bool
			if temp, isOk = self.mapGroup[gId]; !isOk {
				temp = make(map[int64]map[int]*TxConnection)
				self.mapGroup[gId] = temp
			}
			assistFun(temp, conn)
		}
	}

	return err
}

func (self *TxConnectionManager) Send(msgSlice []*ChatStruct.MessageData) {

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

func (self *TxConnectionManager) HandleLogin(conn *TxConnection) {

	if _, ok := self.allConnection[conn]; !ok {
		panic("逻辑异常")
	}

	if conn.UD == nil || conn.DeviceType <= 0 {
		panic("逻辑异常_2")
	}

	assistFun := func(mapCache map[int64]map[int]*TxConnection, conn *TxConnection) {
		var temp map[int]*TxConnection
		var isOk bool
		if temp, isOk = mapCache[conn.UD.Id]; !isOk {
			temp = make(map[int]*TxConnection)
			mapCache[conn.UD.Id] = temp
		}
		temp[conn.DeviceType] = conn
	}

	assistFun(self.mapUser, conn)

	for gId := range conn.UD.Groups {
		var temp map[int64]map[int]*TxConnection
		var isOk bool
		if temp, isOk = self.mapGroup[gId]; !isOk {
			temp = make(map[int64]map[int]*TxConnection)
			self.mapGroup[gId] = temp
		}
		assistFun(temp, conn)
	}
}

func (self *TxConnectionManager) HandleLogout(conn *TxConnection) {

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

func (self *TxConnectionManager) HandleConnected(conn *TxConnection) {

	if _, ok := self.allConnection[conn]; ok {
		panic("逻辑异常")
	}

	self.allConnection[conn] = true
}

func (self *TxConnectionManager) HandleDisconnected(conn *TxConnection) {

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
