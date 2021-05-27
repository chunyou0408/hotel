package main

import (
	ChatRoom "hotel/chatRoom"
	Router "hotel/router"
)

func main() {
	ChatRoom.InitManager() //初始化控制中心

	Router.Run() // 啟動伺服器
}
