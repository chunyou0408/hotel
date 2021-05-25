package main

import (
	"strconv"
	"sync"
)

// Room 房間
type Room struct {
	roomID       int               // 房間ID
	emptyBeds    int               // 空床位
	usingBeds    int               // 正在使用的床位
	torisMapLock *sync.RWMutex     // 旅客的MAP鎖
	toristMap    map[*Tourist]bool // 旅客的MAP

	mux sync.RWMutex // 房間的MAP鎖
}

// 旅客check in
func (r *Room) CheckIn(t *Tourist) error {
	if t != nil {

		const KEY = "chat_id"
		aa := strconv.Itoa(r.roomID)
		id := "room10" + aa
		t.session.Set(KEY, id)

		r.torisMapLock.Lock()
		r.toristMap[t] = true
		t.room = r
		r.emptyBeds = r.emptyBeds - 1
		r.usingBeds = r.usingBeds + 1
		r.torisMapLock.Unlock()
	}
	return nil
}

// 旅客check out
func (r *Room) CheckOut(t *Tourist) error {
	if t != nil {
		r.torisMapLock.Lock()
		delete(r.toristMap, t)
		r.emptyBeds = r.emptyBeds + 1
		r.usingBeds = r.usingBeds - 1
		r.torisMapLock.Unlock()
	}
	return nil
}

// 房間開始運作
func (r *Room) running() {

}
