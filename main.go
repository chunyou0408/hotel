package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"gopkg.in/olahol/melody.v1"
)

// 設定訊息物件
type Message struct {
	Event   string `json:"event"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// 設定訊息方法
func NewMessage(event, name, content string) *Message {
	return &Message{
		Event:   event,
		Name:    name,
		Content: content,
	}
}

//由於透過 WebSocket 傳送訊息要使用 []byte 格式，因此這邊我們也將轉換的方法進行封裝
func (m *Message) GetByteMessage() []byte {
	result, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err)
	}
	return result
}

var server *melody.Melody

func main() {
	InitManager() //初始化控制中心

	server = melody.New()

	r := gin.Default()
	r.LoadHTMLGlob("template/html/*")
	r.Static("/assets", "./template/assets")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	//接著我們使用 gin 來處理 routing，透過 melody 來處理連線的 request。
	r.GET("/ws", func(c *gin.Context) {
		server.HandleRequest(c.Writer, c.Request)
	})

	// 連線開始
	server.HandleConnect(firstConnect)
	// 收到訊息
	server.HandleMessage(getMessage)

	// 連線結束
	// melody 有提供 HandleClose 方法讓我們處理離線的 session，
	// 我們這邊設定離線就發送一個 xxx 離開聊天室 的訊息給全部人
	server.HandleClose(sendLeaveRoom)

	// 監聽
	r.Run(":5000")
}

func remindToCheckOut(id string, count int) {
	// 計時30秒,30秒後傳訊息給旅客,提醒退房
	const KEY = "user_id"
	DefaultRoomManager.UUIDMap[id].checkOutTime = DefaultRoomManager.UUIDMap[id].checkInTime.Add(time.Duration(count) * time.Second)
	timer := time.NewTimer(time.Duration(count) * time.Second)

	// Current time
	// now := time.Now()
	// fmt.Printf("time : %v.\n", now)

	<-timer.C
	// fmt.Printf("time : %v.\n", expire)
	server.BroadcastFilter(NewMessage("other", id, id+"退房時間到了").GetByteMessage(), func(session *melody.Session) bool {
		compareID, _ := session.Get(KEY)
		return compareID == "user_id" || compareID == id
	})
}

// 收到訊息後處理
func getMessage(s *melody.Session, msg []byte) {
	// 取得名字
	id := s.Request.URL.Query().Get("id")
	// 顯示內容
	cmd := gjson.Get(string(msg), "content").String()
	fmt.Println(cmd)

	//如果第一個是斜線
	if string(cmd[0]) == "/" {
		const KEY = "user_id"
		commend := string(cmd[1:len(string(cmd))])
		// 執行指令
		if commend == "info" {

			DefaultRoomManager.infoHandler(s, KEY)

		} else if commend == "室友" {

			DefaultRoomManager.roommateHandler(s, KEY)

		} else if commend == "time" {

			DefaultRoomManager.checkOutTimeHandler(s, KEY)

		} else if commend == "addmoney" {

			DefaultRoomManager.addMoneyHandler(s, KEY)

		} else if commend == "help" {

			DefaultRoomManager.helpHandler(s, KEY)

		} else {
			server.BroadcastFilter(NewMessage("other", id, "指令錯誤,可輸入/help查看指令").GetByteMessage(), func(session *melody.Session) bool {
				compareID, _ := session.Get(KEY)
				return compareID == "user_id" || compareID == id
			})
		}
	} else {
		// server.Broadcast(msg)
		const KEY = "chat_id"
		// 查傳資料的房客在哪間房間
		r := DefaultRoomManager.findUserRoom(id)
		roomID := strconv.Itoa(r.roomID)
		id := "room_" + roomID
		server.BroadcastFilter(msg, func(session *melody.Session) bool {
			compareID, _ := session.Get(KEY)
			return compareID == "chat_id" || compareID == id
		})
	}
}

// 第一次連線處理
func firstConnect(session *melody.Session) {

	// 取得名字
	id := session.Request.URL.Query().Get("id")

	// 有人進入就登記他的id
	const KEY = "user_id"
	session.Set(KEY, id)

	// 預設值金錢
	money := 1000
	// 現在時間
	t := time.Now().Format("2006-01-02 15:04:05")

	// 當旅館滿人進入
	if len(DefaultRoomManager.UUIDMap) >= 8 {
		fmt.Println("目前人數", len(DefaultRoomManager.UUIDMap), "超過人數(最多8人),無法入住")
		server.BroadcastFilter(NewMessage("other", id, "因為客滿被踢出").GetByteMessage(), func(session *melody.Session) bool {
			compareID, _ := session.Get(KEY)
			return compareID == "user_id" || compareID == id
		})
		time.Sleep(time.Millisecond * 300)
		session.Close()
		return
	}

	// 當有人連線時,將資料寫入旅客名單
	tt := NewTourist(id, money, session)

	// 交給控制中心,新增使用者
	DefaultRoomManager.SignNewMember(tt)
	// 加入房間的工作
	room := DefaultRoomManager.work(tt)
	// 以上成功之後顯示

	server.Broadcast(NewMessage("other", id, "加入聊天室,房間號碼："+room+",時間："+t).GetByteMessage())

	// 提醒退房,第二個是秒數
	go remindToCheckOut(id, 30)

}

// 離線時處理
func sendLeaveRoom(session *melody.Session, i int, s string) error {

	id := session.Request.URL.Query().Get("id")
	const KEY = "chat_id"

	if DefaultRoomManager.findUser(id) {
		room := "room_" + strconv.Itoa(DefaultRoomManager.UUIDMap[id].room.roomID)

		server.BroadcastFilter(NewMessage("other", id, "離開聊天室").GetByteMessage(), func(session *melody.Session) bool {
			compareID, _ := session.Get(KEY)
			return compareID == "chat_id" || compareID == room
		})

	} else {
		// server.Broadcast(NewMessage("other", id, "因為客滿而離開聊天室").GetByteMessage())
	}

	DefaultRoomManager.SignOutMember(id)
	return nil
}
