package http

import (
        "net/http"

        "esports-fantasy-backend/internal/services"

        "github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
        analyticsService *services.AnalyticsService
}

func NewAnalyticsHandler(analyticsService *services.AnalyticsService) *AnalyticsHandler {
        return &AnalyticsHandler{
                analyticsService: analyticsService,
        }
}

// GetDashboardStats gets comprehensive dashboard statistics
func (h *AnalyticsHandler) GetDashboardStats(c *gin.Context) {
        // Check if user is admin
        isAdmin := c.GetBool("is_admin")
        if !isAdmin {
                c.JSON(http.StatusForbidden, gin.H{
                        "success": false,
                        "message": "Admin access required",
                })
                return
        }

        stats, err := h.analyticsService.GetCachedStats()
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{
                        "success": false,
                        "message": "Failed to retrieve dashboard statistics",
                        "error":   err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Dashboard statistics retrieved successfully",
                "data":    stats,
        })
}

// GetUserAnalytics gets detailed user analytics
func (h *AnalyticsHandler) GetUserAnalytics(c *gin.Context) {
        userID := c.Param("userId")
        currentUserID := c.GetString("user_id")
        isAdmin := c.GetBool("is_admin")

        // Users can only view their own analytics unless they're admin
        if userID != currentUserID && !isAdmin {
                c.JSON(http.StatusForbidden, gin.H{
                        "success": false,
                        "message": "Access denied",
                })
                return
        }

        analytics, err := h.analyticsService.GetUserAnalytics(userID)
        if err != nil {
                c.JSON(http.StatusNotFound, gin.H{
                        "success": false,
                        "message": "User analytics not found",
                        "error":   err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "User analytics retrieved successfully",
                "data":    analytics,
        })
}

// GetMatchAnalytics gets detailed match analytics
func (h *AnalyticsHandler) GetMatchAnalytics(c *gin.Context) {
        matchID := c.Param("matchId")
        if matchID == "" {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Match ID is required",
                })
                return
        }

        analytics, err := h.analyticsService.GetMatchAnalytics(matchID)
        if err != nil {
                c.JSON(http.StatusNotFound, gin.H{
                        "success": false,
                        "message": "Match analytics not found",
                        "error":   err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Match analytics retrieved successfully",
                "data":    analytics,
        })
}

// RefreshAnalytics manually refreshes analytics cache
func (h *AnalyticsHandler) RefreshAnalytics(c *gin.Context) {
        // Check if user is admin
        isAdmin := c.GetBool("is_admin")
        if !isAdmin {
                c.JSON(http.StatusForbidden, gin.H{
                        "success": false,
                        "message": "Admin access required",
                })
                return
        }

        if err := h.analyticsService.CacheStats(); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{
                        "success": false,
                        "message": "Failed to refresh analytics",
                        "error":   err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Analytics refreshed successfully",
        })
}