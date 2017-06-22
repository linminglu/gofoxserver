package mj_base

import (
	. "mj/common/cost"
	"mj/common/msg"
	"mj/common/msg/mj_hz_msg"
	"mj/gameServer/common"
	"mj/gameServer/conf"
	"mj/gameServer/db/model"
	"mj/gameServer/db/model/base"
	"mj/gameServer/user"
	"time"

	"github.com/lovelly/leaf/log"
)

type NewUserMgrFunc func(int, int, *base.GameServiceOption) common.UserManager
type NewDataMgrFunc func(int, int, int, int, *base.GameServiceOption) common.DataManager
type NewBaseMgrFunc func(int, int, int, int, *base.GameServiceOption) common.BaseManager
type NewLogicMgrFunc func(int, int, int, int, *base.GameServiceOption) common.LogicManager
type NewTimerMgrFunc func(int, int, int, int, *base.GameServiceOption) common.TimerManager

type Mj_base struct {
	common.BaseManager
	DataMgr  common.DataManager
	UserMgr  common.UserManager
	LogicMgr common.LogicManager
	TimerMgr common.TimerManager

	Temp   *base.GameServiceOption //模板
	Status int
}

type NewMjCtlConfig struct {
	NUserF  NewUserMgrFunc
	NDataF  NewDataMgrFunc
	NBaseF  NewBaseMgrFunc
	NLogicF NewLogicMgrFunc
	NTimerF NewTimerMgrFunc

}

func NewMJBase(roomId, uid int, NewUserMgr NewUserMgrFunc) *Mj_base {
	lk, ok := model.CreateRoomInfoOp.Get(roomId)
	if !ok {
		return nil
	}

	Temp, ok1 := base.GameServiceOptionCache.Get(lk.KindId, lk.ServiceId)
	if !ok1 {
		return nil
	}

	mj := new(Mj_base)
	mj.UserMgr = NewUserMgr(roomId, lk.MaxPlayerCnt, Temp)

	return mj
}

//坐下
func (r *Mj_base) Sitdown(args []interface{}) {
	recvMsg := args[0].(*msg.C2G_UserSitdown)
	u := args[1].(*user.User)

	retcode := 0
	defer func() {
		if retcode != 0 {
			u.WriteMsg(RenderErrorMessage(retcode))
		}
	}()
	if r.Status == RoomStatusStarting && r.Temp.DynamicJoin == 1 {
		retcode = GameIsStart
		return
	}

	retcode = r.UserMgr.Sit(u, recvMsg.ChairID)
}

//起立
func (r *Mj_base) UserStandup(args []interface{}) {
	//recvMsg := args[0].(*msg.C2G_UserStandup{})
	u := args[1].(*user.User)
	retcode := 0
	defer func() {
		if retcode != 0 {
			u.WriteMsg(RenderErrorMessage(retcode))
		}
	}()

	if r.Status == RoomStatusStarting {
		retcode = ErrGameIsStart
		return
	}

	r.UserMgr.Standup(u)
}

//获取对方信息
func (room *Mj_base) GetUserChairInfo(args []interface{}) {
	recvMsg := args[0].(*msg.C2G_REQUserChairInfo)
	u := args[1].(*user.User)
	info := room.UserMgr.GetUserInfoByChairId(recvMsg.ChairID).(*msg.G2C_UserEnter)
	if info == nil {
		log.Error("at GetUserChairInfo no foud tagUser %v, userId:%d", args[0], u.Id)
		return
	}
	u.WriteMsg(info)
}

//解散房间
func (room *Mj_base) DissumeRoom(args []interface{}) {
	u := args[0].(*user.User)
	retcode := 0
	defer func() {
		if retcode != 0 {
			u.WriteMsg(RenderErrorMessage(retcode, "解散房间失败."))
		}
	}()

	if !room.UserMgr.CanOperatorRoom(u.Id) {
		retcode = NotOwner
		return
	}

	room.UserMgr.ForEachUser(func(u *user.User) {
		room.UserMgr.LeaveRoom(u)
	})

	room.OnEventGameConclude(0, nil, GER_DISMISS)
	room.Destroy(room.DataMgr.GetRoomId())
}

//玩家准备
func (room *Mj_base) UserReady(args []interface{}) {
	//recvMsg := args[0].(*msg.C2G_UserReady)
	u := args[1].(*user.User)
	if u.Status == US_READY {
		log.Debug("user status is ready at UserReady")
		return
	}

	room.UserMgr.SetUsetStatus(u, US_READY)
	if room.UserMgr.IsAllReady() {
		room.DataMgr.InitRoom(room.UserMgr.GetMaxPlayerCnt())
		room.DataMgr.StartGame()
		room.Status = RoomStatusStarting
		room.DataMgr.SendGameStart(room.LogicMgr, room.UserMgr)

	}
}

