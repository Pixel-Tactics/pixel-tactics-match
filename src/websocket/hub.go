package ws

import (
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
	playerIdToClient map[string]*Client
	clientToPlayerId map[*Client]string
}

func (hub *PlayerHub) RegisterPlayer(playerId string, client *Client) {
	hub.playerIdToClient[playerId] = client
	hub.clientToPlayerId[client] = playerId
}

func (hub *PlayerHub) UnregisterPlayer(client *Client) {
	playerId, ok := hub.clientToPlayerId[client]
	if ok {
		delete(hub.clientToPlayerId, client)
	}
	_, ok = hub.playerIdToClient[playerId]
	if ok {
		delete(hub.playerIdToClient, playerId)
	}
}

type ClientHub struct {
	playerHub      *PlayerHub
	clientList     map[*Client]bool
	registerClient chan *Client
	registerPlayer chan *PlayerRegistration
	unregister     chan *Client
	message        chan *MessageWithClient
}

func (hub *ClientHub) GetClientFromPlayerId(playerId string) (*Client, bool) {
	client, ok := hub.playerHub.playerIdToClient[playerId]
	if !ok {
		return nil, false
	}
	return client, true
}

func (hub *ClientHub) GetPlayerIdFromClient(client *Client) (string, bool) {
	playerId, ok := hub.playerHub.clientToPlayerId[client]
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
	_, ok := hub.clientList[client]
	if ok {
		delete(hub.clientList, client)
		close(client.receive)
	}
}

// TODO: Handle client reconnecting mid game -> player
func (hub *ClientHub) Run() {
	authHandler := handlers.NewAuthHandler()
	go authHandler.Run()

	sessionHandler := handlers.NewSessionHandler()
	go sessionHandler.Run()

	for {
		select {
		case client := <-hub.registerClient:
			hub.clientList[client] = true
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
					SendToClient: func(message *types.Message) {
						client.receive <- message
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
		playerIdToClient: make(map[string]*Client),
		clientToPlayerId: make(map[*Client]string),
	}
	return &ClientHub{
		playerHub:      playerHub,
		clientList:     make(map[*Client]bool),
		registerClient: make(chan *Client),
		registerPlayer: make(chan *PlayerRegistration),
		unregister:     make(chan *Client),
		message:        make(chan *MessageWithClient),
	}
}
