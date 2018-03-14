package main

import (
	"errors"
	"sync"

	"github.com/go-xorm/xorm"
)

var (
	errParamsValue = errors.New("Params value error") //参数的值有问题.
	errKeyNotFound = errors.New("Not found key")      //在map里找不到对应的key.
)

type UserGroup struct {
	mtxUser   *sync.Mutex              //互斥锁
	mapUserI  map[int64]UserData       //以Id为key
	mapUserA  map[string]UserData      //以alias为key
	mtxGroup  *sync.Mutex              //互斥锁
	mapGroupI map[int64]GroupData      //以Id为key
	mapGroupA map[string]GroupData     //以alias为key
	mtxGU     *sync.Mutex              //互斥锁
	mapGU     map[int64]map[int64]bool //衍生的user和group的数据
}

func NewUserGroup() *UserGroup {
	newData := new(UserGroup)

	newData.mtxUser = new(sync.Mutex)
	newData.mapUserI = make(map[int64]UserData)
	newData.mapUserA = make(map[string]UserData)
	newData.mapGroupI = make(map[int64]GroupData)
	newData.mapGroupA = make(map[string]GroupData)
	newData.mtxGU = new(sync.Mutex)
	newData.mapGU = make(map[int64]map[int64]bool)

	return newData
}

func (self *UserGroup) FindUserWithLock(uId *int64, uAlias *string) (data UserData, err error) {
	//[线程安全]
	self.mtxUser.Lock()
	defer self.mtxUser.Unlock()
	var ok bool
	if uId != nil && uAlias == nil {
		if data, ok = self.mapUserI[*uId]; !ok {
			err = errKeyNotFound
		}
	} else if uId == nil && uAlias != nil {
		if data, ok = self.mapUserA[*uAlias]; !ok {
			err = errKeyNotFound
		}
	} else {
		err = errParamsValue
	}
	return
}

// 如果找不到,就立即退出.
func (self *UserGroup) findAndMergeUser(suI []int64, suA []string) (nuI []int64, nuA []string, err error) {
	//[线程安全]
	tmpData := make(map[int64]UserData)
	var ud UserData
	if suI != nil {
		for _, uI := range suI {
			if ud, err = self.FindUserWithLock(&uI, nil); err != nil {
				return
			} else {
				tmpData[ud.Id] = ud
			}
		}
	}
	if suA != nil {
		for _, uA := range suA {
			if ud, err = self.FindUserWithLock(nil, &uA); err != nil {
				return
			} else {
				tmpData[ud.Id] = ud
			}
		}
	}
	if len(tmpData) > 0 {
		nuI = make([]int64, 0)
		nuA = make([]string, 0)
		for _, ud = range tmpData {
			nuI = append(nuI, ud.Id)
			nuA = append(nuA, ud.Alias)
		}
	}
	return
}

func (self *UserGroup) FindGroupWithLock(gId *int64, gAlias *string) (data GroupData, err error) {
	self.mtxGroup.Lock()
	defer self.mtxGroup.Unlock()
	var ok bool
	if gId != nil && gAlias == nil {
		if data, ok = self.mapGroupI[*gId]; !ok {
			err = errKeyNotFound
		}
	} else if gId == nil && gAlias != nil {
		if data, ok = self.mapGroupA[*gAlias]; !ok {
			err = errKeyNotFound
		}
	} else {
		err = errParamsValue
	}
	return
}

func (self *UserGroup) FindAndMergeGroup(sgI []int64, sgA []string) (ngI []int64, ngA []string, err error) {
	tmpData := make(map[int64]GroupData)
	var gd GroupData
	if sgI != nil {
		for _, gI := range sgI {
			if gd, err = self.FindGroupWithLock(&gI, nil); err != nil {
				return
			} else {
				tmpData[gd.Id] = gd
			}
		}
	}
	if sgA != nil {
		for _, gA := range sgA {
			if gd, err = self.FindGroupWithLock(nil, &gA); err != nil {
				return
			} else {
				tmpData[gd.Id] = gd
			}
		}
	}
	if len(tmpData) > 0 {
		ngI = make([]int64, 0)
		ngA = make([]string, 0)
		for _, gd = range tmpData {
			ngI = append(ngI, gd.Id)
			ngA = append(ngA, gd.Alias)
		}
	}
	return
}

