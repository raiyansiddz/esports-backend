package repository

import (
	"esports-fantasy-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserAchievementRepository interface {
	Create(userAchievement *models.UserAchievement) error
	GetByUserID(userID uuid.UUID) ([]models.UserAchievement, error)
	GetByUserAndAchievement(userID, achievementID uuid.UUID) (*models.UserAchievement, error)
	GetAll() ([]models.UserAchievement, error)
	HasUserUnlockedAchievement(userID, achievementID uuid.UUID) bool
}

type userAchievementRepository struct {
	db *gorm.DB
}

func NewUserAchievementRepository(db *gorm.DB) UserAchievementRepository {
	return &userAchievementRepository{db: db}
}

func (r *userAchievementRepository) Create(userAchievement *models.UserAchievement) error {
	return r.db.Create(userAchievement).Error
}

func (r *userAchievementRepository) GetByUserID(userID uuid.UUID) ([]models.UserAchievement, error) {
	var userAchievements []models.UserAchievement
	err := r.db.Preload("Achievement").Where("user_id = ?", userID).Find(&userAchievements).Error
	return userAchievements, err
}

func (r *userAchievementRepository) GetByUserAndAchievement(userID, achievementID uuid.UUID) (*models.UserAchievement, error) {
	var userAchievement models.UserAchievement
	err := r.db.Where("user_id = ? AND achievement_id = ?", userID, achievementID).First(&userAchievement).Error
	if err != nil {
		return nil, err
	}
	return &userAchievement, nil
}

func (r *userAchievementRepository) GetAll() ([]models.UserAchievement, error) {
	var userAchievements []models.UserAchievement
	err := r.db.Preload("User").Preload("Achievement").Find(&userAchievements).Error
	return userAchievements, err
}

func (r *userAchievementRepository) HasUserUnlockedAchievement(userID, achievementID uuid.UUID) bool {
	var count int64
	r.db.Model(&models.UserAchievement{}).Where("user_id = ? AND achievement_id = ?", userID, achievementID).Count(&count)
	return count > 0
}