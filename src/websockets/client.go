package websockets

import (
	"pixeltactics.com/match/src/messages"

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
	Hub     *ClientHub
	Conn    *websocket.Conn
	Receive chan *messages.Message
}

func (client *Client) handleReceive() {
	defer func() {
		client.Hub.UnregisterClientQueue <- client
		client.Conn.Close()
	}()
	client.Conn.SetReadLimit(maxMessageSize)
	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error { client.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, jsonBytes, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		message, err := messages.JsonBytesToMessage(jsonBytes)
		if err != nil {
			continue
		}

		client.Hub.MessageQueue <- &MessageRequest{
			Message: message,
			Client:  client,
		}
	}
}

func (client *Client) handleSend() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.Receive:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			jsonBytes, err := messages.MessageToJsonBytes(message)
			if err != nil {
				return
			}
			w.Write(jsonBytes)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
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
		Hub:     hub,
		Conn:    conn,
		Receive: make(chan *messages.Message, 256),
	}

	client.Hub.RegisterClient <- client

	go client.handleSend()
	go client.handleReceive()
}
