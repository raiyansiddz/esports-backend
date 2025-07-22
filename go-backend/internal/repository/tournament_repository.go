package repository

import (
	"esports-fantasy-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TournamentRepository interface {
	CreateTournament(tournament *models.Tournament) error
	GetTournaments() ([]models.Tournament, error)
	GetTournamentByID(id uuid.UUID) (*models.Tournament, error)
	UpdateTournament(tournament *models.Tournament) error
	DeleteTournament(id uuid.UUID) error
	CreateESportsTeam(team *models.ESportsTeam) error
	GetESportsTeams() ([]models.ESportsTeam, error)
	GetESportsTeamByID(id uuid.UUID) (*models.ESportsTeam, error)
}

type tournamentRepository struct {
	db *gorm.DB
}

func NewTournamentRepository(db *gorm.DB) TournamentRepository {
	return &tournamentRepository{db: db}
}

func (r *tournamentRepository) CreateTournament(tournament *models.Tournament) error {
	return r.db.Create(tournament).Error
}

func (r *tournamentRepository) GetTournaments() ([]models.Tournament, error) {
	var tournaments []models.Tournament
	err := r.db.Find(&tournaments).Error
	return tournaments, err
}

func (r *tournamentRepository) GetTournamentByID(id uuid.UUID) (*models.Tournament, error) {
	var tournament models.Tournament
	err := r.db.First(&tournament, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tournament, nil
}

func (r *tournamentRepository) UpdateTournament(tournament *models.Tournament) error {
	return r.db.Save(tournament).Error
}

func (r *tournamentRepository) DeleteTournament(id uuid.UUID) error {
	return r.db.Delete(&models.Tournament{}, "id = ?", id).Error
}

func (r *tournamentRepository) CreateESportsTeam(team *models.ESportsTeam) error {
	return r.db.Create(team).Error
}

func (r *tournamentRepository) GetESportsTeams() ([]models.ESportsTeam, error) {
	var teams []models.ESportsTeam
	err := r.db.Preload("Players").Find(&teams).Error
	return teams, err
}

func (r *tournamentRepository) GetESportsTeamByID(id uuid.UUID) (*models.ESportsTeam, error) {
	var team models.ESportsTeam
	err := r.db.Preload("Players").First(&team, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}