package ws

import (
	ws_types "pixeltactics.com/match/src/websocket/types"

	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{}

type Client struct {
	hub     *ClientHub
	conn    *websocket.Conn
	receive chan *ws_types.Message
}

func (client *Client) handleReceive() {
	defer func() {
		client.hub.unregister <- client
		client.conn.Close()
	}()
	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, jsonBytes, err := client.conn.ReadMessage()
		if err != nil {
			break
		}

		message, err := ws_types.JsonBytesToMessage(jsonBytes)
		if err != nil {
			continue
		}

		client.hub.message <- &MessageWithClient{
			Message: message,
			Client:  client,
		}
	}
}

func (client *Client) handleSend() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.receive:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			jsonBytes, err := ws_types.MessageToJsonBytes(message)
			if err != nil {
				return
			}
			w.Write(jsonBytes)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWebSocket(hub *ClientHub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		hub:     hub,
		conn:    conn,
		receive: make(chan *ws_types.Message, 256),
	}

	client.hub.registerClient <- client

	go client.handleSend()
	go client.handleReceive()
}
