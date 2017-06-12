package internal

import (
	"mj/gameServer/base"

	"mj/gameServer/user"

	"github.com/lovelly/leaf/gate"
	"github.com/lovelly/leaf/log"
	"github.com/lovelly/leaf/module"
)

type ChatRoom struct {
	members      map[int]gate.Agent
	memChatIndex map[int]int //成员聊天索引
}

var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer
	roomList = make(map[int]*ChatRoom)
	roomID   int
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton

}

func (m *Module) OnDestroy() {

}

//创建聊天房间
func createRoom(args []interface{}) (interface{}, error) {
	ag := args[0].(gate.Agent)
	user := ag.UserData().(*user.User)
	room := &ChatRoom{members: make(map[int]gate.Agent), memChatIndex: make(map[int]int)}
	room.members[user.Id] = ag
	roomList[roomID] = room
	room.memChatIndex[user.Id] = 0
	user.ChatRoomId = roomID
	roomID++
	return roomID, nil
}

//增加聊天房间成员
func addRoomMember(args []interface{}) {
	roomID := args[0].(int)
	ag := args[1].(gate.Agent)
	room, ok := roomList[roomID]
	if !ok {
		log.Error("聊天房间：%s不存在", roomID)
		return
	}
	user := ag.UserData().(*user.User)
	room.members[user.Id] = ag
	room.memChatIndex[user.Id] = 0
}

//删除聊天房间成员
func delRoomMember(args []interface{}) {
	GetRoomID := args[0].(int)
	UserID := args[1].(int)
	room, ok := roomList[GetRoomID]
	if !ok {
		log.Error("聊天房间：%s不存在", GetRoomID)
		return
	}

	size := len(room.members)
	if size > 1 {
		delete(room.members, UserID)
		delete(room.memChatIndex, UserID)
	} else {
		delete(room.memChatIndex, UserID)
		delete(room.members, UserID)
		delete(roomList, GetRoomID)
	}
}

func SendMsgToUser(getRoomID int, userID int, data interface{}) {
	room, ok := roomList[getRoomID]
	if !ok {
		log.Error("聊天房间：%s不存在", getRoomID)
		return
	}
	for id, ag := range room.members {
		if id == userID {
			ag.WriteMsg(data)
			return
		}
	}
}

func SendMsgToAll(getRoomID int, data interface{}) {
	room, ok := roomList[getRoomID]
	if !ok {
		log.Error("聊天房间：%s不存在", getRoomID)
		return
	}

	for _, ag := range room.members {
		ag.WriteMsg(data)
		user := ag.UserData().(*user.User)
		room.memChatIndex[user.Id]++
	}
}