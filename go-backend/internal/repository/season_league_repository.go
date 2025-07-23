package repository

import (
	"esports-fantasy-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SeasonLeagueRepository interface {
	Create(league *models.SeasonLeague) error
	GetByID(id uuid.UUID) (*models.SeasonLeague, error)
	GetAll() ([]models.SeasonLeague, error)
	GetActive() ([]models.SeasonLeague, error)
	GetByGameID(gameID uuid.UUID) ([]models.SeasonLeague, error)
	GetUpcoming() ([]models.SeasonLeague, error)
	Update(league *models.SeasonLeague) error
	Delete(id uuid.UUID) error
}

type seasonLeagueRepository struct {
	db *gorm.DB
}

func NewSeasonLeagueRepository(db *gorm.DB) SeasonLeagueRepository {
	return &seasonLeagueRepository{db: db}
}

func (r *seasonLeagueRepository) Create(league *models.SeasonLeague) error {
	return r.db.Create(league).Error
}

func (r *seasonLeagueRepository) GetByID(id uuid.UUID) (*models.SeasonLeague, error) {
	var league models.SeasonLeague
	err := r.db.Preload("Game").Where("id = ?", id).First(&league).Error
	if err != nil {
		return nil, err
	}
	return &league, nil
}

func (r *seasonLeagueRepository) GetAll() ([]models.SeasonLeague, error) {
	var leagues []models.SeasonLeague
	err := r.db.Preload("Game").Find(&leagues).Error
	return leagues, err
}

func (r *seasonLeagueRepository) GetActive() ([]models.SeasonLeague, error) {
	var leagues []models.SeasonLeague
	err := r.db.Preload("Game").Where("status = ?", "active").Find(&leagues).Error
	return leagues, err
}

func (r *seasonLeagueRepository) GetByGameID(gameID uuid.UUID) ([]models.SeasonLeague, error) {
	var leagues []models.SeasonLeague
	err := r.db.Preload("Game").Where("game_id = ?", gameID).Find(&leagues).Error
	return leagues, err
}

func (r *seasonLeagueRepository) GetUpcoming() ([]models.SeasonLeague, error) {
	var leagues []models.SeasonLeague
	err := r.db.Preload("Game").Where("status = ?", "upcoming").Find(&leagues).Error
	return leagues, err
}

func (r *seasonLeagueRepository) Update(league *models.SeasonLeague) error {
	return r.db.Save(league).Error
}

func (r *seasonLeagueRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.SeasonLeague{}, id).Error
}