package services

import (
	"esports-fantasy-backend/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type FantasyTeamService interface {
	CreateFantasyTeam(userID uuid.UUID, req *models.CreateTeamRequest) (*models.FantasyTeam, error)
	GetUserTeams(userID uuid.UUID) ([]models.FantasyTeam, error)
	GetTeamByID(id uuid.UUID) (*models.FantasyTeam, error)
	ValidateTeamCreation(req *models.CreateTeamRequest) error
}

type fantasyTeamService struct {
	fantasyTeamRepo repository.FantasyTeamRepository
	playerRepo      repository.PlayerRepository
}

func NewFantasyTeamService(fantasyTeamRepo repository.FantasyTeamRepository, playerRepo repository.PlayerRepository) FantasyTeamService {
	return &fantasyTeamService{
		fantasyTeamRepo: fantasyTeamRepo,
		playerRepo:      playerRepo,
	}
}

func (s *fantasyTeamService) CreateFantasyTeam(userID uuid.UUID, req *models.CreateTeamRequest) (*models.FantasyTeam, error) {
	// Validate team creation
	if err := s.ValidateTeamCreation(req); err != nil {
		return nil, err
	}

	// Check if user already has a team in this contest
	exists, err := s.fantasyTeamRepo.CheckUserAlreadyInContest(userID, req.ContestID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing team: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user already has a team in this contest")
	}

	// Create fantasy team
	fantasyTeam := &models.FantasyTeam{
		ID:        uuid.New(),
		UserID:    userID,
		ContestID: req.ContestID,
		TeamName:  req.TeamName,
	}

	if err := s.fantasyTeamRepo.CreateFantasyTeam(fantasyTeam); err != nil {
		return nil, fmt.Errorf("failed to create fantasy team: %w", err)
	}

	// Add players to the team
	for _, playerID := range req.PlayerIDs {
		fantasyPlayer := models.FantasyTeamPlayer{
			FantasyTeamID: fantasyTeam.ID,
			PlayerID:      playerID,
			IsCaptain:     playerID == req.CaptainID,
			IsViceCaptain: playerID == req.ViceCaptainID,
		}

		// Note: This should be handled in a transaction in production
		// For now, we'll assume a method exists to create fantasy team players
		_ = fantasyPlayer // Placeholder for now
	}

	return fantasyTeam, nil
}

func (s *fantasyTeamService) GetUserTeams(userID uuid.UUID) ([]models.FantasyTeam, error) {
	teams, err := s.fantasyTeamRepo.GetUserTeams(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user teams: %w", err)
	}
	return teams, nil
}

func (s *fantasyTeamService) GetTeamByID(id uuid.UUID) (*models.FantasyTeam, error) {
	team, err := s.fantasyTeamRepo.GetTeamByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	return team, nil
}

func (s *fantasyTeamService) ValidateTeamCreation(req *models.CreateTeamRequest) error {
	// Check if we have exactly 5 players (BGMI/PUBG standard)
	if len(req.PlayerIDs) != 5 {
		return fmt.Errorf("team must have exactly 5 players")
	}

	// Check if captain and vice-captain are in the player list
	captainFound := false
	viceCaptainFound := false
	for _, playerID := range req.PlayerIDs {
		if playerID == req.CaptainID {
			captainFound = true
		}
		if playerID == req.ViceCaptainID {
			viceCaptainFound = true
		}
	}

	if !captainFound {
		return fmt.Errorf("captain must be one of the selected players")
	}
	if !viceCaptainFound {
		return fmt.Errorf("vice captain must be one of the selected players")
	}
	if req.CaptainID == req.ViceCaptainID {
		return fmt.Errorf("captain and vice captain must be different players")
	}

	// Check if all players exist
	players, err := s.playerRepo.GetPlayersByIDs(req.PlayerIDs)
	if err != nil {
		return fmt.Errorf("failed to validate players: %w", err)
	}
	if len(players) != len(req.PlayerIDs) {
		return fmt.Errorf("some players not found")
	}

	// Calculate total credit value (assuming 100 credits budget)
	totalCredits := 0.0
	for _, player := range players {
		totalCredits += player.CreditValue
	}
	if totalCredits > 100.0 {
		return fmt.Errorf("team exceeds credit limit (%.1f/100)", totalCredits)
	}

	return nil
}