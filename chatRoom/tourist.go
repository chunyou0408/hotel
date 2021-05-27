package chatRoom

import (
	"time"

	"gopkg.in/olahol/melody.v1"
)

// Tourist 旅人
type Tourist struct {
	Name         string // 名字
	Money        int    // 錢
	Room         *Room  // 住哪間房間
	UuIdentity   string // uuid識別子
	CheckInTime  time.Time
	CheckOutTime time.Time
	Session      *melody.Session
}

// 新的旅客
func NewTourist(id string, money int, session *melody.Session) *Tourist {

	t := &Tourist{
		Name:        id,
		Money:       1000,
		UuIdentity:  id,
		CheckInTime: time.Now(),
		Session:     session,
	}

	return t
}

// 新增金錢
func (t *Tourist) AddMoney(money int) {
	t.Money += 1000
}
