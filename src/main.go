package main

import (
	"log"

	ws "pixeltactics.com/match/src/websocket"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error in loading .env file")
		return
	}

	clientHub := ws.NewClientHub()
	go clientHub.Run()

	router := gin.Default()

	router.GET("/ws", func(context *gin.Context) {
		ws.ServeWebSocket(clientHub, context.Writer, context.Request)
	})

	router.Run("localhost:8000")
}
