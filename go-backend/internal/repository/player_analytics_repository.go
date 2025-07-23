package repository

import (
	"esports-fantasy-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlayerAnalyticsRepository interface {
	Create(analytics *models.PlayerAnalytics) error
	GetByID(id uuid.UUID) (*models.PlayerAnalytics, error)
	GetByPlayerID(playerID uuid.UUID) (*models.PlayerAnalytics, error)
	GetByPlayerAndGame(playerID, gameID uuid.UUID) (*models.PlayerAnalytics, error)
	GetAll() ([]models.PlayerAnalytics, error)
	GetByGameID(gameID uuid.UUID) ([]models.PlayerAnalytics, error)
	GetTopPerformers(gameID uuid.UUID, limit int) ([]models.PlayerAnalytics, error)
	Update(analytics *models.PlayerAnalytics) error
	UpsertPlayerAnalytics(analytics *models.PlayerAnalytics) error
	Delete(id uuid.UUID) error
}

type playerAnalyticsRepository struct {
	db *gorm.DB
}

func NewPlayerAnalyticsRepository(db *gorm.DB) PlayerAnalyticsRepository {
	return &playerAnalyticsRepository{db: db}
}

func (r *playerAnalyticsRepository) Create(analytics *models.PlayerAnalytics) error {
	return r.db.Create(analytics).Error
}

func (r *playerAnalyticsRepository) GetByID(id uuid.UUID) (*models.PlayerAnalytics, error) {
	var analytics models.PlayerAnalytics
	err := r.db.Preload("Player").Preload("Game").Where("id = ?", id).First(&analytics).Error
	if err != nil {
		return nil, err
	}
	return &analytics, nil
}

func (r *playerAnalyticsRepository) GetByPlayerID(playerID uuid.UUID) (*models.PlayerAnalytics, error) {
	var analytics models.PlayerAnalytics
	err := r.db.Preload("Player").Preload("Game").Where("player_id = ?", playerID).First(&analytics).Error
	if err != nil {
		return nil, err
	}
	return &analytics, nil
}

func (r *playerAnalyticsRepository) GetByPlayerAndGame(playerID, gameID uuid.UUID) (*models.PlayerAnalytics, error) {
	var analytics models.PlayerAnalytics
	err := r.db.Preload("Player").Preload("Game").Where("player_id = ? AND game_id = ?", playerID, gameID).First(&analytics).Error
	if err != nil {
		return nil, err
	}
	return &analytics, nil
}

func (r *playerAnalyticsRepository) GetAll() ([]models.PlayerAnalytics, error) {
	var analytics []models.PlayerAnalytics
	err := r.db.Preload("Player").Preload("Game").Find(&analytics).Error
	return analytics, err
}

func (r *playerAnalyticsRepository) GetByGameID(gameID uuid.UUID) ([]models.PlayerAnalytics, error) {
	var analytics []models.PlayerAnalytics
	err := r.db.Preload("Player").Preload("Game").Where("game_id = ?", gameID).Find(&analytics).Error
	return analytics, err
}

func (r *playerAnalyticsRepository) GetTopPerformers(gameID uuid.UUID, limit int) ([]models.PlayerAnalytics, error) {
	var analytics []models.PlayerAnalytics
	err := r.db.Preload("Player").Preload("Game").
		Where("game_id = ?", gameID).
		Order("avg_performance DESC").
		Limit(limit).
		Find(&analytics).Error
	return analytics, err
}

func (r *playerAnalyticsRepository) Update(analytics *models.PlayerAnalytics) error {
	return r.db.Save(analytics).Error
}

func (r *playerAnalyticsRepository) UpsertPlayerAnalytics(analytics *models.PlayerAnalytics) error {
	// Check if analytics already exists
	existing, err := r.GetByPlayerAndGame(analytics.PlayerID, analytics.GameID)
	if err != nil {
		// Create new if doesn't exist
		return r.Create(analytics)
	}
	
	// Update existing
	existing.TotalMatches = analytics.TotalMatches
	existing.TotalKills = analytics.TotalKills
	existing.TotalDeaths = analytics.TotalDeaths
	existing.TotalAssists = analytics.TotalAssists
	existing.AvgPerformance = analytics.AvgPerformance
	existing.Form = analytics.Form
	existing.InjuryStatus = analytics.InjuryStatus
	existing.LastUpdated = analytics.LastUpdated
	
	return r.Update(existing)
}

func (r *playerAnalyticsRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.PlayerAnalytics{}, id).Error
}