package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func main() {
	router := gin.Default()

	router.GET("/ws", func(context *gin.Context) {
		conn, err := upgrader.Upgrade(context.Writer, context.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		// For Logic Here
	})

	router.Run("127.0.0.1:8080")
}
