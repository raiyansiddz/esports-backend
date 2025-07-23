package http

import (
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminEnhancedHandler struct {
	usernameService services.UsernameService
	gameService     services.GameService
}

func NewAdminEnhancedHandler(usernameService services.UsernameService, gameService services.GameService) *AdminEnhancedHandler {
	return &AdminEnhancedHandler{
		usernameService: usernameService,
		gameService:     gameService,
	}
}

// === USERNAME PREFIX MANAGEMENT ===

// CreateUsernamePrefix godoc
// @Summary Create username prefix
// @Description Admin creates a new username prefix
// @Tags admin-enhanced
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param prefix body map[string]string true "Prefix data"
// @Success 201 {object} models.UsernamePrefix
// @Router /admin/username-prefixes [post]
func (h *AdminEnhancedHandler) CreateUsernamePrefix(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	prefix, exists := req["prefix"]
	if !exists || prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prefix is required"})
		return
	}

	usernamePrefix, err := h.usernameService.CreateUsernamePrefix(prefix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, usernamePrefix)
}

// GetUsernamePrefixes godoc
// @Summary Get all username prefixes
// @Description Admin gets all username prefixes
// @Tags admin-enhanced
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.UsernamePrefix
// @Router /admin/username-prefixes [get]
func (h *AdminEnhancedHandler) GetUsernamePrefixes(c *gin.Context) {
	prefixes, err := h.usernameService.GetUsernamePrefixes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get prefixes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"prefixes": prefixes,
		"count":    len(prefixes),
	})
}

// UpdateUsernamePrefix godoc
// @Summary Update username prefix
// @Description Admin updates a username prefix
// @Tags admin-enhanced
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Prefix ID"
// @Param prefix body map[string]interface{} true "Update data"
// @Success 200 {object} map[string]string
// @Router /admin/username-prefixes/{id} [put]
func (h *AdminEnhancedHandler) UpdateUsernamePrefix(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prefix ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	prefix, prefixExists := req["prefix"].(string)
	isActiveInterface, activeExists := req["is_active"]
	
	if !prefixExists || !activeExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prefix and is_active are required"})
		return
	}

	isActive, ok := isActiveInterface.(bool)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "is_active must be a boolean"})
		return
	}

	if err := h.usernameService.UpdateUsernamePrefix(id, prefix, isActive); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Username prefix updated successfully"})
}

// DeleteUsernamePrefix godoc
// @Summary Delete username prefix
// @Description Admin deletes a username prefix
// @Tags admin-enhanced
// @Produce json
// @Security BearerAuth
// @Param id path string true "Prefix ID"
// @Success 200 {object} map[string]string
// @Router /admin/username-prefixes/{id} [delete]
func (h *AdminEnhancedHandler) DeleteUsernamePrefix(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prefix ID"})
		return
	}

	if err := h.usernameService.DeleteUsernamePrefix(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete prefix"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Username prefix deleted successfully"})
}

// === GAME MANAGEMENT ===

// CreateGame godoc
// @Summary Create new game
// @Description Admin creates a new game
// @Tags admin-enhanced
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param game body models.CreateGameRequest true "Game data"
// @Success 201 {object} models.Game
// @Router /admin/games [post]
func (h *AdminEnhancedHandler) CreateGame(c *gin.Context) {
	var req models.CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	game, err := h.gameService.CreateGame(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, game)
}

// GetGames godoc
// @Summary Get all games
// @Description Admin gets all games
// @Tags admin-enhanced
// @Produce json
// @Security BearerAuth
// @Param active query string false "Filter by active status (true/false)"
// @Success 200 {array} models.Game
// @Router /admin/games [get]
func (h *AdminEnhancedHandler) GetGames(c *gin.Context) {
	activeParam := c.Query("active")
	
	var games []models.Game
	var err error

	if activeParam == "true" {
		games, err = h.gameService.GetActiveGames()
	} else if activeParam == "false" {
		// Get inactive games - would need to implement this in service
		games, err = h.gameService.GetAllGames()
	} else {
		games, err = h.gameService.GetAllGames()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get games"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"games": games,
		"count": len(games),
	})
}

