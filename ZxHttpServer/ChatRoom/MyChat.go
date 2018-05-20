package ChatRoom

import (
	"errors"
	"reflect"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zx9229/zxgo/zxxorm"
	"github.com/zx9229/zxgo_temp/ZxHttpServer/ChatRoom/CacheData"
	"github.com/zx9229/zxgo_temp/ZxHttpServer/ChatRoom/MyStruct"
)

type ChatRoom struct {
	engine     *xorm.Engine
	mutex      *sync.Mutex
	cache      *CacheData.CacheData
	allKV      map[string]MyStruct.KeyValue
	chanPmr    chan *MyStruct.PushMessageRaw //存数据库专用
	chanCmr    chan *MyStruct.ChatMessageRaw //存数据库专用
	chanCm     chan *MyStruct.ChatMessage    //存数据库专用
	mapSession map[*websocket.Conn]string
}

func New_ChatRoom() *ChatRoom {
	newData := new(ChatRoom)

	newData.engine = nil
	newData.mutex = new(sync.Mutex)
	newData.cache = nil
	newData.allKV = make(map[string]MyStruct.KeyValue)
	newData.chanPmr = make(chan *MyStruct.PushMessageRaw, 1024)
	newData.chanCmr = make(chan *MyStruct.ChatMessageRaw, 1024)
	newData.chanCm = make(chan *MyStruct.ChatMessage, 1024)
	newData.mapSession = make(map[*websocket.Conn]string)

	return newData
}

func (self *ChatRoom) Engine() *xorm.Engine {
	return self.engine
}

func (self *ChatRoom) Init(driverName string, dataSourceName string, locationName string) error {
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

		if false {
			self.engine.ShowSQL(true)
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

func (self *ChatRoom) createTablesAndSync2() error {
	var err error = nil

	for _ = range "1" {
		beans := make([]interface{}, 0)
		beans = append(beans, new(MyStruct.KeyValue))
		beans = append(beans, new(MyStruct.UserData))
		beans = append(beans, new(MyStruct.GroupData))
		beans = append(beans, &MyStruct.PushMessageRaw{MyTn: CacheData.TableName_PushRow()})
		beans = append(beans, &MyStruct.ChatMessageRaw{MyTn: CacheData.TableName_ChatRow()})
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

func (self *ChatRoom) loadDataFromDbWithLock() error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.loadDataFromDbWithoutLock()
}

func (self *ChatRoom) loadDataFromDbWithoutLock() error {
	var err error

	for _ = range "1" {

		keyValueSlice := make([]MyStruct.KeyValue, 0)
		if err = self.engine.Find(&keyValueSlice); err != nil {
			break
		}

		for k := range self.allKV { //clear
			delete(self.allKV, k)
		}

		for _, kv := range keyValueSlice {
			self.allKV[kv.Key] = kv
		}

		var tmpCache *CacheData.CacheData = new(CacheData.CacheData) //此时(self.cache == nil)是true.
		if kvData, ok := self.allKV[reflect.TypeOf(tmpCache).Name()]; ok {
			if err = tmpCache.FromJson(kvData.Value); err != nil {
				break
			} else {
				self.cache = tmpCache
			}
		} else {
			self.cache = CacheData.New_CacheData()
		}
	}

	return err
}

func (self *ChatRoom) saveDataToDbWithLock() error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.saveDataToDbWithoutLock()
}

func (self *ChatRoom) saveDataToDbWithoutLock() error {
	var err error = nil
	//TODO:如果数据库里面没有这一条记录,我执行update的话,会出现什么行为,这个尚未测试.

	for _ = range "1" {

		var jsonStr string
		if jsonStr, err = self.cache.ToJson(); err != nil {
			break
		}

		kvData := MyStruct.KeyValue{}
		kvData.Key = reflect.TypeOf(*(self.cache)).Name()
		kvData.Value = jsonStr
		self.allKV[kvData.Key] = kvData

		var errWhenUpdate bool = false
		for _, kv := range self.allKV {
			if err = zxxorm.Upsert(self.engine, &kv); err != nil {
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

func (self *ChatRoom) AddUser(alias string, password string) error {
	var err error
	for _ = range "1" {
		if err = self.cache.AddUser(alias, password); err != nil {
			break
		}
		if err = self.saveDataToDbWithLock(); err != nil {
			break
		}
	}
	return err
}

func (self *ChatRoom) AddFriends(fId1 int64, fId2 int64) error {
	var err error
	for _ = range "1" {
		if err = self.cache.AddFriends(fId1, fId2); err != nil {
			break
		}
		cm := new(MyStruct.ChatMessage)
		cm.MyTn = CacheData.TableName_Friend(fId1, fId2)
		if err = self.engine.CreateTables(cm); err != nil { //应该是:只要存在这个tablename,就跳过它.
			break
		}
		if err = self.saveDataToDbWithLock(); err != nil {
			break
		}
	}
	return err
}
