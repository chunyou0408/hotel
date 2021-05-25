package main

import (
	"time"

	"gopkg.in/olahol/melody.v1"
)

// Tourist 旅人
type Tourist struct {
	name         string // 名字
	money        int    // 錢
	room         *Room  // 住哪間房間
	uuIdentity   string // uuid識別子
	checkInTime  time.Time
	checkOutTime time.Time
	session      *melody.Session
}

// 新的旅客
func NewTourist(id string, money int, session *melody.Session) *Tourist {

	t := &Tourist{
		name:         id,
		money:        1000,
		uuIdentity:   id,
		checkInTime:  time.Now(),
		session:      session,
	}

	return t
}

// 新增金錢
func (t *Tourist) addMoney() {

}
