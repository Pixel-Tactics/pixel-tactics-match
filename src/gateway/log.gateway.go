package gateway

import (
	"log"

	"github.com/go-playground/validator/v10"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/messages"
	"pixeltactics.com/match/src/services"
	convert_utils "pixeltactics.com/match/src/utils/convert"
)

type LogGateway interface {
	SetMessager(messager messages.Messager)
}

type LogGatewayImpl struct {
	LogService   services.LogService
	EventManager events.EventManager
	BaseGateway
}

func (gateway *LogGatewayImpl) sendLogUpdates(event interface{}) error {
	log.Println("Got State Update")
	concreteEvent, ok := event.(*services.LogAddedEvent)
	if !ok {
		log.Println("invalid event on State Update Event")
		log.Fatalln(event)
	}
	response, err := convert_utils.ObjectToMap(concreteEvent)
	if err != nil {
		log.Println("invalid event on State Update Event")
		log.Fatalln(err)
	}
	for _, playerID := range concreteEvent.PlayerIDs {
		gateway.Messager.Send(playerID, &messages.Message{
			Type: TYPE_STATE_CHANGE,
			Body: response,
		})
	}
	return nil
}

func (gateway *LogGatewayImpl) SetMessager(messager messages.Messager) {
	gateway.Messager = messager
}

func NewLogGateway(
	logService services.LogService,
	eventManager events.EventManager,
	validator *validator.Validate,
) LogGateway {
	gateway := &LogGatewayImpl{
		LogService:   logService,
		EventManager: eventManager,
		BaseGateway: BaseGateway{
			Messager:  nil,
			Validator: validator,
		},
	}
	eventManager.On(services.LOG_ADDED_EVENT, gateway.sendLogUpdates)
	return gateway
}
