package router

import (
	"encoding/json"

	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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

func Run() {
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
