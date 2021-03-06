package stats

import (
	"fmt"
	"mj/hallServer/db"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lovelly/leaf/log"
)

//This file is generate by scripts,don't edit it

//room_log
//

// +gen *
type RoomLog struct {
	RecodeId     int        `db:"recode_id" json:"recode_id"`         // 房间数据记录的Id
	RoomId       int        `db:"room_id" json:"room_id"`             // 房间id
	UserId       int64      `db:"user_id" json:"user_id"`             // 用户索引
	RoomName     string     `db:"room_name" json:"room_name"`         //
	KindId       int        `db:"kind_id" json:"kind_id"`             // 房间索引
	ServiceId    int        `db:"service_id" json:"service_id"`       // 游戏标识
	NodeId       int        `db:"node_id" json:"node_id"`             // 在哪个服务器上
	CreateTime   *time.Time `db:"create_time" json:"create_time"`     // 录入日期
	EndTime      *time.Time `db:"end_time" json:"end_time"`           // 结束日期
	CreateOthers int        `db:"create_others" json:"create_others"` // 是否为他人开房 0否，1是
	PayType      int        `db:"pay_type" json:"pay_type"`           // 支付方式 1是全服 2是AA
	GameEndType  int        `db:"game_end_type" json:"game_end_type"` // 游戏结束类型 0是常规结束 1是游戏解散 2是玩家请求解散 3是没开始就解散
	RoomEndType  int        `db:"room_end_type" json:"room_end_type"` // 解散房间类型 1出错解散房间 2正常解散房间
	NomalOpen    int        `db:"nomal_open" json:"nomal_open"`       // 是否正常开房 0否 1是
}

type roomLogOp struct{}

var RoomLogOp = &roomLogOp{}
var DefaultRoomLog = &RoomLog{}

// 按主键查询. 注:未找到记录的话将触发sql.ErrNoRows错误，返回nil, false
func (op *roomLogOp) Get(recode_id int) (*RoomLog, bool) {
	obj := &RoomLog{}
	sql := "select * from room_log where recode_id=? "
	err := db.StatsDB.Get(obj, sql,
		recode_id,
	)

	if err != nil {
		log.Error("Get data error:%v", err.Error())
		return nil, false
	}
	return obj, true
}
func (op *roomLogOp) SelectAll() ([]*RoomLog, error) {
	objList := []*RoomLog{}
	sql := "select * from room_log "
	err := db.StatsDB.Select(&objList, sql)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return objList, nil
}

func (op *roomLogOp) QueryByMap(m map[string]interface{}) ([]*RoomLog, error) {
	result := []*RoomLog{}
	var params []interface{}

	sql := "select * from room_log where 1=1 "
	for k, v := range m {
		sql += fmt.Sprintf(" and %s=? ", k)
		params = append(params, v)
	}
	err := db.StatsDB.Select(&result, sql, params...)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return result, nil
}

func (op *roomLogOp) GetByMap(m map[string]interface{}) (*RoomLog, error) {
	lst, err := op.QueryByMap(m)
	if err != nil {
		return nil, err
	}
	if len(lst) > 0 {
		return lst[0], nil
	}
	return nil, nil
}

/*
func (i *RoomLog) Insert() error {
    err := db.StatsDBMap.Insert(i)
    if err != nil{
		log.Error("Insert sql error:%v, data:%v", err.Error(),i)
        return err
    }
}
*/

// 插入数据，自增长字段将被忽略
func (op *roomLogOp) Insert(m *RoomLog) (int64, error) {
	return op.InsertTx(db.StatsDB, m)
}

// 插入数据，自增长字段将被忽略
func (op *roomLogOp) InsertTx(ext sqlx.Ext, m *RoomLog) (int64, error) {
	sql := "insert into room_log(room_id,user_id,room_name,kind_id,service_id,node_id,create_time,end_time,create_others,pay_type,game_end_type,room_end_type,nomal_open) values(?,?,?,?,?,?,?,?,?,?,?,?,?)"
	result, err := ext.Exec(sql,
		m.RoomId,
		m.UserId,
		m.RoomName,
		m.KindId,
		m.ServiceId,
		m.NodeId,
		m.CreateTime,
		m.EndTime,
		m.CreateOthers,
		m.PayType,
		m.GameEndType,
		m.RoomEndType,
		m.NomalOpen,
	)
	if err != nil {
		log.Error("InsertTx sql error:%v, data:%v", err.Error(), m)
		return -1, err
	}
	affected, _ := result.LastInsertId()
	return affected, nil
}

