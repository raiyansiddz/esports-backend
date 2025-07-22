package ws

import (
	"esports-fantasy-backend/internal/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WebSocketHandler struct {
	hub                *Hub
	leaderboardService services.LeaderboardService
}

func NewWebSocketHandler(hub *Hub, leaderboardService services.LeaderboardService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:                hub,
		leaderboardService: leaderboardService,
	}
}

// HandleWebSocket godoc
// @Summary WebSocket endpoint
// @Description WebSocket connection for real-time updates
// @Tags websocket
// @Success 101 "Switching Protocols"
// @Router /ws [get]
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}

	// Create new client
	client := &Client{
		ID:   uuid.New(),
		Conn: conn,
		Send: make(chan []byte, 256),
		Hub:  h.hub,
	}

	// Register client
	h.hub.Register <- client

	// Start goroutines for handling reads and writes
	go client.WritePump()
	go client.ReadPump()

	log.Printf("🌐 WebSocket client connected: %s", client.ID)
}

// BroadcastLeaderboardUpdate sends leaderboard updates to all subscribed clients
func (h *WebSocketHandler) BroadcastLeaderboardUpdate(contestID uuid.UUID) {
	leaderboard, err := h.leaderboardService.GetLeaderboard(contestID, 100)
	if err != nil {
		log.Printf("Error getting leaderboard for broadcast: %v", err)
		return
	}

	message := WebSocketMessage{
		Type: "leaderboard_update",
		Payload: map[string]interface{}{
			"contest_id": contestID,
			"rankings":   leaderboard,
		},
	}

	roomName := "contest:" + contestID.String()
	h.hub.BroadcastToRoom(roomName, message)
}

// BroadcastMatchStatusUpdate sends match status updates
func (h *WebSocketHandler) BroadcastMatchStatusUpdate(matchID uuid.UUID, status string) {
	message := WebSocketMessage{
		Type: "match_status_update",
		Payload: map[string]interface{}{
			"match_id": matchID,
			"status":   status,
		},
	}

	roomName := "match:" + matchID.String()
	h.hub.BroadcastToRoom(roomName, message)
}