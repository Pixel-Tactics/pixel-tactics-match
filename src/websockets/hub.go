package websockets

import (
	"pixeltactics.com/match/src/gateway"
	"pixeltactics.com/match/src/messages"
	"pixeltactics.com/match/src/utils/datastructures"
)

type MessageRequest struct {
	Message *messages.Message
	Client  *Client
}

type SetUserIdRequest struct {
	UserId string
	Client *Client
}

type ClientHub struct {
	Router gateway.Router

	UserIdToClient        *datastructures.SyncMap[string, *Client]
	ClientToUserId        *datastructures.SyncMap[*Client, string]
	ClientList            *datastructures.SyncMap[*Client, bool]
	RegisterClient        chan *Client
	SetUserIdQueue        chan *SetUserIdRequest
	UnregisterClientQueue chan *Client
	MessageQueue          chan *MessageRequest
}

func (hub *ClientHub) GetAllUserId() []string {
	return hub.UserIdToClient.Keys()
}

func (hub *ClientHub) GetClientFromUserId(userId string) (*Client, bool) {
	client, ok := hub.UserIdToClient.Load(userId)
	if !ok {
		return nil, false
	}
	return client, true
}

func (hub *ClientHub) GetUserIdFromClient(client *Client) (string, bool) {
	userId, ok := hub.ClientToUserId.Load(client)
	if !ok {
		return "", false
	}
	return userId, true
}

func (hub *ClientHub) QueueSetUserId(userId string, client *Client) {
	hub.SetUserIdQueue <- &SetUserIdRequest{
		UserId: userId,
		Client: client,
	}
}

func (hub *ClientHub) setUserId(userId string, client *Client) {
	hub.UserIdToClient.Store(userId, client)
	hub.ClientToUserId.Store(client, userId)
}

func (hub *ClientHub) unregisterClient(client *Client) {
	userId, ok := hub.ClientToUserId.Load(client)
	if ok {
		hub.ClientToUserId.Delete(client)
	}
	_, ok = hub.UserIdToClient.Load(userId)
	if ok {
		hub.UserIdToClient.Delete(userId)
	}
	_, ok = hub.ClientList.Load(client)
	if ok {
		hub.ClientList.Delete(client)
		close(client.Receive)
	}
}

func (hub *ClientHub) SendMessageToUser(userId string, msg *messages.Message) {
	msg.Identifier = "notification"
	otherClient, ok := hub.GetClientFromUserId(userId)
	if ok {
		otherClient.Receive <- msg
	}
}

func (hub *ClientHub) Run() {
	for {
		select {
		case client := <-hub.RegisterClient:
			hub.ClientList.Store(client, true)
		case request := <-hub.SetUserIdQueue:
			hub.setUserId(request.UserId, request.Client)
		case client := <-hub.UnregisterClientQueue:
			hub.unregisterClient(client)
		case request := <-hub.MessageQueue:
			message := request.Message
			client := request.Client

			var clientId *string
			userId, ok := hub.GetUserIdFromClient(client)
			if ok {
				clientId = &userId
			}

			go hub.Router.RouteMessage(&messages.WebSocketMessager{
				Message:  message,
				ClientId: clientId,
				Send: func(userId string, message *messages.Message) {
					message.Identifier = "notification"
					otherClient, ok := hub.GetClientFromUserId(userId)
					if ok {
						otherClient.Receive <- message
					}
				},
				SendBack: func(inMessage *messages.Message) {
					inMessage.Identifier = message.Identifier
					client.Receive <- inMessage
				},
				SetClientId: func(userId string) {
					hub.QueueSetUserId(userId, client)
				},
			})
		}
	}
}

func NewClientHub(
	router gateway.Router,
) *ClientHub {
	return &ClientHub{
		Router:                router,
		UserIdToClient:        datastructures.NewSyncMap[string, *Client](),
		ClientToUserId:        datastructures.NewSyncMap[*Client, string](),
		ClientList:            datastructures.NewSyncMap[*Client, bool](),
		RegisterClient:        make(chan *Client, 256),
		SetUserIdQueue:        make(chan *SetUserIdRequest, 256),
		UnregisterClientQueue: make(chan *Client, 256),
		MessageQueue:          make(chan *MessageRequest, 256),
	}
}
