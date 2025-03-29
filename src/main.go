package main

import (
	"net/http"
	"os"

	"pixeltactics.com/match/src/config"
	"pixeltactics.com/match/src/core/states"
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/gateway"
	"pixeltactics.com/match/src/heroes"
	"pixeltactics.com/match/src/repositories"
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

	os.RemoveAll("./temp/badger") // Debug

	badger := databases.NewBadgerImpl()
	defer badger.Close()

	validator := validator.New()

	heroFactory := heroes.NewBaseHeroFactory()
	stateFactory := states.NewSessionStateFactory()
	eventManager := events.NewSequentialEventManager()

	mapRepo := repositories.NewMapRepository(badger)
	heroRepo := repositories.NewHeroRepository(badger)
	playerRepo := repositories.NewPlayerRepository(badger)
	sessionRepo := repositories.NewSessionRepositoryV2(badger)

	authService := services.NewAuthService()
	mapService := services.NewMapService(mapRepo, nil)
	heroService := services.NewHeroService(heroRepo, nil, nil, heroFactory, badger, eventManager)
	playerService := services.NewPlayerService(playerRepo, nil)
	sessionService := services.NewSessionService(mapService, heroService, playerService, sessionRepo, stateFactory, badger)

	mapService.SetHeroService(heroService)
	playerService.SetSessionService(sessionService)
	heroService.SetSessionService(sessionService)
	heroService.SetPlayerService(playerService)

	authGateway := gateway.NewAuthGateway(authService, validator)
	sessionGateway := gateway.NewSessionGateway(sessionService, validator)
	gatewayRouter := gateway.NewRouter(authGateway, sessionGateway)
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
