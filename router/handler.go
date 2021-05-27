package router

import (
	ChatRoom "hotel/chatRoom"

	"fmt"
	"strconv"
	"time"

	"gopkg.in/olahol/melody.v1"
)

var CmdMap = map[string]func(s *melody.Session, msg string) error{
	"info":     infoHandler,
	"室友":       roommateHandler,
	"time":     checkOutTimeHandler,
	"addmoney": addMoneyHandler,
	"help":     helpHandler,
}

// handler.go  ------------ ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

// 查詢用戶資訊
func infoHandler(s *melody.Session, msg string) error {
	id := s.Request.URL.Query().Get("id")                // 名字
	user, err := ChatRoom.DefaultRoomManager.UUIDMap[id] // 使用者資料
	if !err {
		return fmt.Errorf("找不到旅客資料")
	}
	money := strconv.Itoa(user.Money)                             // 使用者的金錢
	roomID := strconv.Itoa(user.Room.RoomID)                      // 使用者的房間
	checkInTime := user.CheckInTime.Format("2006-01-02 15:04:05") // 使用者入住的時間

	// 將要傳送的文字組合
	text := "名字:" + id + ", 金錢:" + money + ", 房間:" + roomID + ", 入住時間:" + checkInTime

	ChatRoom.SentMessageTo(id, nil, text, "user")

	return nil
}

// 查詢用戶房間內所有成員
func roommateHandler(s *melody.Session, msg string) error {
	id := s.Request.URL.Query().Get("id")                // 名字
	user, err := ChatRoom.DefaultRoomManager.UUIDMap[id] // 使用者資料
	if !err {
		return fmt.Errorf("找不到旅客資料")
	}
	room := user.Room                   // 使用者所在的房間
	roomID := strconv.Itoa(room.RoomID) // 使用者的房間名稱
	var roomMateList string
	for key, _ := range room.ToristMap {
		if id == key.Name {
			roomMateList += key.Name + "(我), "
		} else {
			roomMateList += key.Name + ", "
		}
	}

	// 將要傳送的文字組合
	text := "這間是:" + roomID + ", 室友名單:" + roomMateList
	ChatRoom.SentMessageTo(id, nil, text, "user")

	return nil
}

// 查詢退房時間與剩餘秒數
func checkOutTimeHandler(s *melody.Session, msg string) error {
	id := s.Request.URL.Query().Get("id")                // 名字
	user, err := ChatRoom.DefaultRoomManager.UUIDMap[id] // 使用者資料
	if !err {
		return fmt.Errorf("找不到旅客資料")
	}
	roomCheckInTime := user.CheckInTime   // 使用者入住時間
	roomCheckOutTime := user.CheckOutTime // 使用者入住時間
	subM := roomCheckOutTime.Sub(time.Now())
	var text string
	// fmt.Println(subM.Seconds(), "秒")
	if subM.Seconds() > 0 {
		text = fmt.Sprint("入住時間："+roomCheckInTime.Format("2006-01-02 15:04:05")+", 退房時間："+roomCheckOutTime.Format("2006-01-02 15:04:05"), ", 剩餘時間", int(subM.Seconds()), "秒")
	} else {
		text = fmt.Sprint("入住時間："+roomCheckInTime.Format("2006-01-02 15:04:05")+", 退房時間："+roomCheckOutTime.Format("2006-01-02 15:04:05"), ", 已經超過時間了")
	}

	ChatRoom.SentMessageTo(id, nil, text, "user")
	return nil
}

// 增加金錢
func addMoneyHandler(s *melody.Session, msg string) error {
	id := s.Request.URL.Query().Get("id")                // 名字
	user, err := ChatRoom.DefaultRoomManager.UUIDMap[id] // 使用者資料
	if !err {
		return fmt.Errorf("找不到旅客資料")
	}
	user.AddMoney(1000)
	money := strconv.Itoa(user.Money) // 使用者的金錢

	// 將要傳送的文字組合
	text := "金錢已增加1000, 名字:" + id + ", 金錢:" + money
	ChatRoom.SentMessageTo(id, nil, text, "user")

	return nil
}

// 查詢指令
func helpHandler(s *melody.Session, msg string) error {
	id := s.Request.URL.Query().Get("id") // 名字
	info := "/info 查看自己資料,<br>"
	roomMate := "/室友 查看自己房間內室友,<br>"
	time := "/time 查看剩餘時間,時間到會提醒退房,<br>"
	addmoney := "/addmoney 增加金錢1000,<br>"

	text := fmt.Sprintf("%s%s%s%s", info, roomMate, time, addmoney)
	ChatRoom.SentMessageTo(id, nil, text, "user")

	return nil
}

// handler.go  ------------ 。
