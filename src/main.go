package main

import (
	"net/http"

	"pixeltactics.com/match/src/config"
	"pixeltactics.com/match/src/core/states"
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/gateway"
	"pixeltactics.com/match/src/heroes"
	"pixeltactics.com/match/src/integrations/communication"
	"pixeltactics.com/match/src/repositories"
	"pixeltactics.com/match/src/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	config.Setup()

	rmqManager := communication.NewRMQManager()

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

	mapService := services.NewMapService(mapRepo, nil)
	heroService := services.NewHeroService(heroRepo, nil, nil, heroFactory, badger)
	playerService := services.NewPlayerService(playerRepo, nil)
	logService := services.NewLogService(logRepo, eventManager)
	sessionService := services.NewSessionService(mapService, heroService, playerService, sessionRepo, stateFactory, badger, logService, eventManager)
	actionService := services.NewActionService(mapService, heroService, sessionService, logService, badger)
	inviteService := services.NewInvitationService(inviteRepo, badger, sessionService, eventManager)

	mapService.SetHeroService(heroService)
	playerService.SetSessionService(sessionService)
	heroService.SetSessionService(sessionService)
	heroService.SetPlayerService(playerService)

	sessionGateway := gateway.NewSessionGateway(sessionService, eventManager, validator)
	actionGateway := gateway.NewActionGateway(actionService, validator)
	logGateway := gateway.NewLogGateway(logService, eventManager, validator)
	inviteGateway := gateway.NewInvitationGateway(inviteService, validator, eventManager)
	gatewayRouter := gateway.NewRouter(sessionGateway, actionGateway)

	incomingQueue := communication.NewIncomingQueue(gatewayRouter, rmqManager, eventManager)
	outgoingQueue := communication.NewOutgoingQueue(rmqManager, eventManager)
	incomingStream := communication.NewIncomingStream(rmqManager, eventManager)

	logGateway.SetMessager(outgoingQueue)
	sessionGateway.SetMessager(outgoingQueue)
	inviteGateway.SetMessager(outgoingQueue)

	go incomingQueue.Run()
	go outgoingQueue.Run()
	go incomingStream.Run("matchmaking/invite", gateway.INVITE_REQUEST_EVENT)

	router := gin.Default()

	router.GET("/", func(context *gin.Context) {
		context.JSON(http.StatusOK, map[string]string{
			"message": "match service",
		})
	})

	router.Run("0.0.0.0:8000")
}
