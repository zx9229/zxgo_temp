package main

import (
	"fmt"
)

func main() {
	innerCacheData := new_InnerCacheData()
	err := innerCacheData.check()
	fmt.Println(err)
}
