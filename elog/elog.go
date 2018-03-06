package elog //easylog

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type fileCfgData struct {
	FileName string      //外部配置
	FileFlag int         //外部配置
	FilePerm os.FileMode //外部配置
	file     *os.File    //内部
}

type logCfgData struct {
	FileTag string //外部配置(通过FileTag找到对应的fileCfgData)
	Prefix  string //外部配置
	Flag    int    //外部配置
}

type configData struct {
	AllFileCfg map[string]*fileCfgData
	AllLogCfg  map[string]*logCfgData
}

var Elog *EasyLog = nil
var DLN string = "" //DefaultLoggerName

type EasyLog struct {
	rwMutex    *sync.RWMutex
	mapFileCfg map[string]*fileCfgData
	mapLogger  map[string]*log.Logger
}

func (this *EasyLog) Get(logName string) (logger *log.Logger, ok bool) {
	if this == nil {
		return nil, false
	}

	this.rwMutex.RLock()
	logger, ok = this.mapLogger[logName]
	this.rwMutex.RUnlock()

	return
}

func (this *EasyLog) Printf(logName string, format string, v ...interface{}) bool {
	logger, ok := this.Get(logName)
	if ok {
		logger.Printf(format, v...)
	}
	return ok
}

func (this *EasyLog) Print(logName string, v ...interface{}) bool {
	logger, ok := this.Get(logName)
	if ok {
		logger.Print(v...)
	}
	return ok
}

func (this *EasyLog) Println(logName string, v ...interface{}) bool {
	logger, ok := this.Get(logName)
	if ok {
		logger.Println(v...)
	}
	return ok
}

func (this *EasyLog) Insert(logName string, logger *log.Logger) bool {
	if this == nil || logger == nil {
		return false
	}

	var retVal bool = false
	this.rwMutex.Lock()
	if _, ok := this.mapLogger[logName]; ok {
		retVal = false
	} else {
		this.mapLogger[logName] = logger
		retVal = true
	}
	this.rwMutex.Unlock()
	return retVal
}

func (this *EasyLog) Delete(logName string) bool {
	if this == nil {
		return false
	}

	var retVal bool = false
	this.rwMutex.Lock()
	if _, ok := this.mapLogger[logName]; ok {
		delete(this.mapLogger, logName)
		retVal = true
	}
	this.rwMutex.Unlock()
	return retVal
}

func (this *EasyLog) InitFromCfg(configure configData) error {
	if this != nil {
		return errors.New("对象已存在,不允许重复初始化")
	}

	var retVal error = nil
	const DIR_PERM = 0755 //在Windows上创建一个文件夹,然后在shell下查看其权限,可以看到[drwxr-xr-x],为0755

	for i := 0; i < 1; i++ {
		var err error = nil

		if _, ok := configure.AllLogCfg[DLN]; !ok {
			retVal = errors.New(fmt.Sprintf("找不到默认的logger名,DLN=%v", DLN))
			break
		}

		for _, fileCfg := range configure.AllFileCfg {

			dirname := filepath.Dir(fileCfg.FileName)
			if err = os.MkdirAll(dirname, DIR_PERM); err != nil {
				retVal = errors.New(fmt.Sprintf("创建目录失败,dir=%v,err=%v", dirname, err))
				break
			}

			if fileCfg.file, err = os.OpenFile(fileCfg.FileName, fileCfg.FileFlag, fileCfg.FilePerm); err != nil {
				retVal = errors.New(fmt.Sprintf("OpenFile失败,filename=%v,err=%v", fileCfg.FileName, err))
				fileCfg.file = nil
				break
			}
		}

		mapLogger := map[string]*log.Logger{}

		for logName, logCfg := range configure.AllLogCfg {

			if _, ok := mapLogger[logName]; ok {
				retVal = errors.New(fmt.Sprintf("已经存在同名的logger,name=%v", logName))
				break
			}

			if fileCfg, ok := configure.AllFileCfg[logCfg.FileTag]; ok {
				logger := log.New(fileCfg.file, logCfg.Prefix, logCfg.Flag)
				mapLogger[logName] = logger
			} else {
				retVal = errors.New(fmt.Sprintf("找不到对应的fileCfg,FileTag=%v", logCfg.FileTag))
				break
			}
		}

		Elog = &EasyLog{new(sync.RWMutex), configure.AllFileCfg, mapLogger}
	}

	if retVal != nil {
		for _, fileCfg := range configure.AllFileCfg {
			if fileCfg.file != nil {
				fileCfg.file.Close()
				fileCfg.file = nil
			}
		}
	}

	return retVal
}

func (this *EasyLog) InitFromFile(filename string) error {
	if this != nil {
		return errors.New("对象已存在,不允许重复初始化")
	}

	var err error
	cfgData := configData{}

	for i := 0; i < 1; i++ {
		var data []byte

		if data, err = ioutil.ReadFile(filename); err != nil {
			break
		}

		if err = json.Unmarshal(data, &cfgData); err != nil {
			break
		}

		err = this.InitFromCfg(cfgData)

		break
	}

	return err
}

func (this *EasyLog) Terminate() {
	if this != nil {
		for _, fileCfg := range this.mapFileCfg {
			if fileCfg.file != nil {
				fileCfg.file.Close()
			}
		}
	}
}
