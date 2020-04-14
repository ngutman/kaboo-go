package websocket

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ngutman/kaboo-server-go/models"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"

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
	client *client
	data   []byte
}

type client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID string
}

// Hub registers, un-registers and manages websocket lifecycle
type Hub struct {
	upgrader       websocket.Upgrader
	clients        map[*client]bool
	usersToClients map[string]*client
	incoming       chan ClientMessage
	register       chan *client
	unregister     chan *client
}

// NewHub create a new hub instance
func NewHub() *Hub {
	return &Hub{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:        make(map[*client]bool),
		usersToClients: make(map[string]*client),
		incoming:       make(chan ClientMessage),
		register:       make(chan *client),
		unregister:     make(chan *client),
	}
}

// Run the hub, listens for client connections and disconnections, listens for incoming messages
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.usersToClients[client.userID] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				log.Debugf("Client %v (%v) disconnected\n", client.userID, client.conn.RemoteAddr().String())
				delete(h.clients, client)
				delete(h.usersToClients, client.userID)
				close(client.send)
			}
		case clientMessage := <-h.incoming:
			log.Tracef("Incoming message from %v - %v", clientMessage.client.userID, clientMessage.data)
			// TODO: Handle incoming commands
		}
	}
}

// BroadcastMessageToUsers send a message over WS to the given list of user
func (h *Hub) BroadcastMessageToUsers(users []primitive.ObjectID, message interface{}) {
	rawJSON, err := json.Marshal(message)
	if err != nil {
		log.Errorf("Failed marshalling json, %v", message)
		return
	}

	for _, userID := range users {
		if h.usersToClients[userID.Hex()] != nil {
			h.usersToClients[userID.Hex()].send <- rawJSON
			log.Debugf("Sent message to %v", userID)
		}
	}
}

// HandleWSUpgradeRequest attempt to upgrade the given connection to websocket and register the user
func (h *Hub) HandleWSUpgradeRequest(w http.ResponseWriter, r *http.Request, user *models.User) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Error upgrading client connection, %v\n", err)
	}
	client := &client{
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

func (c *client) readPump() {
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

func (c *client) writePump() {
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
