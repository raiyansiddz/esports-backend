package services

import (
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type TournamentService interface {
	CreateTournament(tournament *models.Tournament) error
	GetTournaments() ([]models.Tournament, error)
	GetTournamentByID(id uuid.UUID) (*models.Tournament, error)
	UpdateTournament(tournament *models.Tournament) error
	CreateESportsTeam(team *models.ESportsTeam) error
	GetESportsTeams() ([]models.ESportsTeam, error)
}

type tournamentService struct {
	tournamentRepo repository.TournamentRepository
}

func NewTournamentService(tournamentRepo repository.TournamentRepository) TournamentService {
	return &tournamentService{
		tournamentRepo: tournamentRepo,
	}
}

func (s *tournamentService) CreateTournament(tournament *models.Tournament) error {
	if err := s.tournamentRepo.CreateTournament(tournament); err != nil {
		return fmt.Errorf("failed to create tournament: %w", err)
	}
	return nil
}

func (s *tournamentService) GetTournaments() ([]models.Tournament, error) {
	tournaments, err := s.tournamentRepo.GetTournaments()
	if err != nil {
		return nil, fmt.Errorf("failed to get tournaments: %w", err)
	}
	return tournaments, nil
}

func (s *tournamentService) GetTournamentByID(id uuid.UUID) (*models.Tournament, error) {
	tournament, err := s.tournamentRepo.GetTournamentByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}
	return tournament, nil
}

func (s *tournamentService) UpdateTournament(tournament *models.Tournament) error {
	if err := s.tournamentRepo.UpdateTournament(tournament); err != nil {
		return fmt.Errorf("failed to update tournament: %w", err)
	}
	return nil
}

func (s *tournamentService) CreateESportsTeam(team *models.ESportsTeam) error {
	if err := s.tournamentRepo.CreateESportsTeam(team); err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}
	return nil
}

func (s *tournamentService) GetESportsTeams() ([]models.ESportsTeam, error) {
	teams, err := s.tournamentRepo.GetESportsTeams()
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}
	return teams, nil
}