//玩家重登
func (room *Mj_base) UserReLogin(args []interface{}) {
	u := args[0].(*user.User)
	if u.Status == US_READY {
		log.Debug("user status is ready at UserReady")
		return
	}

	room.UserMgr.ReLogin(u, room.Status)
}

//玩家离线
func (room *Mj_base) UserOffline(args []interface{}) {
	u := args[0].(*user.User)
	if u.Status == US_READY {
		log.Debug("user status is ready at UserReady")
		return
	}

	room.UserMgr.SetUsetStatus(u, US_OFFLINE)
	if room.Temp.TimeOffLineCount != 0 {
		t := room.Skeleton().AfterFunc(time.Duration(room.Temp.TimeOffLineCount)*time.Second, func() {
			room.OffLineTimeOut(u)
		})
		room.UserMgr.AddKickOutTimer(u.Id, t)
	} else {
		room.OffLineTimeOut(u)
	}
}

//离线超时踢出
func (room *Mj_base) OffLineTimeOut(u *user.User) {
	room.UserMgr.LeaveRoom(user)
	if room.Status != RoomStatusReady {
		room.OnEventGameConclude(0, nil, GER_DISMISS)
	} else {
		if room.UserMgr.GetCurPlayerCnt() == 0 { //没人了直接销毁
			room.Destroy(room.DataMgr.GetRoomId())
		}
	}
}

//获取房间基础信息
func (room *Mj_base) GetBirefInfo() *msg.RoomInfo {
	msg := &msg.RoomInfo{}
	msg.ServerID = room.Temp.ServerID
	msg.KindID = room.Temp.KindID
	msg.NodeID = conf.Server.NodeId
	msg.RoomID = room.DataMgr.GetRoomId()
	msg.CurCnt = room.UserMgr.GetCurPlayerCnt()
	msg.MaxCnt = room.UserMgr.GetMaxPlayerCnt()   //最多多人数
	msg.PayCnt = room.DataMgr.GetMaxPayCnt()      //可玩局数
	msg.CurPayCnt = room.DataMgr.GetCurPayInt()   //已玩局数
	msg.CreateTime = room.DataMgr.GetCreatrTime() //创建时间
	return msg
}

//游戏配置
func (room *Mj_base) SetGameOption(args []interface{}) {
	recvMsg := args[0].(*msg.C2G_GameOption)
	u := args[1].(*user.User)
	retcode := 0
	defer func() {
		if retcode != 0 {
			u.WriteMsg(RenderErrorMessage(retcode))
		}
	}()

	if u.ChairId == INVALID_CHAIR {
		retcode = ErrNoSitdowm
		return
	}

	AllowLookon := 0
	if u.Status == US_LOOKON {
		AllowLookon = 1
	}
	u.WriteMsg(&msg.G2C_GameStatus{
		GameStatus:  room.Status,
		AllowLookon: AllowLookon,
	})

	room.DataMgr.SendPersonalTableTip(u)

	if room.Status == RoomStatusReady { // 没开始
		room.DataMgr.SendStatusReady(u)
	} else { //开始了
		//把所有玩家信息推送给自己
		room.UserMgr.SendUserInfoToSelf(u)
		room.DataMgr.SendStatusPlay(u, room.UserMgr.GetCurPlayerCnt(), room.UserMgr.GetMaxPlayerCnt(), room.logic)
	}
}

//出牌
func (room *Mj_base) OutCard(args []interface{}) {
	recvMsg := args[0].(*mj_hz_msg.C2G_HZMJ_HZOutCard)
	u := args[1].(*user.User)
	retcode := 0
	defer func() {
		if retcode != 0 {
			u.WriteMsg(RenderErrorMessage(retcode))
		}
	}()

	//效验状态
	if room.Status != RoomStatusStarting {
		log.Error("at OnUserOutCard game status != RoomStatusStarting ")
		retcode = ErrGameNotStart
		return
	}

	//效验参数
	if u.ChairId != room.DataMgr.GetCurrentUser() {
		log.Error("at OnUserOutCard not self out ")
		retcode = ErrNotSelfOut
		return
	}

	if !room.LogicMgr.IsValidCard(recvMsg.CardData) {
		log.Error("at OnUserOutCard IsValidCard card ")
		retcode = NotValidCard
	}

	//删除扑克
	if !room.LogicMgr.RemoveCard(room.DataMgr.GetUserCardIndex(u.ChairId), recvMsg.CardData) {
		log.Error("at OnUserOutCard not have card ")
		retcode = ErrNotFoudCard
		return
	}

	u.UserLimit |= ^LimitChiHu
	u.UserLimit |= ^LimitPeng
	u.UserLimit |= ^LimitGang

	room.DataMgr.NotifySendCard(u, recvMsg.CardData, room.UserMgr, false)

	//响应判断
	bAroseAction := room.DataMgr.EstimateUserRespond(u.ChairId, recvMsg.CardData, EstimatKind_OutCard, room.UserMgr)

	//派发扑克
	if !bAroseAction {
		room.DataMgr.DispatchCardData(room.DataMgr.GetCurrentUser(), false)
	}
	return
}

