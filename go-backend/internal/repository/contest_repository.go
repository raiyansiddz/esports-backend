package repository

import (
	"esports-fantasy-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ContestRepository interface {
	CreateContest(contest *models.Contest) error
	GetContestsByMatchID(matchID uuid.UUID) ([]models.Contest, error)
	GetContestByID(id uuid.UUID) (*models.Contest, error)
	UpdateContest(contest *models.Contest) error
	IncrementEntries(contestID uuid.UUID) error
	
	// New methods for enhanced features
	Create(contest *models.Contest) error
	GetByID(id string) (*models.Contest, error)
	Update(contest *models.Contest) error
	GetContestsByStatus(status string) ([]*models.Contest, error)
}

type contestRepository struct {
	db *gorm.DB
}

func NewContestRepository(db *gorm.DB) ContestRepository {
	return &contestRepository{db: db}
}

func (r *contestRepository) CreateContest(contest *models.Contest) error {
	return r.db.Create(contest).Error
}

func (r *contestRepository) GetContestsByMatchID(matchID uuid.UUID) ([]models.Contest, error) {
	var contests []models.Contest
	err := r.db.Where("match_id = ?", matchID).Preload("Match").Find(&contests).Error
	return contests, err
}

func (r *contestRepository) GetContestByID(id uuid.UUID) (*models.Contest, error) {
	var contest models.Contest
	err := r.db.Preload("Match").First(&contest, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &contest, nil
}

func (r *contestRepository) UpdateContest(contest *models.Contest) error {
	return r.db.Save(contest).Error
}

func (r *contestRepository) IncrementEntries(contestID uuid.UUID) error {
	return r.db.Model(&models.Contest{}).Where("id = ?", contestID).
		Update("current_entries", gorm.Expr("current_entries + 1")).Error
}