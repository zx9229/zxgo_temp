package TxStruct

import (
	"time"
)

// 将所有的消息并入一个线程中,让这一个线程进行处理,这样,ReportId和Id就顺序了.

type UserData struct {
	Id       int64
	Name     string //名字(和Id一样是唯一的)
	Tag      string //标签
	Memo     string //备注
	ReportId int64  //这个用户使用到哪一个ID了.
}

type ReportReq struct {
	UserId     int64
	ReportId   int64
	UpdateTime time.Time
	Status     int // (三态) 0=>正常;1=>警告;其他值=>错误
	Message    string
	Group1     string
	Group2     string
	Group3     string
	Group4     string
}

type ReportRsp struct {
	UserId   int64
	ReportId int64
	Id       int64  //0=>没有入库;正数=>写入数据库
	Code     int    //0=>处理成功;其他值=>处理失败
	Message  string //对code的详细描述
}

type ReportData struct {
	Id         int64
	UpdateTime time.Time
	data       ReportReq //TODO:展开到这个结构体里.
}
