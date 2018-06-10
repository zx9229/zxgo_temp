package main

import (
	"encoding/json"
	"fmt"

	"golang.org/x/net/websocket"
)

type LoginData struct {
	DeviceType int //设备类型(多设备登录时)
	ud         UserData
}

func toLoginData(deviceType int, ud *UserData) *LoginData {
	curData := new(LoginData)
	curData.DeviceType = deviceType
	curData.ud = *ud
	return curData
}

type CacheOnline struct {
	mapSession map[*websocket.Conn]*LoginData              //登录的数据
	mapUser    map[int64]map[int]*websocket.Conn           //计算出来的缓存数据(            int64=>用户名;int=>设备类型)
	mapGroup   map[int64]map[int64]map[int]*websocket.Conn //计算出来的缓存数据(int64=>组ID;int64=>用户名;int=>设备类型)
}

func (self *CacheOnline) Flush(jsonStr string) error {
	var err error

	tmpObj := new(InnerCacheData)
	if err = json.Unmarshal([]byte(jsonStr), tmpObj); err != nil {
		return err
	}

	sliceToDel := make([]*websocket.Conn, 0) //当删除了一个/多个用户之后.
	for ws, oldData := range self.mapSession {
		if newUd, ok := tmpObj.AllUser[oldData.ud.Id]; ok {
			self.mapSession[ws] = toLoginData(oldData.DeviceType, newUd)
		} else {
			sliceToDel = append(sliceToDel, ws)
		}
	}
	for _, conn := range sliceToDel {
		delete(self.mapSession, conn)
	}

	assistFun := func(mapCache map[int64]map[int]*websocket.Conn, data *LoginData, ws *websocket.Conn) {
		var temp map[int]*websocket.Conn
		var isOk bool
		if temp, isOk = mapCache[data.ud.Id]; !isOk {
			temp = make(map[int]*websocket.Conn)
			mapCache[data.ud.Id] = temp
		}
		temp[data.DeviceType] = ws
	}

	//重新刷新user的cache.
	for k := range self.mapUser { //clear
		delete(self.mapUser, k)
	}
	for ws, data := range self.mapSession {
		assistFun(self.mapUser, data, ws)
	}

	//重新刷新group的cache.
	for k := range self.mapGroup { //clear
		delete(self.mapGroup, k)
	}

	for ws, data := range self.mapSession {
		for gId := range data.ud.Groups {
			var temp map[int64]map[int]*websocket.Conn
			var isOk bool
			if temp, isOk = self.mapGroup[gId]; !isOk {
				temp = make(map[int64]map[int]*websocket.Conn)
				self.mapGroup[gId] = temp
			}
			assistFun(temp, data, ws)
		}
	}

	return err
}

func (self *CacheOnline) Send(msgSlice []*MessageData) {

	if msgSlice == nil {
		return
	}

	for _, msgData := range msgSlice {
		if isUser, ok := TagName_ReceiverIsUserOrGroup(msgData.Tag); ok {
			if isUser {
				if temp, ok := self.mapUser[msgData.Receiver]; ok {
					for _, ws := range temp {
						websocket.Message.Send(ws, msgData)
					}
				}
			} else {
				if mapMember, ok := self.mapGroup[msgData.Receiver]; ok {
					for _, temp := range mapMember {
						for _, ws := range temp {
							websocket.Message.Send(ws, msgData)
						}
					}
				}
			}
		} else {
			panic(fmt.Sprintf("未知的Tag=%v", msgData.Tag))
		}
	}
}

func (self *CacheOnline) HandleLogin(ws *websocket.Conn, ud *UserData, deviceType int) error {
	var err error

	if oldData, ok := self.mapSession[ws]; ok {
		err = fmt.Errorf("已经登录userId=%v用户", oldData.ud.Id)
		return err
	}

	if temp, ok := self.mapUser[ud.Id]; ok {
		if _, ok := temp[deviceType]; ok {
			err = fmt.Errorf("已经登录userId=%v,deviceType=%v用户", ud.Id, deviceType)
			return err
		}
	}

	loginData := toLoginData(deviceType, ud)
	self.mapSession[ws] = loginData

	assistFun := func(mapCache map[int64]map[int]*websocket.Conn, data *LoginData, ws *websocket.Conn) {
		var temp map[int]*websocket.Conn
		var isOk bool
		if temp, isOk = mapCache[data.ud.Id]; !isOk {
			temp = make(map[int]*websocket.Conn)
			mapCache[data.ud.Id] = temp
		}
		temp[data.DeviceType] = ws
	}

	assistFun(self.mapUser, loginData, ws)

	for gId := range loginData.ud.Groups {
		var temp map[int64]map[int]*websocket.Conn
		var isOk bool
		if temp, isOk = self.mapGroup[gId]; !isOk {
			temp = make(map[int64]map[int]*websocket.Conn)
			self.mapGroup[gId] = temp
		}
		assistFun(temp, loginData, ws)
	}

	return err
}

func (self *CacheOnline) HandleLogout(ws *websocket.Conn) {
	var data *LoginData
	var isOk bool
	if data, isOk = self.mapSession[ws]; !isOk {
		return
	}

	delete(self.mapSession, ws)

	delete(self.mapUser[data.ud.Id], data.DeviceType)

	for gId := range data.ud.Groups {
		delete(self.mapGroup[gId][data.ud.Id], data.DeviceType)
	}
}
