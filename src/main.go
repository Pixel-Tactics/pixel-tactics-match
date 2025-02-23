package main

import (
	"net/http"

	"pixeltactics.com/match/src/config"
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/gateway"
	"pixeltactics.com/match/src/services"
	"pixeltactics.com/match/src/utils/cloud"
	"pixeltactics.com/match/src/websockets"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	config.Setup()

	badger := databases.NewBadgerImpl()
	defer badger.Close()

	validator := validator.New()
	authService := services.NewAuthService()
	authGateway := gateway.NewAuthGateway(authService, validator)
	gatewayRouter := gateway.NewRouter(authGateway)
	clientHub := websockets.NewClientHub(gatewayRouter)
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
			"players": clientHub.GetAllUserId(),
		})
	})

	router.GET("/ws", func(context *gin.Context) {
		websockets.ServeWebSocket(clientHub, context.Writer, context.Request)
	})

	router.Run("0.0.0.0:8000")
}
