package repository

import (
	"esports-fantasy-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FantasyTeamRepository interface {
	CreateFantasyTeam(team *models.FantasyTeam) error
	GetUserTeams(userID uuid.UUID) ([]models.FantasyTeam, error)
	GetTeamsByContestID(contestID uuid.UUID) ([]models.FantasyTeam, error)
	GetTeamByID(id uuid.UUID) (*models.FantasyTeam, error)
	UpdateTeam(team *models.FantasyTeam) error
	CheckUserAlreadyInContest(userID, contestID uuid.UUID) (bool, error)
	GetLeaderboard(contestID uuid.UUID, limit int) ([]models.FantasyTeam, error)
}

type fantasyTeamRepository struct {
	db *gorm.DB
}

func NewFantasyTeamRepository(db *gorm.DB) FantasyTeamRepository {
	return &fantasyTeamRepository{db: db}
}

func (r *fantasyTeamRepository) CreateFantasyTeam(team *models.FantasyTeam) error {
	return r.db.Create(team).Error
}

func (r *fantasyTeamRepository) GetUserTeams(userID uuid.UUID) ([]models.FantasyTeam, error) {
	var teams []models.FantasyTeam
	err := r.db.Where("user_id = ?", userID).
		Preload("Contest").
		Preload("Contest.Match").
		Preload("Players").
		Preload("Players.Player").
		Find(&teams).Error
	return teams, err
}

func (r *fantasyTeamRepository) GetTeamsByContestID(contestID uuid.UUID) ([]models.FantasyTeam, error) {
	var teams []models.FantasyTeam
	err := r.db.Where("contest_id = ?", contestID).
		Preload("User").
		Preload("Players").
		Preload("Players.Player").
		Order("total_points DESC").
		Find(&teams).Error
	return teams, err
}

func (r *fantasyTeamRepository) GetTeamByID(id uuid.UUID) (*models.FantasyTeam, error) {
	var team models.FantasyTeam
	err := r.db.Preload("User").
		Preload("Contest").
		Preload("Players").
		Preload("Players.Player").
		First(&team, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *fantasyTeamRepository) UpdateTeam(team *models.FantasyTeam) error {
	return r.db.Save(team).Error
}

func (r *fantasyTeamRepository) CheckUserAlreadyInContest(userID, contestID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.FantasyTeam{}).
		Where("user_id = ? AND contest_id = ?", userID, contestID).
		Count(&count).Error
	return count > 0, err
}

func (r *fantasyTeamRepository) GetLeaderboard(contestID uuid.UUID, limit int) ([]models.FantasyTeam, error) {
	var teams []models.FantasyTeam
	query := r.db.Where("contest_id = ?", contestID).
		Preload("User").
		Order("total_points DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&teams).Error
	return teams, err
}