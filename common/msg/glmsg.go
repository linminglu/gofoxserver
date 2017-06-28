package msg

//逻辑服和大厅服消息

//房间简要信息
type RoomInfo struct {
	ServerID   int              //第二类型
	KindID     int              //第一类型
	RoomID     int              //6位房号
	NodeID     int              //在哪个节点上
	CurCnt     int              //当前人数
	MaxCnt     int              //最多多人数
	PayCnt     int              //可玩局数
	PayType    int              //支付类型
	CurPayCnt  int              //已玩局数
	CreateTime int64            //创建时间
	Idx        int              //服务器标识用的字段
	PlayerIds  map[int]struct{} //玩家id
}

///通知大厅房间结束
type RoomEndInfo struct {
	RoomId int //房间id
}