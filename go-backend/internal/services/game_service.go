package services

import (
	"encoding/json"
	"esports-fantasy-backend/config"
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type GameService interface {
	CreateGame(req *models.CreateGameRequest) (*models.Game, error)
	GetAllGames() ([]models.Game, error)
	GetActiveGames() ([]models.Game, error)
	GetGameByID(id uuid.UUID) (*models.Game, error)
	UpdateGame(id uuid.UUID, req *models.CreateGameRequest) error
	ToggleGameStatus(id uuid.UUID) error
	DeleteGame(id uuid.UUID) error
	
	// Scoring rules management
	CreateScoringRule(req *models.CreateScoringRuleRequest) (*models.GameScoringRule, error)
	GetScoringRulesByGame(gameID uuid.UUID) ([]models.GameScoringRule, error)
	UpdateScoringRule(id uuid.UUID, req *models.CreateScoringRuleRequest) error
	DeleteScoringRule(id uuid.UUID) error
}

type gameService struct {
	gameRepo       repository.GameRepository
	scoringRepo    repository.GameScoringRuleRepository
	config         *config.Config
}

func NewGameService(gameRepo repository.GameRepository, scoringRepo repository.GameScoringRuleRepository, config *config.Config) GameService {
	return &gameService{
		gameRepo:    gameRepo,
		scoringRepo: scoringRepo,
		config:      config,
	}
}

func (s *gameService) CreateGame(req *models.CreateGameRequest) (*models.Game, error) {
	// Check if game name already exists
	existing, err := s.gameRepo.GetByName(req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("game with name '%s' already exists", req.Name)
	}

	// Convert player roles to JSON
	rolesJSON, err := json.Marshal(req.PlayerRoles)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize player roles: %w", err)
	}

	game := &models.Game{
		Name:              req.Name,
		DisplayName:       req.DisplayName,
		Icon:              req.Icon,
		IsActive:          true,
		MaxPlayersPerTeam: req.MaxPlayersPerTeam,
		MinPlayersPerTeam: req.MinPlayersPerTeam,
		PlayerRoles:       string(rolesJSON),
		ScoringRules:      "{}",
	}

	if err := s.gameRepo.Create(game); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	return game, nil
}

func (s *gameService) GetAllGames() ([]models.Game, error) {
	return s.gameRepo.GetAll()
}

func (s *gameService) GetActiveGames() ([]models.Game, error) {
	return s.gameRepo.GetActive()
}

func (s *gameService) GetGameByID(id uuid.UUID) (*models.Game, error) {
	return s.gameRepo.GetByID(id)
}

func (s *gameService) UpdateGame(id uuid.UUID, req *models.CreateGameRequest) error {
	game, err := s.gameRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Check if new name conflicts with another game
	if game.Name != req.Name {
		existing, err := s.gameRepo.GetByName(req.Name)
		if err == nil && existing != nil && existing.ID != id {
			return fmt.Errorf("game with name '%s' already exists", req.Name)
		}
	}

	// Convert player roles to JSON
	rolesJSON, err := json.Marshal(req.PlayerRoles)
	if err != nil {
		return fmt.Errorf("failed to serialize player roles: %w", err)
	}

	game.Name = req.Name
	game.DisplayName = req.DisplayName
	game.Icon = req.Icon
	game.MaxPlayersPerTeam = req.MaxPlayersPerTeam
	game.MinPlayersPerTeam = req.MinPlayersPerTeam
	game.PlayerRoles = string(rolesJSON)

	if err := s.gameRepo.Update(game); err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	return nil
}

func (s *gameService) ToggleGameStatus(id uuid.UUID) error {
	game, err := s.gameRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	game.IsActive = !game.IsActive

	if err := s.gameRepo.Update(game); err != nil {
		return fmt.Errorf("failed to update game status: %w", err)
	}

	return nil
}

func (s *gameService) DeleteGame(id uuid.UUID) error {
	// Check if game has any tournaments
	// This should be implemented when tournament service is updated
	return s.gameRepo.Delete(id)
}

// Scoring rules management
func (s *gameService) CreateScoringRule(req *models.CreateScoringRuleRequest) (*models.GameScoringRule, error) {
	// Verify game exists
	_, err := s.gameRepo.GetByID(req.GameID)
	if err != nil {
		return nil, fmt.Errorf("game not found: %w", err)
	}

	rule := &models.GameScoringRule{
		GameID:      req.GameID,
		ActionType:  req.ActionType,
		Points:      req.Points,
		Description: req.Description,
		IsActive:    true,
	}

	if err := s.scoringRepo.Create(rule); err != nil {
		return nil, fmt.Errorf("failed to create scoring rule: %w", err)
	}

	return rule, nil
}

func (s *gameService) GetScoringRulesByGame(gameID uuid.UUID) ([]models.GameScoringRule, error) {
	return s.scoringRepo.GetByGameID(gameID)
}

func (s *gameService) UpdateScoringRule(id uuid.UUID, req *models.CreateScoringRuleRequest) error {
	rule, err := s.scoringRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get scoring rule: %w", err)
	}

	// Verify game exists if changed
	if rule.GameID != req.GameID {
		_, err := s.gameRepo.GetByID(req.GameID)
		if err != nil {
			return fmt.Errorf("game not found: %w", err)
		}
	}

	rule.GameID = req.GameID
	rule.ActionType = req.ActionType
	rule.Points = req.Points
	rule.Description = req.Description

	if err := s.scoringRepo.Update(rule); err != nil {
		return fmt.Errorf("failed to update scoring rule: %w", err)
	}

	return nil
}

func (s *gameService) DeleteScoringRule(id uuid.UUID) error {
	return s.scoringRepo.Delete(id)
}