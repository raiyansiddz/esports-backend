package services

import (
	"esports-fantasy-backend/config"
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PlayerAnalyticsService interface {
	UpdatePlayerAnalytics(playerID, gameID uuid.UUID, stats *UpdatePlayerAnalyticsRequest) error
	GetPlayerAnalytics(playerID, gameID uuid.UUID) (*models.PlayerAnalytics, error)
	GetTopPerformers(gameID uuid.UUID, limit int) ([]models.PlayerAnalytics, error)
	GetAnalyticsByGame(gameID uuid.UUID) ([]models.PlayerAnalytics, error)
	CalculatePlayerForm(playerID, gameID uuid.UUID) (string, error)
	UpdateInjuryStatus(playerID, gameID uuid.UUID, status string) error
	GetPlayerHeatmap(playerID, gameID uuid.UUID) (*PlayerHeatmapData, error)
}

type UpdatePlayerAnalyticsRequest struct {
	TotalMatches   int     `json:"total_matches"`
	TotalKills     int     `json:"total_kills"`
	TotalDeaths    int     `json:"total_deaths"`
	TotalAssists   int     `json:"total_assists"`
	AvgPerformance float64 `json:"avg_performance"`
	Form           string  `json:"form"`
	InjuryStatus   string  `json:"injury_status"`
}

type PlayerHeatmapData struct {
	PlayerID       uuid.UUID              `json:"player_id"`
	GameID         uuid.UUID              `json:"game_id"`
	PerformanceMap map[string]interface{} `json:"performance_map"`
	FormTrend      []float64              `json:"form_trend"`
	LastUpdated    time.Time              `json:"last_updated"`
}

type playerAnalyticsService struct {
	analyticsRepo repository.PlayerAnalyticsRepository
	playerRepo    repository.PlayerRepository
	gameRepo      repository.GameRepository
	config        *config.Config
}

func NewPlayerAnalyticsService(
	analyticsRepo repository.PlayerAnalyticsRepository,
	playerRepo repository.PlayerRepository,
	gameRepo repository.GameRepository,
	config *config.Config,
) PlayerAnalyticsService {
	return &playerAnalyticsService{
		analyticsRepo: analyticsRepo,
		playerRepo:    playerRepo,
		gameRepo:      gameRepo,
		config:        config,
	}
}

func (s *playerAnalyticsService) UpdatePlayerAnalytics(playerID, gameID uuid.UUID, stats *UpdatePlayerAnalyticsRequest) error {
	// Verify player and game exist
	_, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return fmt.Errorf("player not found: %w", err)
	}

	_, err = s.gameRepo.GetByID(gameID)
	if err != nil {
		return fmt.Errorf("game not found: %w", err)
	}

	analytics := &models.PlayerAnalytics{
		PlayerID:       playerID,
		GameID:         gameID,
		TotalMatches:   stats.TotalMatches,
		TotalKills:     stats.TotalKills,
		TotalDeaths:    stats.TotalDeaths,
		TotalAssists:   stats.TotalAssists,
		AvgPerformance: stats.AvgPerformance,
		Form:           stats.Form,
		InjuryStatus:   stats.InjuryStatus,
		LastUpdated:    time.Now(),
	}

	if err := s.analyticsRepo.UpsertPlayerAnalytics(analytics); err != nil {
		return fmt.Errorf("failed to update player analytics: %w", err)
	}

	return nil
}

func (s *playerAnalyticsService) GetPlayerAnalytics(playerID, gameID uuid.UUID) (*models.PlayerAnalytics, error) {
	return s.analyticsRepo.GetByPlayerAndGame(playerID, gameID)
}

func (s *playerAnalyticsService) GetTopPerformers(gameID uuid.UUID, limit int) ([]models.PlayerAnalytics, error) {
	return s.analyticsRepo.GetTopPerformers(gameID, limit)
}

func (s *playerAnalyticsService) GetAnalyticsByGame(gameID uuid.UUID) ([]models.PlayerAnalytics, error) {
	return s.analyticsRepo.GetByGameID(gameID)
}

func (s *playerAnalyticsService) CalculatePlayerForm(playerID, gameID uuid.UUID) (string, error) {
	analytics, err := s.analyticsRepo.GetByPlayerAndGame(playerID, gameID)
	if err != nil {
		return "stable", nil // Default form if no analytics
	}

	// Simple form calculation based on average performance
	if analytics.AvgPerformance >= 8.0 {
		return "hot", nil
	} else if analytics.AvgPerformance <= 5.0 {
		return "cold", nil
	}

	return "stable", nil
}

func (s *playerAnalyticsService) UpdateInjuryStatus(playerID, gameID uuid.UUID, status string) error {
	analytics, err := s.analyticsRepo.GetByPlayerAndGame(playerID, gameID)
	if err != nil {
		// Create new analytics record if doesn't exist
		analytics = &models.PlayerAnalytics{
			PlayerID:     playerID,
			GameID:       gameID,
			InjuryStatus: status,
			LastUpdated:  time.Now(),
		}
		return s.analyticsRepo.Create(analytics)
	}

	analytics.InjuryStatus = status
	analytics.LastUpdated = time.Now()

	return s.analyticsRepo.Update(analytics)
}

func (s *playerAnalyticsService) GetPlayerHeatmap(playerID, gameID uuid.UUID) (*PlayerHeatmapData, error) {
	analytics, err := s.analyticsRepo.GetByPlayerAndGame(playerID, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player analytics: %w", err)
	}

	// Create performance heatmap data
	performanceMap := map[string]interface{}{
		"kills_per_match":   float64(analytics.TotalKills) / float64(analytics.TotalMatches),
		"deaths_per_match":  float64(analytics.TotalDeaths) / float64(analytics.TotalMatches),
		"assists_per_match": float64(analytics.TotalAssists) / float64(analytics.TotalMatches),
		"kd_ratio":          float64(analytics.TotalKills) / float64(analytics.TotalDeaths),
		"form_status":       analytics.Form,
		"injury_status":     analytics.InjuryStatus,
	}

	// Mock form trend data (in real implementation, this would come from historical data)
	formTrend := []float64{
		analytics.AvgPerformance - 1.0,
		analytics.AvgPerformance - 0.5,
		analytics.AvgPerformance,
		analytics.AvgPerformance + 0.2,
		analytics.AvgPerformance + 0.1,
	}

	return &PlayerHeatmapData{
		PlayerID:       playerID,
		GameID:         gameID,
		PerformanceMap: performanceMap,
		FormTrend:      formTrend,
		LastUpdated:    analytics.LastUpdated,
	}, nil
}