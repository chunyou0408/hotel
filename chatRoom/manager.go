package chatRoom

import (
	"fmt"
	"strconv"
	"sync"
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

	fmt.Println("已將客人加入房號：", r.RoomID, ",此房旅客名單：")
	for key, _ := range r.ToristMap {
		fmt.Println(key.Name)
	}
}

// 建立全新的房間
func (m *Manager) CreateRoom() *Room {

	r := &Room{
		RoomID:       len(m.roomsM) + 1,
		EmptyBeds:    4,
		UsingBeds:    0,
		TorisMapLock: new(sync.RWMutex),       // 旅客的MAP鎖
		ToristMap:    make(map[*Tourist]bool), // 旅客的MAP
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
		ma.UUIDMap[t.UuIdentity] = t
		ma.UUIDRWLocker.Unlock()

		fmt.Println("目前旅客有", len(ma.UUIDMap), "人")
	}
	return nil
}

// SignOutMember 新的使用者連線從UUID清除
func (m *Manager) SignOutMember(UUID string) error {
	if DefaultRoomManager.UUIDMap[UUID] != nil {
		tt := DefaultRoomManager.UUIDMap[UUID]
		r := DefaultRoomManager.UUIDMap[UUID].Room

		r.CheckOut(tt)

		DefaultRoomManager.UUIDRWLocker.Lock()
		delete(DefaultRoomManager.UUIDMap, UUID)
		DefaultRoomManager.UUIDRWLocker.Unlock()
	}
	return nil
}

func (m *Manager) Work(tt *Tourist) string {
	// fmt.Println("控制中心接待" + tt.name + "中")
	// fmt.Println("開始判斷人數是否超過40")
	// fmt.Println("低於40人,可入住")
	// fmt.Println("判斷所有房間是否有空位,加入房間")
	// fmt.Println("若已滿將新增房間")

	for key, _ := range m.roomsM {
		if key.EmptyBeds > 0 {
			fmt.Println("還有空位不用新增")

			// 控制中心.加入房間(房間,旅客)
			m.JoinRoom(key, tt)

			return strconv.Itoa(key.RoomID)
		}
	}
	fmt.Println("新增房間")
	r := m.CreateRoom()
	fmt.Println("目前房間數量有", len(m.roomsM), "間")
	r.CheckIn(tt)
	return strconv.Itoa(r.RoomID)
}

// 查詢使用者是不是有被新增過
func (m *Manager) FindUser(s string) bool {

	if m.UUIDMap[s] == nil {
		return false
	}
	return true
}

// 查詢是不是有在房間內
func (m *Manager) findUserRoom(s string) *Room {

	if m.UUIDMap[s] != nil {
		return m.UUIDMap[s].Room
	}
	return nil
}
