package main

import (
	"net/http"

	"pixeltactics.com/match/src/config"
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/utils/cloud"
	ws "pixeltactics.com/match/src/websocket/core"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	config.Setup()

	badger := databases.NewBadgerImpl()
	defer badger.Close()

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
			"region": cloud.GetServerRegion(),
		})
	})

	router.GET("/players", func(context *gin.Context) {
		context.JSON(http.StatusOK, map[string]interface{}{
			"players": clientHub.GetAllPlayerId(),
		})
	})

	router.GET("/ws", func(context *gin.Context) {
		ws.ServeWebSocket(clientHub, context.Writer, context.Request)
	})

	router.Run("0.0.0.0:8000")
}
