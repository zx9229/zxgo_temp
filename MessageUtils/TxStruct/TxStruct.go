package TxStruct

import (
	"time"
)

type ReportReq struct {
	UserId  int64
	RefId   int64     //rowId
	RefTime time.Time //rowUpdateTime
	Status  int       // (三态) 0=>正常;1=>警告;其他值=>错误
	Message string
	Group1  string
	Group2  string
	Group3  string
	Group4  string
}

type ReportRsp struct {
	UserId  int64
	RefId   int64
	Id      int64 // 0=>没有入库;正数=>写入数据库
	Code    int   // 0=>处理成功;其他值=>处理失败
	Message string
}

type ReportData struct {
	Id   int64     `xorm:"notnull pk autoincr"` //数据库的递增序号.
	Time time.Time `xorm:"created"`             //插入数据库的时刻.
	//
	UserId  int64
	RefId   int64     //rowId
	RefTime time.Time //rowUpdateTime
	Status  int       // (三态) 0=>正常;1=>警告;其他值=>错误
	Message string
	Group1  string
	Group2  string
	Group3  string
	Group4  string
}

type TxData struct {
	Type string
	Data interface{}
}
