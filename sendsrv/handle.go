package main

import (
	"fmt"
	"gim/common"
	"github.com/garyburd/redigo/redis"
	"net/rpc"
	"strconv"
)

const (
	USER_MAX_MSGID_PREFIX = "maxmsgid#"
)

type GroupArgs struct {
	GroupId int
}

//获取用户分配的消息id
func getUserMsgId(userId int) int {
	maxId, _ := redis.Int(redClient.Do("INCR", USER_MAX_MSGID_PREFIX+strconv.Itoa(userId)))
	return maxId
}

//发过来消息字符串
func HandleServerMsg(msg *common.Message, msClient *rpc.Client) {
	//点对点消息
	if msg.Type == common.MESSAGE_TYPE_USER {
		//获取发送者消息空间ID
		senderMaxId := getUserMsgId(msg.From)

		mFrom := <-msgGet
		mFrom.Mid = senderMaxId
		mFrom.Uid = msg.From
		mFrom.Content = msg.Content
		mFrom.Type = msg.Type
		mFrom.Time = msg.Time
		mFrom.From = msg.From
		mFrom.To = msg.To
		mFrom.Group = msg.Group

		//获取接收者消息空间ID
		recMaxId := getUserMsgId(msg.To)
		mTo := <-msgGet
		mTo.Mid = recMaxId
		mTo.Uid = msg.To
		mTo.Content = msg.Content
		mTo.Type = msg.Type
		mTo.Time = msg.Time
		mTo.From = msg.From
		mTo.To = msg.To
		mTo.Group = msg.Group

		//接收者消息发送给目的地push srv
		msgPushPool <- mTo

		//将生成的消息给MS存储
		msgStorePool <- mFrom
		msgStorePool <- mTo
	} else if msg.Type == common.MESSAGE_TYPE_GROUP {
		fmt.Printf("%v\n", *msg)
		//群组消息
		//获取群组成员
		args := GroupArgs{
			GroupId: msg.Group,
		}
		var members []int
		err := msClient.Call("MS.GetGroupMembers", args, &members)
		if err != nil || members == nil || len(members) == 0 {
			return
		}

		//获取发送者消息空间ID
		senderMaxId := getUserMsgId(msg.From)
		mFrom := <-msgGet
		mFrom.Mid = senderMaxId
		mFrom.Uid = msg.From
		mFrom.Content = msg.Content
		mFrom.Type = msg.Type
		mFrom.Time = msg.Time
		mFrom.From = msg.From
		mFrom.To = msg.To
		mFrom.Group = msg.Group

		msgStorePool <- mFrom
		for i := 0; i < len(members); i++ {
			//获取接收者消息空间ID
			if members[i] == msg.From {
				continue
			}

			recMaxId := getUserMsgId(members[i])
			mTo := <-msgGet
			mTo.Mid = recMaxId
			mTo.Uid = members[i]
			mTo.Content = msg.Content
			mTo.Type = msg.Type
			mTo.Time = msg.Time
			mTo.From = msg.From
			mTo.To = members[i]
			mTo.Group = msg.Group

			//接收者消息发送给目的地push srv
			msgPushPool <- mTo

			//存储消息
			msgStorePool <- mTo
		}
	} else if msg.Type == common.MESSAGE_TYPE_SUB {

		//接收者消息发送给目的地push srv
		msgPushPool <- msg
	} else if msg.Type == common.MESSAGE_TYPE_PUBLIC {
		//公开消息
		//接收者消息发送给目的地push srv
		msgPushPool <- msg
	}
}