//存在就更新， 不存在就插入
func (op *roomLogOp) InsertUpdate(obj *RoomLog, m map[string]interface{}) error {
	sql := "insert into room_log(room_id,user_id,room_name,kind_id,service_id,node_id,create_time,end_time,create_others,pay_type,game_end_type,room_end_type,nomal_open) values(?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE "
	var params = []interface{}{obj.RoomId,
		obj.UserId,
		obj.RoomName,
		obj.KindId,
		obj.ServiceId,
		obj.NodeId,
		obj.CreateTime,
		obj.EndTime,
		obj.CreateOthers,
		obj.PayType,
		obj.GameEndType,
		obj.RoomEndType,
		obj.NomalOpen,
	}
	var set_sql string
	for k, v := range m {
		if set_sql != "" {
			set_sql += ","
		}
		set_sql += fmt.Sprintf(" %s=? ", k)
		params = append(params, v)
	}

	_, err := db.StatsDB.Exec(sql+set_sql, params...)
	return err
}

/*
func (i *RoomLog) Update()  error {
    _,err := db.StatsDBMap.Update(i)
    if err != nil{
		log.Error("update sql error:%v, data:%v", err.Error(),i)
        return err
    }
}
*/

// 用主键(属性)做条件，更新除主键外的所有字段
func (op *roomLogOp) Update(m *RoomLog) error {
	return op.UpdateTx(db.StatsDB, m)
}

// 用主键(属性)做条件，更新除主键外的所有字段
func (op *roomLogOp) UpdateTx(ext sqlx.Ext, m *RoomLog) error {
	sql := `update room_log set room_id=?,user_id=?,room_name=?,kind_id=?,service_id=?,node_id=?,create_time=?,end_time=?,create_others=?,pay_type=?,game_end_type=?,room_end_type=?,nomal_open=? where recode_id=?`
	_, err := ext.Exec(sql,
		m.RoomId,
		m.UserId,
		m.RoomName,
		m.KindId,
		m.ServiceId,
		m.NodeId,
		m.CreateTime,
		m.EndTime,
		m.CreateOthers,
		m.PayType,
		m.GameEndType,
		m.RoomEndType,
		m.NomalOpen,
		m.RecodeId,
	)

	if err != nil {
		log.Error("update sql error:%v, data:%v", err.Error(), m)
		return err
	}

	return nil
}

// 用主键做条件，更新map里包含的字段名
func (op *roomLogOp) UpdateWithMap(recode_id int, m map[string]interface{}) error {
	return op.UpdateWithMapTx(db.StatsDB, recode_id, m)
}

// 用主键做条件，更新map里包含的字段名
func (op *roomLogOp) UpdateWithMapTx(ext sqlx.Ext, recode_id int, m map[string]interface{}) error {

	sql := `update room_log set %s where 1=1 and recode_id=? ;`

	var params []interface{}
	var set_sql string
	for k, v := range m {
		if set_sql != "" {
			set_sql += ","
		}
		set_sql += fmt.Sprintf(" %s=? ", k)
		params = append(params, v)
	}
	params = append(params, recode_id)
	_, err := ext.Exec(fmt.Sprintf(sql, set_sql), params...)
	return err
}

/*
func (i *RoomLog) Delete() error{
    _,err := db.StatsDBMap.Delete(i)
	log.Error("Delete sql error:%v", err.Error())
    return err
}
*/
// 根据主键删除相关记录
func (op *roomLogOp) Delete(recode_id int) error {
	return op.DeleteTx(db.StatsDB, recode_id)
}

// 根据主键删除相关记录,Tx
func (op *roomLogOp) DeleteTx(ext sqlx.Ext, recode_id int) error {
	sql := `delete from room_log where 1=1
        and recode_id=?
        `
	_, err := ext.Exec(sql,
		recode_id,
	)
	return err
}

// 返回符合查询条件的记录数
func (op *roomLogOp) CountByMap(m map[string]interface{}) (int64, error) {

	var params []interface{}
	sql := `select count(*) from room_log where 1=1 `
	for k, v := range m {
		sql += fmt.Sprintf(" and  %s=? ", k)
		params = append(params, v)
	}
	count := int64(-1)
	err := db.StatsDB.Get(&count, sql, params...)
	if err != nil {
		log.Error("CountByMap  error:%v data :%v", err.Error(), m)
		return 0, err
	}
	return count, nil
}

func (op *roomLogOp) DeleteByMap(m map[string]interface{}) (int64, error) {
	return op.DeleteByMapTx(db.StatsDB, m)
}

func (op *roomLogOp) DeleteByMapTx(ext sqlx.Ext, m map[string]interface{}) (int64, error) {
	var params []interface{}
	sql := "delete from room_log where 1=1 "
	for k, v := range m {
		sql += fmt.Sprintf(" and %s=? ", k)
		params = append(params, v)
	}
	result, err := ext.Exec(sql, params...)
	if err != nil {
		return -1, err
	}
	return result.RowsAffected()
}
