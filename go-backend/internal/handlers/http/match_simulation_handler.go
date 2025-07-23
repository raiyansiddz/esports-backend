package http

import (
        "net/http"

        "esports-fantasy-backend/internal/services"

        "github.com/gin-gonic/gin"
        "github.com/google/uuid"
        "github.com/gorilla/websocket"
)

type MatchSimulationHandler struct {
        matchSimulationService *services.MatchSimulationService
        upgrader               websocket.Upgrader
}

func NewMatchSimulationHandler(matchSimulationService *services.MatchSimulationService) *MatchSimulationHandler {
        return &MatchSimulationHandler{
                matchSimulationService: matchSimulationService,
                upgrader: websocket.Upgrader{
                        CheckOrigin: func(r *http.Request) bool {
                                return true // Allow all origins in development
                        },
                },
        }
}

// StartSimulation starts live match simulation (admin only)
func (h *MatchSimulationHandler) StartSimulation(c *gin.Context) {
        // Check if user is admin
        isAdmin := c.GetBool("is_admin")
        if !isAdmin {
                c.JSON(http.StatusForbidden, gin.H{
                        "success": false,
                        "message": "Admin access required",
                })
                return
        }

        matchID := c.Param("matchId")
        if matchID == "" {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Match ID is required",
                })
                return
        }

        if err := h.matchSimulationService.StartMatchSimulation(matchID); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Failed to start match simulation",
                        "error":   err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Match simulation started successfully",
                "data": gin.H{
                        "match_id": matchID,
                        "status":   "SIMULATION_STARTED",
                },
        })
}

// StopSimulation stops live match simulation (admin only)
func (h *MatchSimulationHandler) StopSimulation(c *gin.Context) {
        // Check if user is admin
        isAdmin := c.GetBool("is_admin")
        if !isAdmin {
                c.JSON(http.StatusForbidden, gin.H{
                        "success": false,
                        "message": "Admin access required",
                })
                return
        }

        matchID := c.Param("matchId")
        if matchID == "" {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Match ID is required",
                })
                return
        }

        if err := h.matchSimulationService.StopMatchSimulation(matchID); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Failed to stop match simulation",
                        "error":   err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Match simulation stopped successfully",
                "data": gin.H{
                        "match_id": matchID,
                        "status":   "SIMULATION_STOPPED",
                },
        })
}

// GetActiveSimulations gets all active match simulations (admin only)
func (h *MatchSimulationHandler) GetActiveSimulations(c *gin.Context) {
        // Check if user is admin
        isAdmin := c.GetBool("is_admin")
        if !isAdmin {
                c.JSON(http.StatusForbidden, gin.H{
                        "success": false,
                        "message": "Admin access required",
                })
                return
        }

        simulations := h.matchSimulationService.GetActiveSimulations()
        
        // Convert to response format
        activeList := make([]gin.H, 0)
        for matchID, sim := range simulations {
                activeList = append(activeList, gin.H{
                        "match_id":     matchID,
                        "match_name":   sim.Match.Name,
                        "start_time":   sim.StartTime,
                        "clients":      len(sim.Clients),
                        "events":       len(sim.Events),
                        "is_active":    sim.IsActive,
                })
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Active simulations retrieved successfully",
                "data": gin.H{
                        "active_simulations": activeList,
                        "total_count":        len(activeList),
                },
        })
}

// GetMatchEvents gets match events for a specific match
func (h *MatchSimulationHandler) GetMatchEvents(c *gin.Context) {
        matchID := c.Param("matchId")
        if matchID == "" {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Match ID is required",
                })
                return
        }

        events, err := h.matchSimulationService.GetMatchEvents(matchID)
        if err != nil {
                c.JSON(http.StatusNotFound, gin.H{
                        "success": false,
                        "message": "Match events not found",
                        "error":   err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Match events retrieved successfully",
                "data": gin.H{
                        "match_id": matchID,
                        "events":   events,
                        "count":    len(events),
                },
        })
}

// WebSocketHandler handles WebSocket connections for live match updates
func (h *MatchSimulationHandler) WebSocketHandler(c *gin.Context) {
        matchID := c.Param("matchId")
        if matchID == "" {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Match ID is required",
                })
                return
        }

        // Upgrade connection to WebSocket
        conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
        if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Failed to upgrade connection",
                        "error":   err.Error(),
                })
                return
        }

        // Generate client ID
        clientID := uuid.New().String()

        // Add client to simulation
        if err := h.matchSimulationService.AddWebSocketClient(matchID, clientID, conn); err != nil {
                conn.Close()
                return
        }

        // Handle client disconnection
        defer func() {
                h.matchSimulationService.RemoveWebSocketClient(matchID, clientID)
        }()

        // Keep connection alive and handle client messages
        for {
                _, _, err := conn.ReadMessage()
                if err != nil {
                        break
                }
                // Echo back or handle client messages if needed
        }
}