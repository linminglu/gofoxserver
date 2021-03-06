package room

import (
	"mj/gameServer/RoomMgr"
	"mj/gameServer/common"
	"mj/gameServer/common/mj/mj_base"

	"mj/gameServer/common/room_base"
	"mj/gameServer/db/model/base"

	"mj/common/msg"

	"github.com/lovelly/leaf/log"
)

func CreaterRoom(args []interface{}) RoomMgr.IRoom {
	info := args[0].(*msg.L2G_CreatorRoom)
	if info.KindId != common.KIND_TYPE_HZMJ {
		log.Error("at CreaterRoom info.KindId != common.KIND_TYPE_HZMJ uid:%d", info.CreatorUid)
		return nil
	}

	temp, ok := base.GameServiceOptionCache.Get(info.KindId, info.ServiceId)
	if !ok {
		log.Error("at CreaterRoom not foud template kind:%d, serverId:%d, uid:%d", info.KindId, info.ServiceId, info.CreatorUid)
		return nil
	}

	r := NewHZEntry(info.KindId, info.ServiceId)
	rbase := room_base.NewRoomBase()
	cfg := &mj_base.NewMjCtlConfig{
		BaseMgr:  rbase,
		DataMgr:  NewHZDataMgr(info.RoomID, info.CreatorUid, mj_base.IDX_HZMJ, "", temp, r, info),
		UserMgr:  room_base.NewRoomUserMgr(info, temp),
		LogicMgr: NewHZlogic(mj_base.IDX_HZMJ),
		TimerMgr: room_base.NewRoomTimerMgr(info.PlayCnt, temp, rbase.GetSkeleton()),
	}
	r.Init(cfg)

	if cfg.BaseMgr == nil || cfg.DataMgr == nil || cfg.UserMgr == nil || cfg.LogicMgr == nil || cfg.TimerMgr == nil {
		log.Error("at CreaterRoom mermber faild kind:%d, RoomID:%d uid:%d", info.KindId, info.RoomID, info.CreatorUid)
		return nil
	}
	if r == nil {
		log.Error("at CreaterRoom NewMJBase error, uid:%d", info.CreatorUid)
		return nil
	}

	RegisterHandler(r)
	return r
}
