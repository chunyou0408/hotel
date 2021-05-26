package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"gopkg.in/olahol/melody.v1"
)

// Manager 控制中心
type Manager struct {
	roomsM       map[*Room]bool      // 房間的MAP
	roomsMapLock *sync.RWMutex       // 使用Map記得要一起用Lock
	UUIDRWLocker *sync.RWMutex       // 使用者UUID清單讀寫鎖
	UUIDMap      map[string]*Tourist // 使用者清單，key:uuid，value:使用者資料
}

// DefaultRoomManager 預設的房間控制中心變數
var DefaultRoomManager Manager

var cmdMap = map[string]func(s *melody.Session, msg string) error{
	"info":     infoHandler,
	"室友":       roommateHandler,
	"time":     checkOutTimeHandler,
	"addmoney": addMoneyHandler,
	"help":     helpHandler,
}

// 初始化一個管理員
func InitManager() {
	DefaultRoomManager = Manager{
		roomsMapLock: new(sync.RWMutex),
		roomsM:       make(map[*Room]bool),
		UUIDRWLocker: new(sync.RWMutex),
		UUIDMap:      make(map[string]*Tourist),
	}

}

// 協助旅客加入房間
func (m *Manager) JoinRoom(r *Room, tt *Tourist) {
	r.CheckIn(tt)

	fmt.Println("已將客人加入房號：", r.roomID, ",此房旅客名單：")
	for key, _ := range r.toristMap {
		fmt.Println(key.name)
	}
}

// 建立全新的房間
func (m *Manager) CreateRoom() *Room {

	r := &Room{
		roomID:       len(m.roomsM) + 1,
		emptyBeds:    4,
		usingBeds:    0,
		torisMapLock: new(sync.RWMutex),       // 旅客的MAP鎖
		toristMap:    make(map[*Tourist]bool), // 旅客的MAP
	}

	m.roomsMapLock.Lock()
	m.roomsM[r] = true
	m.roomsMapLock.Unlock()

	return r
}

// SignNewMember 紀錄新的使用者在清單裡面 // TODO test
func (ma *Manager) SignNewMember(t *Tourist) error {
	if t != nil {
		ma.UUIDRWLocker.Lock()
		ma.UUIDMap[t.uuIdentity] = t
		ma.UUIDRWLocker.Unlock()

		fmt.Println("目前旅客有", len(ma.UUIDMap), "人")
	}
	return nil
}

// SignOutMember 新的使用者連線從UUID清除
func (m *Manager) SignOutMember(UUID string) error {
	if DefaultRoomManager.UUIDMap[UUID] != nil {
		tt := DefaultRoomManager.UUIDMap[UUID]
		r := DefaultRoomManager.UUIDMap[UUID].room

		r.CheckOut(tt)

		DefaultRoomManager.UUIDRWLocker.Lock()
		delete(DefaultRoomManager.UUIDMap, UUID)
		DefaultRoomManager.UUIDRWLocker.Unlock()
	}
	return nil
}

func (m *Manager) work(tt *Tourist) string {
	// fmt.Println("控制中心接待" + tt.name + "中")
	// fmt.Println("開始判斷人數是否超過40")
	// fmt.Println("低於40人,可入住")
	// fmt.Println("判斷所有房間是否有空位,加入房間")
	// fmt.Println("若已滿將新增房間")

	for key, _ := range m.roomsM {
		if key.emptyBeds > 0 {
			fmt.Println("還有空位不用新增")

			// 控制中心.加入房間(房間,旅客)
			m.JoinRoom(key, tt)

			return strconv.Itoa(key.roomID)
		}
	}
	fmt.Println("新增房間")
	r := m.CreateRoom()
	fmt.Println("目前房間數量有", len(m.roomsM), "間")
	r.CheckIn(tt)
	return strconv.Itoa(r.roomID)
}

// 查詢使用者是不是有被新增過
func (m *Manager) findUser(s string) bool {

	if m.UUIDMap[s] == nil {
		return false
	}
	return true
}

// 查詢是不是有在房間內
func (m *Manager) findUserRoom(s string) *Room {

	if m.UUIDMap[s] != nil {
		return m.UUIDMap[s].room
	}
	return nil
}

// 查詢用戶資訊
func infoHandler(s *melody.Session, msg string) error {
	id := s.Request.URL.Query().Get("id")       // 名字
	user, err := DefaultRoomManager.UUIDMap[id] // 使用者資料
	if !err {
		return fmt.Errorf("找不到旅客資料")
	}
	money := strconv.Itoa(user.money)                             // 使用者的金錢
	roomID := strconv.Itoa(user.room.roomID)                      // 使用者的房間
	checkInTime := user.checkInTime.Format("2006-01-02 15:04:05") // 使用者入住的時間

	// 將要傳送的文字組合
	text := "名字:" + id + ", 金錢:" + money + ", 房間:" + roomID + ", 入住時間:" + checkInTime

	sentMessageTo(id, nil, text, "user")

	return nil
}

// 查詢用戶房間內所有成員
func roommateHandler(s *melody.Session, msg string) error {
	id := s.Request.URL.Query().Get("id")       // 名字
	user, err := DefaultRoomManager.UUIDMap[id] // 使用者資料
	if !err {
		return fmt.Errorf("找不到旅客資料")
	}
	room := user.room                   // 使用者所在的房間
	roomID := strconv.Itoa(room.roomID) // 使用者的房間名稱
	var roomMateList string
	for key, _ := range room.toristMap {
		if id == key.name {
			roomMateList += key.name + "(我), "
		} else {
			roomMateList += key.name + ", "
		}
	}

	// 將要傳送的文字組合
	text := "這間是:" + roomID + ", 室友名單:" + roomMateList
	sentMessageTo(id, nil, text, "user")

	return nil
}

// 查詢退房時間與剩餘秒數
func checkOutTimeHandler(s *melody.Session, msg string) error {
	id := s.Request.URL.Query().Get("id")       // 名字
	user, err := DefaultRoomManager.UUIDMap[id] // 使用者資料
	if !err {
		return fmt.Errorf("找不到旅客資料")
	}
	roomCheckInTime := user.checkInTime   // 使用者入住時間
	roomCheckOutTime := user.checkOutTime // 使用者入住時間
	subM := roomCheckOutTime.Sub(time.Now())
	var text string
	// fmt.Println(subM.Seconds(), "秒")
	if subM.Seconds() > 0 {
		text = fmt.Sprint("入住時間："+roomCheckInTime.Format("2006-01-02 15:04:05")+", 退房時間："+roomCheckOutTime.Format("2006-01-02 15:04:05"), ", 剩餘時間", int(subM.Seconds()), "秒")
	} else {
		text = fmt.Sprint("入住時間："+roomCheckInTime.Format("2006-01-02 15:04:05")+", 退房時間："+roomCheckOutTime.Format("2006-01-02 15:04:05"), ", 已經超過時間了")
	}

	sentMessageTo(id, nil, text, "user")
	return nil
}

// 增加金錢
func addMoneyHandler(s *melody.Session, msg string) error {
	id := s.Request.URL.Query().Get("id")       // 名字
	user, err := DefaultRoomManager.UUIDMap[id] // 使用者資料
	if !err {
		return fmt.Errorf("找不到旅客資料")
	}
	user.addMoney(1000)
	money := strconv.Itoa(user.money) // 使用者的金錢

	// 將要傳送的文字組合
	text := "金錢已增加1000, 名字:" + id + ", 金錢:" + money
	sentMessageTo(id, nil, text, "user")

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
	sentMessageTo(id, nil, text, "user")

	return nil
}
