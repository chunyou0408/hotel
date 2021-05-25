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
	result, _ := json.Marshal(m)
	return result
}

func main() {
	InitManager() //初始化控制中心

	server := melody.New()

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

	//收到訊息
	server.HandleMessage(func(s *melody.Session, msg []byte) {
		// 取得名字
		id := s.Request.URL.Query().Get("id")
		// 顯示內容
		cmd := gjson.Get(string(msg), "content").String()
		fmt.Println(cmd)

		//如果第一個是斜線
		if string(cmd[0]) == "/" {
			// 執行指令
			if string(cmd[1:len(string(cmd))]) == "info" {
				id := s.Request.URL.Query().Get("id")
				tt := DefaultRoomManager.UUIDMap[id]
				mm := strconv.Itoa(tt.money)
				rr := strconv.Itoa(tt.room.roomID)
				time:=tt.checkInTime.Format("2006-01-02 15:04:05")
				
				// server.Broadcast(NewMessage("other", id, "名字:"+id+",金錢:"+mm+"房間:"+rr).GetByteMessage())
				const KEY = "user_id"
				server.BroadcastFilter(NewMessage("other", id, "名字:"+id+", 金錢:"+mm+", 房間:"+rr+", 入住時間:"+time).GetByteMessage(), func(session *melody.Session) bool {
					compareID, _ := session.Get(KEY)
					return compareID == "user_id" || compareID == id
				})
			} else {
				const KEY = "user_id"
				server.BroadcastFilter(NewMessage("other", id, "指令錯誤").GetByteMessage(), func(session *melody.Session) bool {
					compareID, _ := session.Get(KEY)
					return compareID == "user_id" || compareID == id
				})
			}
		} else {
			// server.Broadcast(msg)
			const KEY = "chat_id"
			// 查傳資料的房客在哪間房間
			r := DefaultRoomManager.findUserRoom(id)
			aa := strconv.Itoa(r.roomID)
			id := "room10" + aa
			server.BroadcastFilter(msg, func(session *melody.Session) bool {
				compareID, _ := session.Get(KEY)
				return compareID == "chat_id" || compareID == id
			})
		}
	})

	// 連線開始
	server.HandleConnect(func(session *melody.Session) {

		// 取得名字
		id := session.Request.URL.Query().Get("id")

		const KEY = "byebye"
		session.Set(KEY, id)

		money := 1000
		t := time.Now().Format("2006-01-02 15:04:05")

		if len(DefaultRoomManager.UUIDMap) >= 8 {
			fmt.Println("目前人數", len(DefaultRoomManager.UUIDMap), "超過人數(最多8人),無法入住")
			server.BroadcastFilter(NewMessage("other", id, "因為客滿被踢出").GetByteMessage(), func(session *melody.Session) bool {
				compareID, _ := session.Get(KEY)
				return compareID == "byebye" || compareID == id
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

	})

	// 連線結束
	// melody 有提供 HandleClose 方法讓我們處理離線的 session，
	// 我們這邊設定離線就發送一個 xxx 離開聊天室 的訊息給全部人
	server.HandleClose(func(session *melody.Session, i int, s string) error {

		id := session.Request.URL.Query().Get("id")
		DefaultRoomManager.SignOutMember(id)
		if !DefaultRoomManager.findUser(id) {
			// server.Broadcast(NewMessage("other", id, "因為客滿而離開聊天室").GetByteMessage())
		} else {
			server.Broadcast(NewMessage("other", id, "離開聊天室").GetByteMessage())
		}
		return nil
	})

	// 監聽
	r.Run(":5000")
}
