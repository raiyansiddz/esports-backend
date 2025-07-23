package http

import (
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserAdvancedHandler struct {
	achievementService     services.AchievementService
	referralService        services.ReferralService
	seasonLeagueService    services.SeasonLeagueService
	playerAnalyticsService services.PlayerAnalyticsService
}

func NewUserAdvancedHandler(
	achievementService services.AchievementService,
	referralService services.ReferralService,
	seasonLeagueService services.SeasonLeagueService,
	playerAnalyticsService services.PlayerAnalyticsService,
) *UserAdvancedHandler {
	return &UserAdvancedHandler{
		achievementService:     achievementService,
		referralService:        referralService,
		seasonLeagueService:    seasonLeagueService,
		playerAnalyticsService: playerAnalyticsService,
	}
}

// === USER ACHIEVEMENTS ===

// GetMyAchievements godoc
// @Summary Get user achievements
// @Description Get authenticated user's achievements
// @Tags user-advanced
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.UserAchievement
// @Router /user/achievements [get]
func (h *UserAdvancedHandler) GetMyAchievements(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	userModel := user.(*models.User)

	achievements, err := h.achievementService.GetUserAchievements(userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get achievements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"achievements": achievements,
		"count":        len(achievements),
	})
}

// GetAvailableAchievements godoc
// @Summary Get available achievements
// @Description Get all available achievements for users
// @Tags user-advanced
// @Produce json
// @Success 200 {array} models.Achievement
// @Router /user/achievements/available [get]
func (h *UserAdvancedHandler) GetAvailableAchievements(c *gin.Context) {
	achievements, err := h.achievementService.GetActiveAchievements()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get achievements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"achievements": achievements,
		"count":        len(achievements),
	})
}

// === REFERRAL SYSTEM ===

// GenerateReferralCode godoc
// @Summary Generate referral code
// @Description Generate referral code for user
// @Tags user-advanced
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Router /user/referral/generate [post]
func (h *UserAdvancedHandler) GenerateReferralCode(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	userModel := user.(*models.User)

	referralCode, err := h.referralService.GenerateReferralCode(userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate referral code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"referral_code": referralCode,
		"message":       "Referral code generated successfully",
	})
}

// ApplyReferralCode godoc
// @Summary Apply referral code
// @Description Apply a referral code to get bonus
// @Tags user-advanced
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ReferralRequest true "Referral code"
// @Success 200 {object} map[string]string
// @Router /user/referral/apply [post]
func (h *UserAdvancedHandler) ApplyReferralCode(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	userModel := user.(*models.User)

	var req models.ReferralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if err := h.referralService.ApplyReferralCode(userModel.ID, req.ReferralCode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Referral code applied successfully"})
}

// GetReferralStats godoc
// @Summary Get referral statistics
// @Description Get user's referral statistics
// @Tags user-advanced
// @Produce json
// @Security BearerAuth
// @Success 200 {object} services.ReferralStats
// @Router /user/referral/stats [get]
func (h *UserAdvancedHandler) GetReferralStats(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	userModel := user.(*models.User)

	stats, err := h.referralService.GetReferralStats(userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get referral stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetReferralLeaderboard godoc
// @Summary Get referral leaderboard
// @Description Get top referrers leaderboard
// @Tags user-advanced
// @Produce json
// @Param limit query int false "Limit results (default: 10)"
// @Success 200 {array} services.ReferralEntry
// @Router /user/referral/leaderboard [get]
func (h *UserAdvancedHandler) GetReferralLeaderboard(c *gin.Context) {
	limit := 10 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	leaderboard, err := h.referralService.GetReferralLeaderboard(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get referral leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
		"count":       len(leaderboard),
	})
}

// === SEASON LEAGUES ===

// GetActiveSeasonLeagues godoc
// @Summary Get active season leagues
// @Description Get all active season leagues
// @Tags user-advanced
// @Produce json
// @Success 200 {array} models.SeasonLeague
// @Router /user/season-leagues [get]
func (h *UserAdvancedHandler) GetActiveSeasonLeagues(c *gin.Context) {
	leagues, err := h.seasonLeagueService.GetActiveSeasonLeagues()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get season leagues"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leagues": leagues,
		"count":   len(leagues),
	})
}

// JoinSeasonLeague godoc
// @Summary Join season league
// @Description Join a season league
// @Tags user-advanced
// @Produce json
// @Security BearerAuth
// @Param id path string true "Season League ID"
// @Success 200 {object} map[string]string
// @Router /user/season-leagues/{id}/join [post]
func (h *UserAdvancedHandler) JoinSeasonLeague(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	userModel := user.(*models.User)

	idStr := c.Param("id")
	leagueID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	if err := h.seasonLeagueService.JoinSeasonLeague(leagueID, userModel.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully joined season league"})
}

// GetSeasonLeagueLeaderboard godoc
// @Summary Get season league leaderboard
// @Description Get leaderboard for a season league
// @Tags user-advanced
// @Produce json
// @Param id path string true "Season League ID"
// @Success 200 {array} services.SeasonLeagueEntry
// @Router /user/season-leagues/{id}/leaderboard [get]
func (h *UserAdvancedHandler) GetSeasonLeagueLeaderboard(c *gin.Context) {
	idStr := c.Param("id")
	leagueID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	leaderboard, err := h.seasonLeagueService.GetSeasonLeagueLeaderboard(leagueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
		"count":       len(leaderboard),
	})
}

// === PLAYER ANALYTICS ===

// GetPlayerHeatmap godoc
// @Summary Get player heatmap
// @Description Get player performance heatmap data
// @Tags user-advanced
// @Produce json
// @Param player_id query string true "Player ID"
// @Param game_id query string true "Game ID"
// @Success 200 {object} services.PlayerHeatmapData
// @Router /user/player-heatmap [get]
func (h *UserAdvancedHandler) GetPlayerHeatmap(c *gin.Context) {
	playerIDStr := c.Query("player_id")
	gameIDStr := c.Query("game_id")

	if playerIDStr == "" || gameIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player_id and game_id are required"})
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
		return
	}

	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	heatmap, err := h.playerAnalyticsService.GetPlayerHeatmap(playerID, gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get player heatmap"})
		return
	}

	c.JSON(http.StatusOK, heatmap)
}