package main

import (
	"fmt"
)

func main() {
	var err error
	cData := New_CacheData()
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
