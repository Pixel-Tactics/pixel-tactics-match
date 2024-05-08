package main

import (
	ws "pixeltactics.com/match/src/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	clientHub := ws.NewClientHub()
	go clientHub.Run()

	router := gin.Default()

	router.GET("/ws", func(context *gin.Context) {
		ws.ServeWebSocket(clientHub, context.Writer, context.Request)
	})

	router.Run("127.0.0.1:8080")
}