func (self *UserGroup) saveToDbAndReload(engine *xorm.Engine, mapUserI map[int64]UserData, mapGroupI map[int64]GroupData) error {
	self.mtxUser.Lock()
	defer self.mtxUser.Unlock()
	self.mtxGroup.Lock()
	defer self.mtxGroup.Unlock()
	self.mtxGU.Lock()
	defer self.mtxGU.Unlock()

	var err error
	session := engine.NewSession()
	defer session.Close()

	if err = session.Begin(); err != nil {
		return err
	}
	oldSliceUser := m2sUserDataI(mapUserI)
	if _, err = session.Insert(oldSliceUser); err != nil {
		if err2 := session.Rollback(); err2 != nil {
			panic(err2)
		}
		return err
	}
	oldSliceGroup := m2sGroupDataI(mapGroupI)
	if _, err = session.Insert(oldSliceGroup); err != nil {
		if err2 := session.Rollback(); err2 != nil {
			panic(err2)
		}
		return err
	}
	var newSliceUser []UserData
	var newSliceGroup []GroupData
	if err = session.Find(&newSliceUser); err != nil {
		if err2 := session.Rollback(); err2 != nil {
			panic(err2)
		}
		return err
	}
	if err = session.Find(&newSliceGroup); err != nil {
		if err2 := session.Rollback(); err2 != nil {
			panic(err2)
		}
		return err
	}
	mUDI := s2mUserDataI(newSliceUser)
	mGDI := s2mGroupDataI(newSliceGroup)
	if checkDataAndOk(mUDI, mGDI) == false {
		if err2 := session.Rollback(); err2 != nil {
			panic(err2)
		}
		err = errors.New("数据自检失败!")
		return err
	}
	//TODO:新老数据对比校验.
	if err = session.Commit(); err != nil {
		if err2 := session.Rollback(); err2 != nil {
			panic(err2)
		}
		return err
	}
	return err
}

func (self *UserGroup) LoadFromDb(engine *xorm.Engine) error {
	self.mtxUser.Lock()
	defer self.mtxUser.Unlock()
	self.mtxGroup.Lock()
	defer self.mtxGroup.Unlock()
	self.mtxGU.Lock()
	defer self.mtxGU.Unlock()

	var err error
	var sliceUser []UserData
	if err = engine.Find(&sliceUser); err != nil {
		return err
	}
	var sliceGroup []GroupData
	if err = engine.Find(&sliceGroup); err != nil {
		return err
	}

	mUDI := s2mUserDataI(sliceUser)
	mGDI := s2mGroupDataI(sliceGroup)
	if checkDataAndOk(mUDI, mGDI) == false {
		err = errors.New("数据自检失败!")
		return err
	}

	self.mapUserI = mUDI
	self.mapUserA = s2mUserDataA(sliceUser)
	self.mapGroupI = mGDI
	self.mapGroupA = s2mGroupDataA(sliceGroup)
	self.mapGU = toGU(self.mapUserI, self.mapGroupI)

	return err
}

func checkDataAndOk(mapUser map[int64]UserData, mapGroup map[int64]GroupData) bool {
	var ok bool
	for _, gd := range mapGroup {
		if superUd, ok2 := mapUser[gd.SuperId]; !ok2 {
			ok = ok2
			return ok
		} else {
			if _, ok = superUd.GroupPos[gd.Id]; !ok {
				return ok
			}
		}
		for _, aId := range gd.AdminId {
			if adminUd, ok2 := mapUser[aId]; !ok2 {
				ok = ok2
				return ok
			} else {
				if _, ok = adminUd.GroupPos[gd.Id]; !ok {
					return ok
				}
			}
		}
	}
	for _, ud := range mapUser {
		for fId, _ := range ud.FriendPos {
			if _, ok = mapUser[fId]; !ok {
				return ok
			}
		}
		for gId, _ := range ud.GroupPos {
			if _, ok = mapGroup[gId]; !ok {
				return ok
			}
		}
	}
	return ok
}

func checkDataAndOk2(mapUser map[int64]UserData, mapGroup map[int64]GroupData, mapGU map[int64]map[int64]bool) bool {
	var ok bool
	if ok = checkDataAndOk(mapUser, mapGroup); !ok {
		return ok
	}
	tmpGU := toGU(mapUser, mapGroup)
	if ok = isEqualIM(tmpGU, mapGU); !ok {
		return ok
	}
	return ok
}

