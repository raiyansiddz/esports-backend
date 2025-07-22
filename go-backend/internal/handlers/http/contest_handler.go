package http

import (
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ContestHandler struct {
	contestService      services.ContestService
	fantasyTeamService  services.FantasyTeamService
	leaderboardService  services.LeaderboardService
}

func NewContestHandler(
	contestService services.ContestService,
	fantasyTeamService services.FantasyTeamService,
	leaderboardService services.LeaderboardService,
) *ContestHandler {
	return &ContestHandler{
		contestService:     contestService,
		fantasyTeamService: fantasyTeamService,
		leaderboardService: leaderboardService,
	}
}

// GetContestsByMatch godoc
// @Summary Get contests for a match
// @Description Get all contests available for a specific match
// @Tags contests
// @Produce json
// @Param matchId path string true "Match ID"
// @Success 200 {array} models.Contest
// @Router /contests/match/{matchId} [get]
func (h *ContestHandler) GetContestsByMatch(c *gin.Context) {
	matchIDStr := c.Param("matchId")
	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match ID"})
		return
	}

	contests, err := h.contestService.GetContestsByMatch(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get contests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"contests": contests})
}

// GetContestDetails godoc
// @Summary Get contest details
// @Description Get detailed information about a specific contest
// @Tags contests
// @Produce json
// @Param id path string true "Contest ID"
// @Success 200 {object} models.Contest
// @Router /contests/{id} [get]
func (h *ContestHandler) GetContestDetails(c *gin.Context) {
	contestIDStr := c.Param("id")
	contestID, err := uuid.Parse(contestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contest ID"})
		return
	}

	contest, err := h.contestService.GetContestByID(contestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get contest details"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"contest": contest})
}

// CreateFantasyTeam godoc
// @Summary Create a fantasy team
// @Description Create a new fantasy team for a contest
// @Tags fantasy
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param team body models.CreateTeamRequest true "Fantasy team data"
// @Success 201 {object} models.FantasyTeam
// @Router /fantasy/teams [post]
func (h *ContestHandler) CreateFantasyTeam(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	userModel := user.(*models.User)

	var req models.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	team, err := h.fantasyTeamService.CreateFantasyTeam(userModel.ID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"team": team})
}

// GetUserTeams godoc
// @Summary Get user's fantasy teams
// @Description Get all fantasy teams created by the authenticated user
// @Tags fantasy
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.FantasyTeam
// @Router /fantasy/teams [get]
func (h *ContestHandler) GetUserTeams(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	userModel := user.(*models.User)

	teams, err := h.fantasyTeamService.GetUserTeams(userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get teams"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

// GetLeaderboard godoc
// @Summary Get contest leaderboard
// @Description Get the current leaderboard for a contest
// @Tags contests
// @Produce json
// @Param id path string true "Contest ID"
// @Param limit query int false "Number of entries to return (default: 100)"
// @Success 200 {array} services.LeaderboardEntry
// @Router /contests/{id}/leaderboard [get]
func (h *ContestHandler) GetLeaderboard(c *gin.Context) {
	contestIDStr := c.Param("id")
	contestID, err := uuid.Parse(contestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contest ID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	leaderboard, err := h.leaderboardService.GetLeaderboard(contestID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"leaderboard": leaderboard})
}