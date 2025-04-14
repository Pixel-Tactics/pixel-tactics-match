package main

import (
	"net/http"

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

	badger := databases.NewBadgerImpl()
	defer badger.Close()

	validator := validator.New()

	heroFactory := heroes.NewBaseHeroFactory()
	eventManager := events.NewSequentialEventManager()
	stateFactory := states.NewSessionStateFactory()

	mapRepo := repositories.NewMapRepository(badger)
	heroRepo := repositories.NewHeroRepository(badger)
	playerRepo := repositories.NewPlayerRepository(badger)
	sessionRepo := repositories.NewSessionRepositoryV2(badger)
	logRepo := repositories.NewSessionLogRepository(badger)
	inviteRepo := repositories.NewInvitationRepository()

	authService := services.NewAuthService()
	mapService := services.NewMapService(mapRepo, nil)
	heroService := services.NewHeroService(heroRepo, nil, nil, heroFactory, badger)
	playerService := services.NewPlayerService(playerRepo, nil)
	logService := services.NewLogService(logRepo, eventManager)
	sessionService := services.NewSessionService(mapService, heroService, playerService, sessionRepo, stateFactory, badger, logService, eventManager)
	actionService := services.NewActionService(mapService, heroService, sessionService, logService, badger)
	inviteService := services.NewInvitationService(inviteRepo, badger, sessionService)

	mapService.SetHeroService(heroService)
	playerService.SetSessionService(sessionService)
	heroService.SetSessionService(sessionService)
	heroService.SetPlayerService(playerService)

	authGateway := gateway.NewAuthGateway(authService, validator)
	sessionGateway := gateway.NewSessionGateway(sessionService, eventManager, validator)
	actionGateway := gateway.NewActionGateway(actionService, validator)
	logGateway := gateway.NewLogGateway(logService, eventManager, validator)
	inviteGateway := gateway.NewInvitationGateway(inviteService, validator)
	gatewayRouter := gateway.NewRouter(authGateway, sessionGateway, actionGateway, inviteGateway)
	clientHub := websockets.NewClientHub(gatewayRouter)

	logGateway.SetMessager(clientHub)
	sessionGateway.SetMessager(clientHub)

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
