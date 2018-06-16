package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zx9229/zxgo/zxxorm"
	"github.com/zx9229/zxgo_temp/MessageUtils/TxStruct"
)

const (
	DRIVER_NAME      = "sqlite3"
	DATA_SOURCE_NAME = "test_proxy.db"
)

func main() {
	debugPtr := flag.Bool("debug", false, "debug mode")
	jsonPtr := flag.String("json", "", "show file size")
	stodPtr := flag.Bool("stod", false, "[single quotation marks] to [double quotation marks]")
	helpPtr := flag.Bool("help", false, "show this help")
	//所有标志都声明完成以后，调用 flag.Parse() 来执行命令行解析。
	flag.Parse()

	if *helpPtr {
		flag.Usage()
		return
	}

	var jsonStr string
	if *stodPtr {
		jsonStr = strings.Replace(*jsonPtr, `'`, `"`, -1)
	} else {
		jsonStr = *jsonPtr
	}

	if *debugPtr {
		if *stodPtr {
			fmt.Println("======")
			fmt.Println(*jsonPtr)
			fmt.Println("======")
			fmt.Println("=>")
		}
		fmt.Println("======")
		fmt.Println(jsonStr)
		fmt.Println("======")
	}

	var err error

	data := new(TxStruct.ProxyReqRsp)
	if err = json.Unmarshal([]byte(jsonStr), data); err != nil {
		fmt.Fprintln(os.Stderr, "[Unmarshal]", err)
		os.Exit(1)
	}

	if err = XXX(data); err != nil {
		fmt.Fprintln(os.Stderr, "[zxcvb]", err)
		os.Exit(1)
	}

	fmt.Println("DONE.")
	os.Exit(0)
}

func calcDataSourceName() (name string, err error) {
	if name, err = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
		return
	}
	name = filepath.Join(name, DATA_SOURCE_NAME)
	return
}

func XXX(data *TxStruct.ProxyReqRsp) error {
	var err error
	var dataSourceName string
	var engine *xorm.Engine

	if dataSourceName, err = calcDataSourceName(); err != nil {
		return err
	}

	if engine, err = xorm.NewEngine(DRIVER_NAME, dataSourceName); err != nil {
		return err
	}

	defer engine.Close()

	if err = zxxorm.EngineInsertOne(engine, data); err != nil {
		return err
	}

	return err
}
