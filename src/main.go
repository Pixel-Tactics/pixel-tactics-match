package main

import (
	"net/http"

	"pixeltactics.com/match/src/utils"
	ws "pixeltactics.com/match/src/websocket/core"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	clientHub := ws.NewClientHub()
	go clientHub.Run()

	router := gin.Default()

	router.GET("/", func(context *gin.Context) {
		context.JSON(http.StatusOK, map[string]string{
			"message": "match service",
		})
	})

	router.GET("/region", func(context *gin.Context) {
		context.JSON(http.StatusOK, map[string]string{
			"region": utils.GetServerRegion(),
		})
	})

	router.GET("/ws", func(context *gin.Context) {
		ws.ServeWebSocket(clientHub, context.Writer, context.Request)
	})

	router.Run("0.0.0.0:8000")
}