func isEqualIM(data1 map[int64]map[int64]bool, data2 map[int64]map[int64]bool) bool {
	if data1 == nil || data2 == nil {
		panic("参数有问题")
	}
	if len(data1) != len(data2) {
		return false
	}
	for i1, m1 := range data1 {
		if m2, ok := data2[i1]; !ok {
			return false
		} else {
			if !isEqualIB(m1, m2) {
				return false
			}
		}
	}
	return true
}

func isEqualIB(data1 map[int64]bool, data2 map[int64]bool) bool {
	if data1 == nil || data2 == nil {
		panic("参数有问题")
	}
	if len(data1) != len(data2) {
		return false
	}
	for i1, b1 := range data1 {
		if b2, ok := data2[i1]; !ok {
			return false
		} else {
			if b1 != b2 {
				return false
			}
		}
	}
	return true
}

func toGU(mapUser map[int64]UserData, mapGroup map[int64]GroupData) map[int64]map[int64]bool {
	mapGU := make(map[int64]map[int64]bool)
	for _, gd := range mapGroup {
		mapGU[gd.Id][gd.SuperId] = true
		for _, uId := range gd.AdminId {
			mapGU[gd.Id][uId] = true
		}
	}
	for _, ud := range mapUser {
		for gId, _ := range ud.GroupPos {
			mapGU[gId][ud.Id] = true
		}
	}
	return mapGU
}

func s2mUserDataI(sliceData []UserData) map[int64]UserData {
	if sliceData == nil {
		return nil
	}
	mapData := make(map[int64]UserData)
	for _, ud := range sliceData {
		if _, ok := mapData[ud.Id]; ok {
			panic("逻辑错误")
		}
		mapData[ud.Id] = ud
	}
	return mapData
}

func s2mUserDataA(sliceData []UserData) map[string]UserData {
	if sliceData == nil {
		return nil
	}
	mapData := make(map[string]UserData)
	for _, ud := range sliceData {
		if _, ok := mapData[ud.Alias]; ok {
			panic("逻辑错误")
		}
		mapData[ud.Alias] = ud
	}
	return mapData
}

func s2mGroupDataI(sliceData []GroupData) map[int64]GroupData {
	if sliceData == nil {
		return nil
	}
	mapData := make(map[int64]GroupData)
	for _, gd := range sliceData {
		if _, ok := mapData[gd.Id]; ok {
			panic("逻辑错误")
		}
		mapData[gd.Id] = gd
	}
	return mapData
}

func s2mGroupDataA(sliceData []GroupData) map[string]GroupData {
	if sliceData == nil {
		return nil
	}
	mapData := make(map[string]GroupData)
	for _, gd := range sliceData {
		if _, ok := mapData[gd.Alias]; ok {
			panic("逻辑错误")
		}
		mapData[gd.Alias] = gd
	}
	return mapData
}

func m2sUserDataI(mapData map[int64]UserData) []UserData {
	if mapData == nil {
		return nil
	}
	sliceData := make([]UserData, 0)
	for _, ud := range mapData {
		sliceData = append(sliceData, ud)
	}
	//TODO:对sliceData进行排序.
	return sliceData
}

func m2sUserDataA(mapData map[string]UserData) []UserData {
	if mapData == nil {
		return nil
	}
	sliceData := make([]UserData, 0)
	for _, ud := range mapData {
		sliceData = append(sliceData, ud)
	}
	//TODO:对sliceData进行排序.
	return sliceData
}

func m2sGroupDataI(mapData map[int64]GroupData) []GroupData {
	if mapData == nil {
		return nil
	}
	sliceData := make([]GroupData, 0)
	for _, gd := range mapData {
		sliceData = append(sliceData, gd)
	}
	//TODO:对sliceData进行排序.
	return sliceData
}

func m2sGroupDataA(mapData map[string]GroupData) []GroupData {
	if mapData == nil {
		return nil
	}
	sliceData := make([]GroupData, 0)
	for _, gd := range mapData {
		sliceData = append(sliceData, gd)
	}
	//TODO:对sliceData进行排序.
	return sliceData
}
