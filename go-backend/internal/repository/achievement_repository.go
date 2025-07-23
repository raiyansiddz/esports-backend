package repository

import (
	"esports-fantasy-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AchievementRepository interface {
	Create(achievement *models.Achievement) error
	GetByID(id uuid.UUID) (*models.Achievement, error)
	GetAll() ([]models.Achievement, error)
	GetActive() ([]models.Achievement, error)
	Update(achievement *models.Achievement) error
	Delete(id uuid.UUID) error
}

type achievementRepository struct {
	db *gorm.DB
}

func NewAchievementRepository(db *gorm.DB) AchievementRepository {
	return &achievementRepository{db: db}
}

func (r *achievementRepository) Create(achievement *models.Achievement) error {
	return r.db.Create(achievement).Error
}

func (r *achievementRepository) GetByID(id uuid.UUID) (*models.Achievement, error) {
	var achievement models.Achievement
	err := r.db.Where("id = ?", id).First(&achievement).Error
	if err != nil {
		return nil, err
	}
	return &achievement, nil
}

func (r *achievementRepository) GetAll() ([]models.Achievement, error) {
	var achievements []models.Achievement
	err := r.db.Find(&achievements).Error
	return achievements, err
}

func (r *achievementRepository) GetActive() ([]models.Achievement, error) {
	var achievements []models.Achievement
	err := r.db.Where("is_active = ?", true).Find(&achievements).Error
	return achievements, err
}

func (r *achievementRepository) Update(achievement *models.Achievement) error {
	return r.db.Save(achievement).Error
}

func (r *achievementRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Achievement{}, id).Error
}