package repository

import (
	"esports-fantasy-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ContestTemplateRepository interface {
	Create(template *models.ContestTemplate) error
	GetByID(id uuid.UUID) (*models.ContestTemplate, error)
	GetAll() ([]models.ContestTemplate, error)
	GetActive() ([]models.ContestTemplate, error)
	GetByGameID(gameID uuid.UUID) ([]models.ContestTemplate, error)
	GetVIPTemplates() ([]models.ContestTemplate, error)
	Update(template *models.ContestTemplate) error
	Delete(id uuid.UUID) error
}

type contestTemplateRepository struct {
	db *gorm.DB
}

func NewContestTemplateRepository(db *gorm.DB) ContestTemplateRepository {
	return &contestTemplateRepository{db: db}
}

func (r *contestTemplateRepository) Create(template *models.ContestTemplate) error {
	return r.db.Create(template).Error
}

func (r *contestTemplateRepository) GetByID(id uuid.UUID) (*models.ContestTemplate, error) {
	var template models.ContestTemplate
	err := r.db.Preload("Game").Where("id = ?", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *contestTemplateRepository) GetAll() ([]models.ContestTemplate, error) {
	var templates []models.ContestTemplate
	err := r.db.Preload("Game").Find(&templates).Error
	return templates, err
}

func (r *contestTemplateRepository) GetActive() ([]models.ContestTemplate, error) {
	var templates []models.ContestTemplate
	err := r.db.Preload("Game").Where("is_active = ?", true).Find(&templates).Error
	return templates, err
}

func (r *contestTemplateRepository) GetByGameID(gameID uuid.UUID) ([]models.ContestTemplate, error) {
	var templates []models.ContestTemplate
	err := r.db.Preload("Game").Where("game_id = ? AND is_active = ?", gameID, true).Find(&templates).Error
	return templates, err
}

func (r *contestTemplateRepository) GetVIPTemplates() ([]models.ContestTemplate, error) {
	var templates []models.ContestTemplate
	err := r.db.Preload("Game").Where("is_vip = ? AND is_active = ?", true, true).Find(&templates).Error
	return templates, err
}

func (r *contestTemplateRepository) Update(template *models.ContestTemplate) error {
	return r.db.Save(template).Error
}

func (r *contestTemplateRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.ContestTemplate{}, id).Error
}