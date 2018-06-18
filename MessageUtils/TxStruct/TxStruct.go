package TxStruct

import (
	"time"
)

const (
	DRIVER_NAME      = "sqlite3"
	DATA_SOURCE_NAME = "test_proxy.db"
)

type TxInterface interface {
	// 获取字段(TN=>TypeName)的值.
	// 函数体实际上是{ return self.TN }.
	GET_TN() string

	// 计算类型的名字(calc type name).
	// if  (modifyTN == true) { TN = TypeName }.
	CALC_TN(modifyTN bool) string

	// 将自身转成json字符串.
	// 转换失败的话,如果(panicWhenError == true),就panic; 否则返回空字符串.
	TO_JSON(panicWhenError bool) string
}

func inner_check_by_compile() {
	//这个函数作用: 在编译时,检查各个结构体是否进行正常书写.
	slice_ := make([]TxInterface, 0)
	slice_ = append(slice_, new(AddAgentReq))
	slice_ = append(slice_, new(AddAgentRsp))
}

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

type ProxyReqRsp struct {
	UserId     int64     `xorm:"notnull"`             //(req)
	RefId      int64     `xorm:"notnull pk autoincr"` //(req)row_id(数据库里的第几行)
	RefTime    time.Time `xorm:"created"`             //(req)这个Field将在Insert时自动赋值为当前时间
	Status     int       `xorm:"notnull"`             //(req)(三态)0=>正常;1=>警告;其他值=>错误
	Message    string    //(req)
	Group1     string    //(req)
	Group2     string    //(req)
	Group3     string    //(req)
	Group4     string    //(req)
	IsHandled  bool      `xorm:"notnull"` //(是否)已经处理过了(true:已经处理过了)
	RspId      int64     //(rsp)
	RspCode    int       //(rsp)
	RspMessage string    //(rsp)
	UpdateTime time.Time `xorm:"updated"` //这个Field将在Insert或Update时自动赋值为当前时间
}

type AddAgentReq struct {
	TN    string
	ReqId int //请求ID.
	Id    int64
	Memo  string
}

type AddAgentRsp struct {
	TN      string
	Code    int
	Message string
	DataReq *AddAgentReq
}

type AgentInfo struct {
	Id         int64     `xorm:"notnull pk autoincr"`
	Memo       string    //备注
	LastRefId  int64     `xorm:"notnull"`
	CreateTime time.Time `xorm:"created"`
	UpdateTime time.Time `xorm:"updated"`
}
