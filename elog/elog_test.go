package elog

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestOps(t *testing.T) {
	tag := "sometag"
	filename := "elog_test.log"
	var fileFlag int = os.O_CREATE | os.O_WRONLY | os.O_TRUNC //参见os.Create()
	fmt.Printf("fileFlag=%d\n", fileFlag)                     //
	var filePerm os.FileMode = 0666                           //参见os.Create()
	fmt.Printf("filePerm=%d\n", filePerm)                     //
	fileCfg := fileCfgData{filename, fileFlag, filePerm, nil}

	logFlag := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile
	fmt.Printf("logFlag=%d\n", logFlag)
	logFlag = log.Ldate | log.Ltime | log.Lmicroseconds
	fmt.Printf("logFlag=%d\n", logFlag)
	logCfg := logCfgData{tag, "[PREFIX]", logFlag}

	configure := configData{map[string]*fileCfgData{tag: &fileCfg}, map[string]*logCfgData{DLN: &logCfg}}

	if err := Elog.InitFromCfg(configure); err != nil {
		t.Error(fmt.Sprintf("InitFromCfg failed,err=%v", err))
		return
	}
	defer Elog.Terminate()

	Elog.Printf(DLN, "fileCfg=%v", fileCfg)
	Elog.Printf(DLN, "logCfg=%v", logCfg)
	Elog.Print(DLN, configure)
	Elog.Println(DLN, "")
}
