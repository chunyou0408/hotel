package router

import (
	ChatRoom "hotel/chatRoom"

	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"gopkg.in/olahol/melody.v1"
)

var server *melody.Melody

func Run() {
	server = melody.New()
	ChatRoom.Server = server

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

// server.go ------------ ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

func remindToCheckOut(id string, count int) {
	// 計時30秒,30秒後傳訊息給旅客,提醒退房

	ChatRoom.DefaultRoomManager.UUIDMap[id].CheckOutTime = ChatRoom.DefaultRoomManager.UUIDMap[id].CheckInTime.Add(time.Duration(count) * time.Second)
	timer := time.NewTimer(time.Duration(count) * time.Second)

	<-timer.C
	timer.Stop() //時間到之後停止

	text := id + "退房時間到了"
	ChatRoom.SentMessageTo(id, nil, text, "user")
}

// 收到訊息後處理
func getMessage(s *melody.Session, msg []byte) {

	cmd := gjson.Get(string(msg), "content").String()
	// 如果開頭是"/",就代表是指令
	// 否則廣播房間內所有人訊息
	if string(cmd[0]) == "/" {
		Entry(s, string(msg))
	} else {
		// 取得名字
		id := s.Request.URL.Query().Get("id")
		ChatRoom.SentMessageTo(id, msg, "", "room")
	}
}

// 第一次連線處理
func firstConnect(session *melody.Session) {

	// 取得名字
	id := session.Request.URL.Query().Get("id")

	// 有人進入就登記他的id
	session.Set("user_id", id)

	// 預設值金錢
	money := 1000
	// 現在時間
	t := time.Now().Format("2006-01-02 15:04:05")

	// 當旅館滿人進入
	if len(ChatRoom.DefaultRoomManager.UUIDMap) >= 8 {
		fmt.Println("目前人數", len(ChatRoom.DefaultRoomManager.UUIDMap), "超過人數(最多8人),無法入住")

		text := "因為客滿被踢出"
		ChatRoom.SentMessageTo(id, nil, text, "user")

		time.Sleep(time.Millisecond * 300)
		session.Close()
		return
	}

	// 當有人連線時,將資料寫入旅客名單
	tt := ChatRoom.NewTourist(id, money, session)

	// 交給控制中心,新增使用者
	ChatRoom.DefaultRoomManager.SignNewMember(tt)
	// 加入房間的工作
	room := ChatRoom.DefaultRoomManager.Work(tt)
	// 以上成功之後顯示

	text := "加入聊天室,房間號碼：" + room + ",時間：" + t
	ChatRoom.SentMessageTo(id, nil, text, "room")

	// 提醒退房,第二個是秒數
	go remindToCheckOut(id, 30)

}

// 離線時處理
func sendLeaveRoom(session *melody.Session, i int, s string) error {

	id := session.Request.URL.Query().Get("id")

	if ChatRoom.DefaultRoomManager.FindUser(id) {
		text := "離開聊天室"
		ChatRoom.SentMessageTo(id, nil, text, "room")

	} else {
		// server.Broadcast(NewMessage("other", id, "因為客滿而離開聊天室").GetByteMessage())
	}

	ChatRoom.DefaultRoomManager.SignOutMember(id)
	return nil
}

// 縮減為一個ChatRoom.SentMessageTo

// func ChatRoom.SentMessageToChatRoom(id string, msg []byte) {
// 	// server.Broadcast(msg)
// 	room := "room_" + strconv.Itoa(ChatRoom.DefaultRoomManager.UUIDMap[id].room.roomID)
// 	server.BroadcastFilter(msg, func(session *melody.Session) bool {
// 		compareID, _ := session.Get("chat_id")
// 		return compareID == "chat_id" || compareID == room
// 	})
// }

// func sentOtherToChatRoom(id string, text string) {
// 	room := "room_" + strconv.Itoa(ChatRoom.DefaultRoomManager.UUIDMap[id].room.roomID)
// 	server.BroadcastFilter(NewMessage("other", id, text).GetByteMessage(), func(session *melody.Session) bool {
// 		compareID, _ := session.Get("chat_id")
// 		return compareID == "chat_id" || compareID == room
// 	})
// }

// func sentOtherReturn(id string, text string) {
// 	server.BroadcastFilter(NewMessage("other", id, text).GetByteMessage(), func(session *melody.Session) bool {
// 		compareID, _ := session.Get("user_id")
// 		return compareID == "user_id" || compareID == id
// 	})
// }

func Entry(s *melody.Session, msg string) {

	cmd := gjson.Get(string(msg), "content").String()
	cmd = string(cmd[1:len(string(cmd))]) // 去除前面的斜線
	fmt.Println(cmd)
	fn, ok := CmdMap[cmd]
	if !ok {
		id := s.Request.URL.Query().Get("id") // 名字
		ChatRoom.SentMessageTo(id, nil, "指令錯誤,可輸入/help查看指令", "user")

		return
	}

	fn(s, msg)
}

// 外部如何使用
// func SendBroadcastFilter(context []byte, KEY string, target string) {
// 	server.BroadcastFilter(context, func(session *melody.Session) bool {
// 		compareID, _ := session.Get(KEY)
// 		return compareID == "chat_id" || compareID == target
// 	})
// }

// server.go ------------ 。
