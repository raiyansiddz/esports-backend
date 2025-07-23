package http

import (
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdvancedAdminHandler struct {
	achievementService      services.AchievementService
	contestTemplateService  services.ContestTemplateService
	playerAnalyticsService  services.PlayerAnalyticsService
	seasonLeagueService     services.SeasonLeagueService
}

func NewAdvancedAdminHandler(
	achievementService services.AchievementService,
	contestTemplateService services.ContestTemplateService,
	playerAnalyticsService services.PlayerAnalyticsService,
	seasonLeagueService services.SeasonLeagueService,
) *AdvancedAdminHandler {
	return &AdvancedAdminHandler{
		achievementService:     achievementService,
		contestTemplateService: contestTemplateService,
		playerAnalyticsService: playerAnalyticsService,
		seasonLeagueService:    seasonLeagueService,
	}
}

// === ACHIEVEMENT MANAGEMENT ===

// CreateAchievement godoc
// @Summary Create achievement
// @Description Admin creates a new achievement
// @Tags admin-advanced
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param achievement body services.CreateAchievementRequest true "Achievement data"
// @Success 201 {object} models.Achievement
// @Router /admin/achievements [post]
func (h *AdvancedAdminHandler) CreateAchievement(c *gin.Context) {
	var req services.CreateAchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	achievement, err := h.achievementService.CreateAchievement(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, achievement)
}

// GetAchievements godoc
// @Summary Get all achievements
// @Description Admin gets all achievements
// @Tags admin-advanced
// @Produce json
// @Security BearerAuth
// @Param active query string false "Filter by active status (true/false)"
// @Success 200 {array} models.Achievement
// @Router /admin/achievements [get]
func (h *AdvancedAdminHandler) GetAchievements(c *gin.Context) {
	activeParam := c.Query("active")
	
	var achievements []models.Achievement
	var err error

	if activeParam == "true" {
		achievements, err = h.achievementService.GetActiveAchievements()
	} else {
		achievements, err = h.achievementService.GetAllAchievements()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get achievements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"achievements": achievements,
		"count":        len(achievements),
	})
}

// UpdateAchievement godoc
// @Summary Update achievement
// @Description Admin updates an achievement
// @Tags admin-advanced
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Param achievement body services.CreateAchievementRequest true "Achievement data"
// @Success 200 {object} map[string]string
// @Router /admin/achievements/{id} [put]
func (h *AdvancedAdminHandler) UpdateAchievement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid achievement ID"})
		return
	}

	var req services.CreateAchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if err := h.achievementService.UpdateAchievement(id, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Achievement updated successfully"})
}

// ToggleAchievementStatus godoc
// @Summary Toggle achievement status
// @Description Admin toggles achievement active/inactive status
// @Tags admin-advanced
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Success 200 {object} map[string]string
// @Router /admin/achievements/{id}/toggle [patch]
func (h *AdvancedAdminHandler) ToggleAchievementStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid achievement ID"})
		return
	}

	if err := h.achievementService.ToggleAchievementStatus(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Achievement status toggled successfully"})
}

// === CONTEST TEMPLATE MANAGEMENT ===

// CreateContestTemplate godoc
// @Summary Create contest template
// @Description Admin creates a new contest template
// @Tags admin-advanced
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param template body services.CreateContestTemplateRequest true "Contest template data"
// @Success 201 {object} models.ContestTemplate
// @Router /admin/contest-templates [post]
func (h *AdvancedAdminHandler) CreateContestTemplate(c *gin.Context) {
	var req services.CreateContestTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	template, err := h.contestTemplateService.CreateTemplate(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetContestTemplates godoc
// @Summary Get all contest templates
// @Description Admin gets all contest templates
// @Tags admin-advanced
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ContestTemplate
// @Router /admin/contest-templates [get]
func (h *AdvancedAdminHandler) GetContestTemplates(c *gin.Context) {
	templates, err := h.contestTemplateService.GetAllTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"count":     len(templates),
	})
}

// === SEASON LEAGUE MANAGEMENT ===

// CreateSeasonLeague godoc
// @Summary Create season league
// @Description Admin creates a new season league
// @Tags admin-advanced
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param league body services.CreateSeasonLeagueRequest true "Season league data"
// @Success 201 {object} models.SeasonLeague
// @Router /admin/season-leagues [post]
func (h *AdvancedAdminHandler) CreateSeasonLeague(c *gin.Context) {
	var req services.CreateSeasonLeagueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	league, err := h.seasonLeagueService.CreateSeasonLeague(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, league)
}

// GetSeasonLeagues godoc
// @Summary Get all season leagues
// @Description Admin gets all season leagues
// @Tags admin-advanced
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.SeasonLeague
// @Router /admin/season-leagues [get]
func (h *AdvancedAdminHandler) GetSeasonLeagues(c *gin.Context) {
	leagues, err := h.seasonLeagueService.GetAllSeasonLeagues()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get season leagues"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leagues": leagues,
		"count":   len(leagues),
	})
}

// === PLAYER ANALYTICS MANAGEMENT ===

// UpdatePlayerAnalytics godoc
// @Summary Update player analytics
// @Description Admin updates player analytics data
// @Tags admin-advanced
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param player_id query string true "Player ID"
// @Param game_id query string true "Game ID"
// @Param analytics body services.UpdatePlayerAnalyticsRequest true "Analytics data"
// @Success 200 {object} map[string]string
// @Router /admin/player-analytics [put]
func (h *AdvancedAdminHandler) UpdatePlayerAnalytics(c *gin.Context) {
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

	var req services.UpdatePlayerAnalyticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if err := h.playerAnalyticsService.UpdatePlayerAnalytics(playerID, gameID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Player analytics updated successfully"})
}

// GetTopPerformers godoc
// @Summary Get top performers
// @Description Admin gets top performing players for a game
// @Tags admin-advanced
// @Produce json
// @Security BearerAuth
// @Param game_id query string true "Game ID"
// @Param limit query int false "Limit results (default: 10)"
// @Success 200 {array} models.PlayerAnalytics
// @Router /admin/top-performers [get]
func (h *AdvancedAdminHandler) GetTopPerformers(c *gin.Context) {
	gameIDStr := c.Query("game_id")
	if gameIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id is required"})
		return
	}

	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	limit := 10 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	performers, err := h.playerAnalyticsService.GetTopPerformers(gameID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top performers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"performers": performers,
		"count":      len(performers),
	})
}