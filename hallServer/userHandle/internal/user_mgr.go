package internal

import (
	"mj/hallServer/center"
	"mj/hallServer/user"
	"sync"

	"github.com/lovelly/leaf/log"
)

var (
	Users     = make(map[int64]*user.User) //key is userId
	UsersLock sync.RWMutex
)

//此api 尽量少用
func ForEachUser(f func(u *user.User)) {
	UsersLock.RLock()
	defer UsersLock.RUnlock()
	for _, u := range Users {
		if u != nil {
			f(u)
		}
	}
}

//此函数不到处  要跟user 联络请用center
func getUser(uid int64) *user.User {
	UsersLock.RLock()
	defer UsersLock.RUnlock()
	u, _ := Users[uid]
	return u
}

func HasUser(uid int64) bool {
	UsersLock.RLock()
	defer UsersLock.RUnlock()
	_, ok := Users[uid]
	return ok
}

func AddUser(uid int64, u *user.User) {
	log.Debug("AddUser: %d =======================", uid)
	center.ChanRPC.Go("SelfNodeAddPlayer", uid, u.ChanRPC())
	UsersLock.Lock()
	defer UsersLock.Unlock()
	Users[uid] = u
}

func DelUser(uid int64) {
	log.Debug("deluser %d ======================", uid)
	center.ChanRPC.Go("SelfNodeDelPlayer", uid)
	UsersLock.Lock()
	defer UsersLock.Unlock()
	delete(Users, uid)
}
