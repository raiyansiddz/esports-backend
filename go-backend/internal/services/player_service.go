package services

import (
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type PlayerService interface {
	CreatePlayer(player *models.Player) error
	GetPlayers() ([]models.Player, error)
	GetPlayersByTeamID(teamID uuid.UUID) ([]models.Player, error)
	GetPlayerByID(id uuid.UUID) (*models.Player, error)
	UpdatePlayer(player *models.Player) error
	DeletePlayer(id uuid.UUID) error
}

type playerService struct {
	playerRepo repository.PlayerRepository
}

func NewPlayerService(playerRepo repository.PlayerRepository) PlayerService {
	return &playerService{
		playerRepo: playerRepo,
	}
}

func (s *playerService) CreatePlayer(player *models.Player) error {
	if err := s.playerRepo.CreatePlayer(player); err != nil {
		return fmt.Errorf("failed to create player: %w", err)
	}
	return nil
}

func (s *playerService) GetPlayers() ([]models.Player, error) {
	players, err := s.playerRepo.GetPlayers()
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}
	return players, nil
}

func (s *playerService) GetPlayersByTeamID(teamID uuid.UUID) ([]models.Player, error) {
	players, err := s.playerRepo.GetPlayersByTeamID(teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players by team: %w", err)
	}
	return players, nil
}

func (s *playerService) GetPlayerByID(id uuid.UUID) (*models.Player, error) {
	player, err := s.playerRepo.GetPlayerByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	return player, nil
}

func (s *playerService) UpdatePlayer(player *models.Player) error {
	if err := s.playerRepo.UpdatePlayer(player); err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}
	return nil
}

func (s *playerService) DeletePlayer(id uuid.UUID) error {
	if err := s.playerRepo.DeletePlayer(id); err != nil {
		return fmt.Errorf("failed to delete player: %w", err)
	}
	return nil
}