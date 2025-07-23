package repository

import (
	"esports-fantasy-backend/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MatchRepository interface {
	CreateMatch(match *models.Match) error
	GetMatches() ([]models.Match, error)
	GetMatchesByStatus(status string) ([]models.Match, error)
	GetMatchByID(id uuid.UUID) (*models.Match, error)
	UpdateMatch(match *models.Match) error
	UpdateMatchStatus(id uuid.UUID, status string) error
	GetUpcomingMatches() ([]models.Match, error)
	GetMatchesNeedingLock() ([]models.Match, error)
	
	// New methods for enhanced features
	GetByID(id string) (*models.Match, error)
	Update(match *models.Match) error
}

type matchRepository struct {
	db *gorm.DB
}

func NewMatchRepository(db *gorm.DB) MatchRepository {
	return &matchRepository{db: db}
}

func (r *matchRepository) CreateMatch(match *models.Match) error {
	return r.db.Create(match).Error
}

func (r *matchRepository) GetMatches() ([]models.Match, error) {
	var matches []models.Match
	err := r.db.Preload("Tournament").Order("start_time ASC").Find(&matches).Error
	return matches, err
}

func (r *matchRepository) GetMatchesByStatus(status string) ([]models.Match, error) {
	var matches []models.Match
	err := r.db.Where("status = ?", status).Preload("Tournament").Order("start_time ASC").Find(&matches).Error
	return matches, err
}

func (r *matchRepository) GetMatchByID(id uuid.UUID) (*models.Match, error) {
	var match models.Match
	err := r.db.Preload("Tournament").First(&match, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *matchRepository) UpdateMatch(match *models.Match) error {
	return r.db.Save(match).Error
}

func (r *matchRepository) UpdateMatchStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.Match{}).Where("id = ?", id).Update("status", status).Error
}

func (r *matchRepository) GetUpcomingMatches() ([]models.Match, error) {
	var matches []models.Match
	err := r.db.Where("status = ? AND start_time > ?", "upcoming", time.Now()).
		Preload("Tournament").Order("start_time ASC").Find(&matches).Error
	return matches, err
}

func (r *matchRepository) GetMatchesNeedingLock() ([]models.Match, error) {
	var matches []models.Match
	err := r.db.Where("status = ? AND start_time <= ?", "upcoming", time.Now()).
		Find(&matches).Error
	return matches, err
}

// New methods for enhanced features
func (r *matchRepository) GetByID(id string) (*models.Match, error) {
	var match models.Match
	err := r.db.Preload("Tournament").First(&match, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *matchRepository) Update(match *models.Match) error {
	return r.db.Save(match).Error
}