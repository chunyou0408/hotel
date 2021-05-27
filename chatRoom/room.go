package chatRoom

import (
	"errors"
	"strconv"
	"sync"
)

// Room 房間
type Room struct {
	RoomID       int               // 房間ID
	EmptyBeds    int               // 空床位
	UsingBeds    int               // 正在使用的床位
	TorisMapLock *sync.RWMutex     // 旅客的MAP鎖
	ToristMap    map[*Tourist]bool // 旅客的MAP
}

// 旅客check in
func (r *Room) CheckIn(t *Tourist) error {
	if t == nil {
		return errors.New("錯誤,沒有旅客")
	}
	const KEY = "chat_id"
	roomID := strconv.Itoa(r.RoomID)
	id := "room_" + roomID
	t.Session.Set(KEY, id)

	r.TorisMapLock.Lock()
	r.ToristMap[t] = true
	t.Room = r
	r.EmptyBeds = r.EmptyBeds - 1
	r.UsingBeds = r.UsingBeds + 1
	r.TorisMapLock.Unlock()
	return nil
}

// 旅客check out
func (r *Room) CheckOut(t *Tourist) error {
	if t == nil {
		return errors.New("錯誤,沒有旅客")
	}

	r.TorisMapLock.Lock()
	delete(r.ToristMap, t)
	r.EmptyBeds = r.EmptyBeds + 1
	r.UsingBeds = r.UsingBeds - 1
	r.TorisMapLock.Unlock()

	return nil
}

// 房間開始運作
func (r *Room) running() {

}
