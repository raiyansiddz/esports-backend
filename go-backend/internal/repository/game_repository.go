package repository

import (
	"esports-fantasy-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameRepository interface {
	Create(game *models.Game) error
	GetByID(id uuid.UUID) (*models.Game, error)
	GetByName(name string) (*models.Game, error)
	GetAll() ([]models.Game, error)
	GetActive() ([]models.Game, error)
	Update(game *models.Game) error
	Delete(id uuid.UUID) error
}

type gameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) GameRepository {
	return &gameRepository{db: db}
}

func (r *gameRepository) Create(game *models.Game) error {
	return r.db.Create(game).Error
}

func (r *gameRepository) GetByID(id uuid.UUID) (*models.Game, error) {
	var game models.Game
	err := r.db.First(&game, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *gameRepository) GetByName(name string) (*models.Game, error) {
	var game models.Game
	err := r.db.Where("name = ?", name).First(&game).Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *gameRepository) GetAll() ([]models.Game, error) {
	var games []models.Game
	err := r.db.Find(&games).Error
	return games, err
}

func (r *gameRepository) GetActive() ([]models.Game, error) {
	var games []models.Game
	err := r.db.Where("is_active = ?", true).Find(&games).Error
	return games, err
}

func (r *gameRepository) Update(game *models.Game) error {
	return r.db.Save(game).Error
}

func (r *gameRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Game{}, "id = ?", id).Error
}