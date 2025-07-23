package repository

import (
	"esports-fantasy-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlayerRepository interface {
	CreatePlayer(player *models.Player) error
	GetPlayers() ([]models.Player, error)
	GetPlayersByTeamID(teamID uuid.UUID) ([]models.Player, error)
	GetPlayerByID(id uuid.UUID) (*models.Player, error)
	UpdatePlayer(player *models.Player) error
	DeletePlayer(id uuid.UUID) error
	GetPlayersByIDs(ids []uuid.UUID) ([]models.Player, error)
	GetPlayersByMatchID(matchID string) ([]*models.Player, error)
}

type playerRepository struct {
	db *gorm.DB
}

func NewPlayerRepository(db *gorm.DB) PlayerRepository {
	return &playerRepository{db: db}
}

func (r *playerRepository) CreatePlayer(player *models.Player) error {
	return r.db.Create(player).Error
}

func (r *playerRepository) GetPlayers() ([]models.Player, error) {
	var players []models.Player
	err := r.db.Preload("ESportsTeam").Find(&players).Error
	return players, err
}

func (r *playerRepository) GetPlayersByTeamID(teamID uuid.UUID) ([]models.Player, error) {
	var players []models.Player
	err := r.db.Where("esports_team_id = ?", teamID).Preload("ESportsTeam").Find(&players).Error
	return players, err
}

func (r *playerRepository) GetPlayerByID(id uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.Preload("ESportsTeam").First(&player, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *playerRepository) UpdatePlayer(player *models.Player) error {
	return r.db.Save(player).Error
}

func (r *playerRepository) DeletePlayer(id uuid.UUID) error {
	return r.db.Delete(&models.Player{}, "id = ?", id).Error
}

func (r *playerRepository) GetPlayersByIDs(ids []uuid.UUID) ([]models.Player, error) {
	var players []models.Player
	err := r.db.Where("id IN ?", ids).Preload("ESportsTeam").Find(&players).Error
	return players, err
}

func (r *playerRepository) GetPlayersByMatchID(matchID string) ([]*models.Player, error) {
	var players []*models.Player
	// This query would typically involve joining through matches -> teams -> players
	// For now, return all players as it's a simulation service
	err := r.db.Preload("ESportsTeam").Find(&players).Error
	return players, err
}