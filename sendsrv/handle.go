package main

import (
	"gim/common"
	"strconv"
)

const (
	USER_MAX_MSGID_PREFIX = "maxmsgid#"
)

type GroupArgs struct {
	groupId int
}

//获取用户分配的消息id
func getUserMsgId(userId int) int {
	maxId, _ := redClient.Do("INCR", USER_MAX_MSGID_PREFIX+strconv.Itoa(userId))
	val, ok := maxId.(int)
	if ok {
		return val
	}
	return 0
}

//发过来消息字符串
func HandleServerMsg(ss *SendSrv, msg *common.Message) {
	//点对点消息
	if msg.Type == common.MESSAGE_TYPE_USER {
		//获取发送者消息空间ID
		senderMaxId := getUserMsgId(msg.From)
		mFrom := common.Message{
			Mid:     senderMaxId,
			Uid:     msg.From,
			Content: msg.Content,
			Type:    msg.Type,
			Time:    msg.Time,
			From:    msg.From,
			To:      msg.To,
			Group:   msg.Group,
		}
		//获取接收者消息空间ID
		recMaxId := getUserMsgId(msg.To)
		mTo := common.Message{
			Mid:     recMaxId,
			Uid:     msg.To,
			Content: msg.Content,
			Type:    msg.Type,
			Time:    msg.Time,
			From:    msg.From,
			To:      msg.To,
			Group:   msg.Group,
		}
		//接收者消息发送给目的地push srv
		var reply bool
		err := pushSrvClient.Call("PushSrv.SendMsg", mTo, &reply)

		//将生成的消息给MS存储
		ss.msgPool <- &mFrom
		ss.msgPool <- &mTo
	} else if msg.Type == common.MESSAGE_TYPE_GROUP {
		//群组消息
		//获取群组成员
		args := GroupArgs{
			groupId: msg.Group,
		}
		members := make([]int, 100)
		err := msClient.Call("MS.GetGroupMembers", args, &members)
		if err != nil || len(members) == 0 {
			return
		}

		//获取发送者消息空间ID
		senderMaxId := getUserMsgId(msg.From)
		mFrom := common.Message{
			Mid:     senderMaxId,
			Uid:     msg.From,
			Content: msg.Content,
			Type:    msg.Type,
			Time:    msg.Time,
			From:    msg.From,
			To:      msg.To,
			Group:   msg.Group,
		}
		ss.msgPool <- &mFrom
		for i := 0; i < len(members); i++ {
			//获取接收者消息空间ID
			recMaxId := getUserMsgId(members[i])
			mTo := common.Message{
				Mid:     recMaxId,
				Uid:     members[i],
				Content: msg.Content,
				Type:    msg.Type,
				Time:    msg.Time,
				From:    msg.From,
				To:      members[i],
				Group:   msg.Group,
			}

			//接收者消息发送给目的地push srv
			var reply bool
			err := pushSrvClient.Call("PushSrv.SendMsg", mTo, &reply)

			//存储消息
			ss.msgPool <- &mTo
		}
	} else if msg.Type == common.MESSAGE_TYPE_SUB {

		//接收者消息发送给目的地push srv
		var reply bool
		err := pushSrvClient.Call("PushSrv.SendMsg", *msg, &reply)

	} else if msg.Type == common.MESSAGE_TYPE_PUBLIC {
		//公开消息
		//接收者消息发送给目的地push srv
		var reply bool
		err := pushSrvClient.Call("PushSrv.SendMsg", *msg, &reply)
	}
}
