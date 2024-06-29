package ws

import (
	"pixeltactics.com/match/src/data_structures"
	"pixeltactics.com/match/src/handlers"
	"pixeltactics.com/match/src/types"
)

type MessageWithClient struct {
	Message *types.Message
	Client  *Client
}

type PlayerRegistration struct {
	playerId string
	client   *Client
}

type PlayerHub struct {
	playerIdToClient *data_structures.SyncMap[string, *Client]
	clientToPlayerId *data_structures.SyncMap[*Client, string]
}

func (hub *PlayerHub) RegisterPlayer(playerId string, client *Client) {
	hub.playerIdToClient.Store(playerId, client)
	hub.clientToPlayerId.Store(client, playerId)
}

func (hub *PlayerHub) UnregisterPlayer(client *Client) {
	playerId, ok := hub.clientToPlayerId.Load(client)
	if ok {
		hub.clientToPlayerId.Delete(client)
	}
	_, ok = hub.playerIdToClient.Load(playerId)
	if ok {
		hub.playerIdToClient.Delete(playerId)
	}
}

type ClientHub struct {
	playerHub      *PlayerHub
	clientList     *data_structures.SyncMap[*Client, bool]
	registerClient chan *Client
	registerPlayer chan *PlayerRegistration
	unregister     chan *Client
	message        chan *MessageWithClient
}

func (hub *ClientHub) GetClientFromPlayerId(playerId string) (*Client, bool) {
	client, ok := hub.playerHub.playerIdToClient.Load(playerId)
	if !ok {
		return nil, false
	}
	return client, true
}

func (hub *ClientHub) GetPlayerIdFromClient(client *Client) (string, bool) {
	playerId, ok := hub.playerHub.clientToPlayerId.Load(client)
	if !ok {
		return "", false
	}
	return playerId, true
}

func (hub *ClientHub) RegisterPlayer(playerId string, client *Client) {
	hub.registerPlayer <- &PlayerRegistration{
		playerId: playerId,
		client:   client,
	}
}

func (hub *ClientHub) UnregisterClient(client *Client) {
	hub.playerHub.UnregisterPlayer(client)
	_, ok := hub.clientList.Load(client)
	if ok {
		hub.clientList.Delete(client)
		close(client.receive)
	}
}

func (hub *ClientHub) Run() {
	authHandler := handlers.NewAuthHandler()
	go authHandler.Run()

	sessionHandler := handlers.NewSessionHandler()
	go sessionHandler.Run()

	for {
		select {
		case client := <-hub.registerClient:
			hub.clientList.Store(client, true)
		case register := <-hub.registerPlayer:
			hub.playerHub.RegisterPlayer(register.playerId, register.client)
		case client := <-hub.unregister:
			hub.UnregisterClient(client)
		case pair := <-hub.message:
			message := pair.Message
			client := pair.Client

			interaction := &types.Interaction{
				Request: &types.Request{
					Message: message,
				},
				Response: &types.Response{
					SendToClient: func(inMessage *types.Message) {
						inMessage.Identifier = message.Identifier
						client.receive <- inMessage
					},
					NotifyClient: func(inMessage *types.Message) {
						inMessage.Identifier = "notification"
						client.receive <- inMessage
					},
					NotifyOtherClient: func(playerId string, inMessage *types.Message) {
						inMessage.Identifier = "notification"
						otherClient, ok := hub.GetClientFromPlayerId(playerId)
						if ok {
							otherClient.receive <- inMessage
						}
					},
					RegisterPlayer: func(playerId string) {
						hub.RegisterPlayer(playerId, client)
					},
				},
			}

			if message.Action == types.ACTION_AUTH {
				authHandler.Interaction <- interaction
			} else {
				playerId, hasPlayer := hub.GetPlayerIdFromClient(client)
				if !hasPlayer {
					continue
				}
				message.Body["playerId"] = playerId
				sessionHandler.Interaction <- interaction
			}
		}
	}
}

func NewClientHub() *ClientHub {
	playerHub := &PlayerHub{
		playerIdToClient: data_structures.NewSyncMap[string, *Client](),
		clientToPlayerId: data_structures.NewSyncMap[*Client, string](),
	}
	return &ClientHub{
		playerHub:      playerHub,
		clientList:     data_structures.NewSyncMap[*Client, bool](),
		registerClient: make(chan *Client, 256),
		registerPlayer: make(chan *PlayerRegistration, 256),
		unregister:     make(chan *Client, 256),
		message:        make(chan *MessageWithClient, 256),
	}
}