// GetGame godoc
// @Summary Get game by ID
// @Description Admin gets a specific game
// @Tags admin-enhanced
// @Produce json
// @Security BearerAuth
// @Param id path string true "Game ID"
// @Success 200 {object} models.Game
// @Router /admin/games/{id} [get]
func (h *AdminEnhancedHandler) GetGame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	game, err := h.gameService.GetGameByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	c.JSON(http.StatusOK, game)
}

// UpdateGame godoc
// @Summary Update game
// @Description Admin updates a game
// @Tags admin-enhanced
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Game ID"
// @Param game body models.CreateGameRequest true "Game data"
// @Success 200 {object} map[string]string
// @Router /admin/games/{id} [put]
func (h *AdminEnhancedHandler) UpdateGame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	var req models.CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if err := h.gameService.UpdateGame(id, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game updated successfully"})
}

// ToggleGameStatus godoc
// @Summary Toggle game active status
// @Description Admin toggles game active/inactive status
// @Tags admin-enhanced
// @Produce json
// @Security BearerAuth
// @Param id path string true "Game ID"
// @Success 200 {object} map[string]string
// @Router /admin/games/{id}/toggle [patch]
func (h *AdminEnhancedHandler) ToggleGameStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	if err := h.gameService.ToggleGameStatus(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game status toggled successfully"})
}

// DeleteGame godoc
// @Summary Delete game
// @Description Admin deletes a game
// @Tags admin-enhanced
// @Produce json
// @Security BearerAuth
// @Param id path string true "Game ID"
// @Success 200 {object} map[string]string
// @Router /admin/games/{id} [delete]
func (h *AdminEnhancedHandler) DeleteGame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	if err := h.gameService.DeleteGame(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete game"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game deleted successfully"})
}

// === GAME SCORING RULES ===

// CreateScoringRule godoc
// @Summary Create scoring rule
// @Description Admin creates a new scoring rule for a game
// @Tags admin-enhanced
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param rule body models.CreateScoringRuleRequest true "Scoring rule data"
// @Success 201 {object} models.GameScoringRule
// @Router /admin/scoring-rules [post]
func (h *AdminEnhancedHandler) CreateScoringRule(c *gin.Context) {
	var req models.CreateScoringRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	rule, err := h.gameService.CreateScoringRule(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// GetScoringRules godoc
// @Summary Get scoring rules
// @Description Admin gets scoring rules for a game
// @Tags admin-enhanced
// @Produce json
// @Security BearerAuth
// @Param game_id query string true "Game ID"
// @Success 200 {array} models.GameScoringRule
// @Router /admin/scoring-rules [get]
func (h *AdminEnhancedHandler) GetScoringRules(c *gin.Context) {
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

	rules, err := h.gameService.GetScoringRulesByGame(gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get scoring rules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"count": len(rules),
	})
}

// UpdateScoringRule godoc
// @Summary Update scoring rule
// @Description Admin updates a scoring rule
// @Tags admin-enhanced
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Rule ID"
// @Param rule body models.CreateScoringRuleRequest true "Scoring rule data"
// @Success 200 {object} map[string]string
// @Router /admin/scoring-rules/{id} [put]
func (h *AdminEnhancedHandler) UpdateScoringRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	var req models.CreateScoringRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if err := h.gameService.UpdateScoringRule(id, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scoring rule updated successfully"})
}

// DeleteScoringRule godoc
// @Summary Delete scoring rule
// @Description Admin deletes a scoring rule
// @Tags admin-enhanced
// @Produce json
// @Security BearerAuth
// @Param id path string true "Rule ID"
// @Success 200 {object} map[string]string
// @Router /admin/scoring-rules/{id} [delete]
func (h *AdminEnhancedHandler) DeleteScoringRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	if err := h.gameService.DeleteScoringRule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete scoring rule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scoring rule deleted successfully"})
}