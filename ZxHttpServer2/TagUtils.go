package main

import (
	"fmt"
	"strconv"
	"strings"
)

func TagName_User_Chat(uId1 int64, uId2 int64) string {
	if uId1 <= 0 || uId2 <= 0 || uId1 == uId2 {
		panic(fmt.Sprintf("uId1=%v,uId2=%v", uId1, uId2))
	}
	if uId1 > uId2 {
		uId1, uId2 = uId2, uId1
	}
	return fmt.Sprintf("u_%v_%v_", uId1, uId2)
}

func TagName_User_Push(uId int64) string {
	if uId <= 0 {
		panic(fmt.Sprintf("uId=%v", uId))
	}
	return fmt.Sprintf("pu_%v_", uId)
}

func TagName_Group_Chat(gId int64) string {
	if gId <= 0 {
		panic(fmt.Sprintf("gId=%v", gId))
	}
	return fmt.Sprintf("g_%v_", gId)
}

func TagName_Group_Push(gId int64) string {
	if gId <= 0 {
		panic(fmt.Sprintf("gId=%v", gId))
	}
	return fmt.Sprintf("pg_%v_", gId)
}

func TagName_ReceiverIsUserOrGroup(tagName string) (isUser bool, isOk bool) {
	//(isOk == true) : 这个tag的接收者, 要么是user, 要么是group.
	//(isOk == false): 函数判定失败.
	if strings.HasPrefix(tagName, "g_") || strings.HasPrefix(tagName, "pg_") {
		isOk = true
		isUser = false
		return
	} else if strings.HasPrefix(tagName, "u_") || strings.HasPrefix(tagName, "pu_") {
		isOk = true
		isUser = true
		return
	} else {
		isOk = false
		return
	}
}

type TagNameInfo struct {
	IsUserChat  bool //
	ChatUserId1 int64
	ChatUserId2 int64
	IsUserPush  bool //
	PushUserId  int64
	IsGroupChat bool //
	ChatGroupId int64
	IsGroupPush bool //
	PushGroupId int64
}

func ParseTagName(tagName string) (data *TagNameInfo, ok bool) {
	data = nil
	ok = false
	var err error

	fields := strings.Split(tagName, "_")

	if fields[0] == "u" { //u_%v_%v_
		if len(fields) != 4 {
			return
		}
		data = new(TagNameInfo)
		data.IsUserChat = true
		if data.ChatUserId1, err = strconv.ParseInt(fields[1], 10, 64); err != nil {
			data = nil
			return
		}
		if data.ChatUserId2, err = strconv.ParseInt(fields[2], 10, 64); err != nil {
			data = nil
			return
		}
		ok = true
		return
	} else if fields[0] == "pu" { //pu_%v_
		if len(fields) != 3 {
			return
		}
		data = new(TagNameInfo)
		data.IsUserPush = true
		if data.PushUserId, err = strconv.ParseInt(fields[1], 10, 64); err != nil {
			data = nil
			return
		}
		ok = true
		return
	} else if fields[0] == "g" { //g_%v_
		if len(fields) != 3 {
			return
		}
		data = new(TagNameInfo)
		data.IsGroupChat = true
		if data.ChatGroupId, err = strconv.ParseInt(fields[1], 10, 64); err != nil {
			data = nil
			return
		}
		ok = true
		return
	} else if fields[0] == "pg" { //pg_%v_
		if len(fields) != 3 {
			return
		}
		data = new(TagNameInfo)
		data.IsGroupPush = true
		if data.PushGroupId, err = strconv.ParseInt(fields[1], 10, 64); err != nil {
			data = nil
			return
		}
		ok = true
		return
	} else {
		return
	}
}
