package transport

import (
	"bytes"
	"net/http"
	"time"

	"github.com/ngutman/kaboo-server-go/models"
	log "github.com/sirupsen/logrus"

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

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// ClientMessage a message received from a client
type ClientMessage struct {
	client *WebsocketClient
	data   []byte
}

// WebsocketClient manages a single websocket clients, handling reading and writing
type WebsocketClient struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID string
}

// Hub registers, un-registers and manages websocket lifecycle
type Hub struct {
	upgrader   websocket.Upgrader
	clients    map[*WebsocketClient]bool
	incoming   chan ClientMessage
	register   chan *WebsocketClient
	unregister chan *WebsocketClient
}

func newHub() *Hub {
	return &Hub{
		upgrader:   websocket.Upgrader{},
		clients:    make(map[*WebsocketClient]bool),
		incoming:   make(chan ClientMessage),
		register:   make(chan *WebsocketClient),
		unregister: make(chan *WebsocketClient),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				log.Debugf("Client %v (%v) disconnected\n", client.userID, client.conn.RemoteAddr().String())
				delete(h.clients, client)
				close(client.send)
			}
		case clientMessage := <-h.incoming:
			log.Tracef("Incoming message from %v - %v", clientMessage.client.userID, clientMessage.data)
			// TODO: Handle incoming commands
		}
	}
}

func (h *Hub) handleWSUpgradeRequest(w http.ResponseWriter, r *http.Request, user *models.User) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Error upgrading client connection, %v\n", err)
	}
	client := &WebsocketClient{
		hub:    h,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: user.ID.Hex(),
	}
	log.Debugf("Client %v (%v) connected\n", user, r.RemoteAddr)
	h.register <- client

	go client.readPump()
	go client.writePump()
}

func (c *WebsocketClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// TODO: Validate that we only trim the newline
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.incoming <- ClientMessage{c, message}
	}
}

func (c *WebsocketClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
