package http

import (
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	tournamentService services.TournamentService
	matchService      services.MatchService
	contestService    services.ContestService
	playerService     services.PlayerService
	scoringService    services.ScoringService
}

func NewAdminHandler(
	tournamentService services.TournamentService,
	matchService services.MatchService,
	contestService services.ContestService,
	playerService services.PlayerService,
	scoringService services.ScoringService,
) *AdminHandler {
	return &AdminHandler{
		tournamentService: tournamentService,
		matchService:      matchService,
		contestService:    contestService,
		playerService:     playerService,
		scoringService:    scoringService,
	}
}

// CreateTournament godoc
// @Summary Create a new tournament
// @Description Admin endpoint to create a new tournament
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tournament body models.Tournament true "Tournament data"
// @Success 201 {object} models.Tournament
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /admin/tournaments [post]
func (h *AdminHandler) CreateTournament(c *gin.Context) {
	var tournament models.Tournament
	if err := c.ShouldBindJSON(&tournament); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	tournament.ID = uuid.New()
	if err := h.tournamentService.CreateTournament(&tournament); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tournament"})
		return
	}

	c.JSON(http.StatusCreated, tournament)
}

// GetTournaments godoc
// @Summary Get all tournaments
// @Description Get list of all tournaments
// @Tags admin
// @Produce json
// @Success 200 {array} models.Tournament
// @Router /admin/tournaments [get]
func (h *AdminHandler) GetTournaments(c *gin.Context) {
	tournaments, err := h.tournamentService.GetTournaments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tournaments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tournaments": tournaments})
}

// CreateMatch godoc
// @Summary Create a new match
// @Description Admin endpoint to create a new match
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param match body models.Match true "Match data"
// @Success 201 {object} models.Match
// @Router /admin/matches [post]
func (h *AdminHandler) CreateMatch(c *gin.Context) {
	var match models.Match
	if err := c.ShouldBindJSON(&match); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	match.ID = uuid.New()
	if err := h.matchService.CreateMatch(&match); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create match"})
		return
	}

	c.JSON(http.StatusCreated, match)
}

// GetMatches godoc
// @Summary Get all matches
// @Description Get list of all matches
// @Tags admin
// @Produce json
// @Success 200 {array} models.Match
// @Router /admin/matches [get]
func (h *AdminHandler) GetMatches(c *gin.Context) {
	matches, err := h.matchService.GetMatches()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get matches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"matches": matches})
}

// UpdateMatchStatus godoc
// @Summary Update match status
// @Description Admin endpoint to update match status
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Match ID"
// @Param status body map[string]string true "New status"
// @Success 200 {object} map[string]string
// @Router /admin/matches/{id}/status [put]
func (h *AdminHandler) UpdateMatchStatus(c *gin.Context) {
	matchIDStr := c.Param("id")
	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match ID"})
		return
	}

	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	status := req["status"]
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	if err := h.matchService.UpdateMatchStatus(matchID, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update match status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match status updated successfully"})
}

// CreateContest godoc
// @Summary Create a new contest
// @Description Admin endpoint to create a new contest
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contest body models.Contest true "Contest data"
// @Success 201 {object} models.Contest
// @Router /admin/contests [post]
func (h *AdminHandler) CreateContest(c *gin.Context) {
	var contest models.Contest
	if err := c.ShouldBindJSON(&contest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	contest.ID = uuid.New()
	if err := h.contestService.CreateContest(&contest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contest"})
		return
	}

	c.JSON(http.StatusCreated, contest)
}

// CreatePlayer godoc
// @Summary Create a new player
// @Description Admin endpoint to create a new player
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param player body models.Player true "Player data"
// @Success 201 {object} models.Player
// @Router /admin/players [post]
func (h *AdminHandler) CreatePlayer(c *gin.Context) {
	var player models.Player
	if err := c.ShouldBindJSON(&player); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	player.ID = uuid.New()
	if err := h.playerService.CreatePlayer(&player); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create player"})
		return
	}

	c.JSON(http.StatusCreated, player)
}

// UpdatePlayerStats godoc
// @Summary Update player match statistics
// @Description Admin endpoint to update player stats for a match
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param matchId path string true "Match ID"
// @Param playerId path string true "Player ID"
// @Param stats body models.UpdateStatsRequest true "Player statistics"
// @Success 200 {object} map[string]string
// @Router /admin/stats/match/{matchId}/player/{playerId} [put]
func (h *AdminHandler) UpdatePlayerStats(c *gin.Context) {
	matchIDStr := c.Param("matchId")
	playerIDStr := c.Param("playerId")

	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match ID"})
		return
	}

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
		return
	}

	var stats models.UpdateStatsRequest
	if err := c.ShouldBindJSON(&stats); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if err := h.scoringService.UpdatePlayerStats(matchID, playerID, &stats); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update player stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Player stats updated successfully"})
}

// CreateESportsTeam godoc
// @Summary Create a new eSports team
// @Description Admin endpoint to create a new eSports team
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param team body models.ESportsTeam true "Team data"
// @Success 201 {object} models.ESportsTeam
// @Router /admin/esports-teams [post]
func (h *AdminHandler) CreateESportsTeam(c *gin.Context) {
	var team models.ESportsTeam
	if err := c.ShouldBindJSON(&team); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	team.ID = uuid.New()
	if err := h.tournamentService.CreateESportsTeam(&team); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team"})
		return
	}

	c.JSON(http.StatusCreated, team)
}

// GetESportsTeams godoc
// @Summary Get all eSports teams
// @Description Get list of all eSports teams with their players
// @Tags admin
// @Produce json
// @Success 200 {array} models.ESportsTeam
// @Router /admin/esports-teams [get]
func (h *AdminHandler) GetESportsTeams(c *gin.Context) {
	teams, err := h.tournamentService.GetESportsTeams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get teams"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}