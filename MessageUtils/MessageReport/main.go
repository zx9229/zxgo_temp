package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zx9229/zxgo/zxxorm"
	"github.com/zx9229/zxgo_temp/MessageUtils/TxStruct"
)

func main() {
	tp1 := time.Now()

	debugPtr := flag.Bool("debug", false, "debug mode")
	jsonPtr := flag.String("json", "", "data insert to database")
	stodPtr := flag.Bool("stod", false, "[single quotation marks] to [double quotation marks]")
	helpPtr := flag.Bool("help", false, `show this help. "{ 'Status':0, 'Message':'', 'Group1':'' }"`)
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
			fmt.Println("===+===")
			fmt.Println(*jsonPtr)
			fmt.Println("===-===")
			fmt.Println("=>")
		}
		fmt.Println("===+===")
		fmt.Println(jsonStr)
		fmt.Println("===-===")
	}

	var err error

	data := new(TxStruct.ProxyReqRsp)
	if err = json.Unmarshal([]byte(jsonStr), data); err != nil {
		fmt.Fprintln(os.Stderr, "[Unmarshal]", err)
		os.Exit(1)
	}

	tp2 := time.Now()
	err = InsertToDb(data)
	tp3 := time.Now()

	if *debugPtr {
		fmt.Println(fmt.Sprintf("t2-t1=%v, t3-t2=%v, t3-t1=%v", tp2.Sub(tp1), tp3.Sub(tp2), tp3.Sub(tp1)))
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "[InsertToDb]", err)
		os.Exit(1)
	}
}

func calcDataSourceName() (name string, err error) {
	if name, err = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
		return
	}
	name = filepath.Join(name, TxStruct.DATA_SOURCE_NAME)
	return
}

func InsertToDb(data *TxStruct.ProxyReqRsp) error {
	var err error
	var dataSourceName string
	var engine *xorm.Engine

	if dataSourceName, err = calcDataSourceName(); err != nil {
		return err
	}

	if engine, err = xorm.NewEngine(TxStruct.DRIVER_NAME, dataSourceName); err != nil {
		return err
	}

	if location, err2 := time.LoadLocation("Local"); err2 != nil {
		err = err2
		return err
	} else {
		engine.TZLocation = location
	}

	defer engine.Close()

	if err = zxxorm.EngineInsertOne(engine, data); err != nil {
		return err
	}

	return err
}
