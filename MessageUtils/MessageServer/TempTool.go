package main

import (
	"fmt"

	"github.com/go-xorm/xorm"
	"github.com/zx9229/zxgo/zxxorm"
)

func InsertOne(engine *xorm.Engine, data *ReportData) (id int64, err error) {
	session := engine.NewSession()
	defer session.Close()
	var needRollback bool = true
	defer func() {
		if needRollback {
			session.Rollback()
		}
	}()
	for _ = range "1" {
		if err = session.Begin(); err != nil {
			break
		}
		if err = zxxorm.SessionInsertOne(session, data); err != nil {
			break
		}
		//成功执行insert操作后,xorm内部自动对data的Id填值了,因为传进去的data是一个指针,所以把Id的值带出来了.
		if data.Id <= 0 {
			err = fmt.Errorf("should be a positive number, actually %v", data.Id)
			break
		}
		if err = session.Commit(); err != nil {
			break
		}
		needRollback = false
		id = data.Id
	}
	return
}
