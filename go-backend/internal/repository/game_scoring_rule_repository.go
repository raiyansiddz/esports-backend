package repository

import (
	"esports-fantasy-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameScoringRuleRepository interface {
	Create(rule *models.GameScoringRule) error
	GetByID(id uuid.UUID) (*models.GameScoringRule, error)
	GetByGameID(gameID uuid.UUID) ([]models.GameScoringRule, error)
	GetActiveByGameID(gameID uuid.UUID) ([]models.GameScoringRule, error)
	Update(rule *models.GameScoringRule) error
	Delete(id uuid.UUID) error
}

type gameScoringRuleRepository struct {
	db *gorm.DB
}

func NewGameScoringRuleRepository(db *gorm.DB) GameScoringRuleRepository {
	return &gameScoringRuleRepository{db: db}
}

func (r *gameScoringRuleRepository) Create(rule *models.GameScoringRule) error {
	return r.db.Create(rule).Error
}

func (r *gameScoringRuleRepository) GetByID(id uuid.UUID) (*models.GameScoringRule, error) {
	var rule models.GameScoringRule
	err := r.db.Preload("Game").First(&rule, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *gameScoringRuleRepository) GetByGameID(gameID uuid.UUID) ([]models.GameScoringRule, error) {
	var rules []models.GameScoringRule
	err := r.db.Where("game_id = ?", gameID).Order("action_type ASC").Find(&rules).Error
	return rules, err
}

func (r *gameScoringRuleRepository) GetActiveByGameID(gameID uuid.UUID) ([]models.GameScoringRule, error) {
	var rules []models.GameScoringRule
	err := r.db.Where("game_id = ? AND is_active = ?", gameID, true).Order("action_type ASC").Find(&rules).Error
	return rules, err
}

func (r *gameScoringRuleRepository) Update(rule *models.GameScoringRule) error {
	return r.db.Save(rule).Error
}

func (r *gameScoringRuleRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.GameScoringRule{}, "id = ?", id).Error
}