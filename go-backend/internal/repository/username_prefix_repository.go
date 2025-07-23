package repository

import (
	"esports-fantasy-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UsernamePrefixRepository interface {
	Create(prefix *models.UsernamePrefix) error
	GetByID(id uuid.UUID) (*models.UsernamePrefix, error)
	GetByPrefix(prefix string) (*models.UsernamePrefix, error)
	GetAll() ([]models.UsernamePrefix, error)
	GetActive() ([]models.UsernamePrefix, error)
	Update(prefix *models.UsernamePrefix) error
	Delete(id uuid.UUID) error
}

type usernamePrefixRepository struct {
	db *gorm.DB
}

func NewUsernamePrefixRepository(db *gorm.DB) UsernamePrefixRepository {
	return &usernamePrefixRepository{db: db}
}

func (r *usernamePrefixRepository) Create(prefix *models.UsernamePrefix) error {
	return r.db.Create(prefix).Error
}

func (r *usernamePrefixRepository) GetByID(id uuid.UUID) (*models.UsernamePrefix, error) {
	var prefix models.UsernamePrefix
	err := r.db.First(&prefix, "id = ?", id)
	if err != nil {
		return nil, err.Error
	}
	return &prefix, nil
}

func (r *usernamePrefixRepository) GetByPrefix(prefixStr string) (*models.UsernamePrefix, error) {
	var prefix models.UsernamePrefix
	err := r.db.Where("prefix = ?", prefixStr).First(&prefix)
	if err != nil {
		return nil, err.Error
	}
	return &prefix, nil
}

func (r *usernamePrefixRepository) GetAll() ([]models.UsernamePrefix, error) {
	var prefixes []models.UsernamePrefix
	err := r.db.Find(&prefixes).Error
	return prefixes, err
}

func (r *usernamePrefixRepository) GetActive() ([]models.UsernamePrefix, error) {
	var prefixes []models.UsernamePrefix
	err := r.db.Where("is_active = ?", true).Find(&prefixes).Error
	return prefixes, err
}

func (r *usernamePrefixRepository) Update(prefix *models.UsernamePrefix) error {
	return r.db.Save(prefix).Error
}

func (r *usernamePrefixRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.UsernamePrefix{}, "id = ?", id).Error
}