package main

import (
	"fmt"

	"github.com/zx9229/zxgo_temp/ZxHttpServer2/CacheData"
)

func main() {
	var err error
	cData := CacheData.New_CacheData()
	if err = cData.Check(); true {
		fmt.Println("1:", err)
	}
	if id, err := cData.AddUser("a", "p"); true {
		fmt.Println("2:", id, err)
	}
	if err := cData.AddFriend(1, 2); true {
		fmt.Println("3:", err)
	}
	fmt.Println("DONE.")
}
