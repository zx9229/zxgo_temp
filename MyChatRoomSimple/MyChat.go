package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

type MyChat struct {
	engine     *xorm.Engine
	rwMutex    *sync.Mutex
	cache      *CacheData
	allKV      map[string]KeyValue
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
	newData.cache = nil
	newData.allKV = make(map[string]KeyValue)
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

	}

	return err
}

func (self *MyChat) createTablesAndSync2() error {
	var err error = nil

	for _ = range "1" {
		beans := make([]interface{}, 0)
		beans = append(beans, new(KeyValue))
		beans = append(beans, new(UserData))
		beans = append(beans, new(GroupData))
		beans = append(beans, new(PushMessageRaw))
		beans = append(beans, new(PushMessage))
		beans = append(beans, new(ChatMessageRaw))
		//for _, tablename := range tablenameSlice {
		//	cm := new(ChatMessage)
		//	cm.MyTn = tablename
		//	beans = append(beans, cm)
		//}

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

		var tmpCache *CacheData = new(CacheData) //此时(self.cache == nil)是true.
		if kvData, ok := self.allKV[reflect.TypeOf(tmpCache).Name()]; ok {
			if err = json.Unmarshal([]byte(kvData.Value), tmpCache); err != nil {
				break
			} else {
				self.cache = tmpCache
			}
		} else {
			self.cache = NewCacheData(self.engine)
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

		var jsonByte []byte
		if jsonByte, err = json.Marshal(self.cache); err != nil {
			break
		}

		kvData := KeyValue{}
		kvData.Key = reflect.TypeOf(*(self.cache)).Name()
		kvData.Value = string(jsonByte)
		self.allKV[kvData.Key] = kvData

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

func (self *MyChat) addUser(alias string, password string) error {
	var err error
	if err = self.cache.addUser(alias, password); err != nil {
		return err
	}
	if err = self.saveDataToDbWithLock(); err != nil {
		return err
	}
	return err
}

func (self *MyChat) AddFriends(fId1 int64, fId2 int64) error {
	var err error
	for _ = range "1" {
		if err = self.cache.AddFriends(fId1, fId2); err != nil {
			break
		}
		cm := new(ChatMessage)
		cm.MyTn = calcTableNameFriend(fId1, fId2)
		if err = self.engine.CreateTables(cm); err != nil { //应该是:只要存在这个tablename,就跳过它.
			break
		}

		if err = self.saveDataToDbWithLock(); err != nil {
			break
		}
	}

	return err
}
