package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		return true
	},
}

// Client represents a WebSocket client
type Client struct {
	ID     uuid.UUID
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
	Rooms  map[string]bool // Subscribed rooms/channels
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	Clients map[*Client]bool

	// Register requests from the clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Broadcast to all clients in a room
	Broadcast chan *BroadcastMessage

	// Room subscriptions
	Rooms map[string]map[*Client]bool
}

// BroadcastMessage contains the message and target room
type BroadcastMessage struct {
	Room    string
	Message []byte
}

// WebSocketMessage represents the structure of messages sent over WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Channel string      `json:"channel,omitempty"`
	Action  string      `json:"action,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *BroadcastMessage),
		Rooms:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			client.Rooms = make(map[string]bool)
			log.Printf("🔌 Client connected: %s", client.ID)

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				// Remove client from all rooms
				for room := range client.Rooms {
					if clients, ok := h.Rooms[room]; ok {
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.Rooms, room)
						}
					}
				}
				delete(h.Clients, client)
				close(client.Send)
				log.Printf("🔌 Client disconnected: %s", client.ID)
			}

		case broadcast := <-h.Broadcast:
			if clients, ok := h.Rooms[broadcast.Room]; ok {
				for client := range clients {
					select {
					case client.Send <- broadcast.Message:
					default:
						// Client's send channel is full, remove client
						delete(h.Clients, client)
						delete(clients, client)
						close(client.Send)
					}
				}
			}
		}
	}
}

func (h *Hub) BroadcastToRoom(room string, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling broadcast message: %v", err)
		return
	}

	h.Broadcast <- &BroadcastMessage{
		Room:    room,
		Message: data,
	}
}

func (h *Hub) SubscribeClientToRoom(client *Client, room string) {
	if _, ok := h.Rooms[room]; !ok {
		h.Rooms[room] = make(map[*Client]bool)
	}
	h.Rooms[room][client] = true
	client.Rooms[room] = true
	log.Printf("📺 Client %s subscribed to room: %s", client.ID, room)
}

func (h *Hub) UnsubscribeClientFromRoom(client *Client, room string) {
	if clients, ok := h.Rooms[room]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.Rooms, room)
		}
	}
	delete(client.Rooms, room)
	log.Printf("📺 Client %s unsubscribed from room: %s", client.ID, room)
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, messageData, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming message
		var msg WebSocketMessage
		if err := json.Unmarshal(messageData, &msg); err != nil {
			log.Printf("Error parsing WebSocket message: %v", err)
			continue
		}

		// Handle different message types
		switch msg.Action {
		case "subscribe":
			if msg.Channel != "" {
				c.Hub.SubscribeClientToRoom(c, msg.Channel)
			}
		case "unsubscribe":
			if msg.Channel != "" {
				c.Hub.UnsubscribeClientFromRoom(c, msg.Channel)
			}
		}
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}