// 吃碰杠胡各种操作
func (room *Mj_base) UserOperateCard(args []interface{}) {
	recvMsg := args[0].(*mj_hz_msg.C2G_HZMJ_OperateCard)
	u := args[1].(*user.User)
	retcode := 0
	defer func() {
		if retcode != 0 {
			u.WriteMsg(RenderErrorMessage(retcode))
		}
	}()

	if room.DataMgr.GetCurrentUser() == INVALID_CHAIR {
		//效验状态
		if !room.DataMgr.HasOperator(u.ChairId, recvMsg.OperateCode) {
			retcode = ErrNoOperator
			return
		}

		//变量定义
		cbTargetAction, wTargetUser := room.DataMgr.CheckUserOperator(u, room.UserMgr.GetMaxPlayerCnt(), recvMsg, room.LogicMgr)
		if cbTargetAction < 0 {
			log.Debug("wait other user")
			return
		}

		//放弃操作
		if cbTargetAction == WIK_NULL {
			//用户状态
			room.DataMgr.DispatchCardData(room.DataMgr.GetResumeUser(), room.DataMgr.GetGangStatus() != WIK_GANERAL)
			return
		}

		//胡牌操作
		if cbTargetAction == WIK_CHI_HU {
			room.DataMgr.UserChiHu(wTargetUser, room.UserMgr.GetMaxPlayerCnt(), room.LogicMgr)
			return
		}

		//收集牌
		room.DataMgr.WeaveCard(cbTargetAction, wTargetUser)

		//删除扑克
		if room.DataMgr.RemoveCardByOP(wTargetUser, cbTargetAction, room.LogicMgr) {
			return
		}

		room.DataMgr.OperateResult(wTargetUser, cbTargetAction, room.UserMgr, room.LogicMgr)
	} else {
		//扑克效验
		if (recvMsg.OperateCode != WIK_NULL) && (recvMsg.OperateCode != WIK_CHI_HU) && (!room.LogicMgr.IsValidCard(recvMsg.OperateCard[0])) {
			return
		}

		//设置变量
		// room.UserAction[room.CurrentUser] = WIK_NULL

		//执行动作
		switch recvMsg.OperateCode {
		case WIK_GANG: //杠牌操作
			cbGangKind := room.DataMgr.AnGang(u, recvMsg.OperateCode, recvMsg.OperateCard, room.UserMgr, room.LogicMgr)
			//效验动作
			bAroseAction := false
			if cbGangKind == WIK_MING_GANG {
				bAroseAction = room.DataMgr.EstimateUserRespond(u.ChairId, recvMsg.OperateCard[0], EstimatKind_GangCard, room.UserMgr)
			}

			//发送扑克
			if !bAroseAction {
				room.DataMgr.DispatchCardData(u.ChairId, true)
			}
		case WIK_CHI_HU: //自摸
			//结束游戏
			room.DataMgr.ZiMo(u, room.LogicMgr)
			room.OnEventGameConclude(room.DataMgr.GetProvideUser(), nil, GER_NORMAL)
		}
	}
}

//游戏结束
func (room *Mj_base) OnEventGameConclude(ChairId int, user *user.User, cbReason int) {
	switch cbReason {
	case GER_NORMAL: //常规结束
		room.DataMgr.NormalEnd(room.UserMgr, room.LogicMgr, room.Temp)
		room.AfertEnd(false)
	case GER_USER_LEAVE: //用户强退
		if (room.Temp.ServerType & GAME_GENRE_PERSONAL) != 0 { //房卡模式
			return
		}
		//自动托管
		room.DataMgr.OnUserTrustee(user.ChairId, true)
	case GER_DISMISS: //游戏解散
		room.DataMgr.DismissEnd(room.UserMgr, room.LogicMgr)
		room.AfertEnd(true)
	}

	log.Error("at OnEventGameConclude error  ")
	return
}

// 如果这里不能满足 afertEnd 请重构这个到个个组件里面
func (room *Mj_base) AfertEnd(Forced bool) {
	if Forced {
		room.Destroy(room.DataMgr.GetRoomId())
		return
	}

	room.UserMgr.ForEachUser(func(u *user.User) {
		room.UserMgr.SetUsetStatus(u, US_FREE)
	})

	room.TimerMgr.AddPlayCount()
	if room.TimerMgr.GetPlayCount() >= room.Temp.PlayTurnCount {
		room.Destroy(room.DataMgr.GetRoomId())
	}
}