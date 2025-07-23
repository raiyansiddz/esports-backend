package services

import (
	"esports-fantasy-backend/config"
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SeasonLeagueService interface {
	CreateSeasonLeague(req *CreateSeasonLeagueRequest) (*models.SeasonLeague, error)
	GetAllSeasonLeagues() ([]models.SeasonLeague, error)
	GetActiveSeasonLeagues() ([]models.SeasonLeague, error)
	GetUpcomingSeasonLeagues() ([]models.SeasonLeague, error)
	GetSeasonLeaguesByGame(gameID uuid.UUID) ([]models.SeasonLeague, error)
	GetSeasonLeagueByID(id uuid.UUID) (*models.SeasonLeague, error)
	UpdateSeasonLeague(id uuid.UUID, req *CreateSeasonLeagueRequest) error
	UpdateSeasonLeagueStatus(id uuid.UUID, status string) error
	DeleteSeasonLeague(id uuid.UUID) error
	JoinSeasonLeague(leagueID, userID uuid.UUID) error
	GetSeasonLeagueLeaderboard(leagueID uuid.UUID) ([]SeasonLeagueEntry, error)
}

type CreateSeasonLeagueRequest struct {
	Name            string    `json:"name" binding:"required"`
	GameID          uuid.UUID `json:"game_id" binding:"required"`
	StartDate       time.Time `json:"start_date" binding:"required"`
	EndDate         time.Time `json:"end_date" binding:"required"`
	EntryFee        float64   `json:"entry_fee" binding:"required"`
	PrizePool       float64   `json:"prize_pool" binding:"required"`
	MaxParticipants int       `json:"max_participants" binding:"required"`
}

type SeasonLeagueEntry struct {
	UserID       uuid.UUID `json:"user_id"`
	Username     string    `json:"username"`
	TotalPoints  int64     `json:"total_points"`
	Rank         int       `json:"rank"`
	TotalPrize   float64   `json:"total_prize"`
	Participated time.Time `json:"participated"`
}

type seasonLeagueService struct {
	leagueRepo repository.SeasonLeagueRepository
	gameRepo   repository.GameRepository
	userRepo   repository.UserRepository
	config     *config.Config
}

func NewSeasonLeagueService(
	leagueRepo repository.SeasonLeagueRepository,
	gameRepo repository.GameRepository,
	userRepo repository.UserRepository,
	config *config.Config,
) SeasonLeagueService {
	return &seasonLeagueService{
		leagueRepo: leagueRepo,
		gameRepo:   gameRepo,
		userRepo:   userRepo,
		config:     config,
	}
}

func (s *seasonLeagueService) CreateSeasonLeague(req *CreateSeasonLeagueRequest) (*models.SeasonLeague, error) {
	// Verify game exists
	_, err := s.gameRepo.GetByID(req.GameID)
	if err != nil {
		return nil, fmt.Errorf("game not found: %w", err)
	}

	// Validate dates
	if req.EndDate.Before(req.StartDate) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	// Set status based on dates
	var status string
	now := time.Now()
	if req.StartDate.After(now) {
		status = "upcoming"
	} else if req.EndDate.Before(now) {
		status = "completed"
	} else {
		status = "active"
	}

	league := &models.SeasonLeague{
		Name:            req.Name,
		GameID:          req.GameID,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		EntryFee:        req.EntryFee,
		PrizePool:       req.PrizePool,
		MaxParticipants: req.MaxParticipants,
		Status:          status,
	}

	if err := s.leagueRepo.Create(league); err != nil {
		return nil, fmt.Errorf("failed to create season league: %w", err)
	}

	return league, nil
}

func (s *seasonLeagueService) GetAllSeasonLeagues() ([]models.SeasonLeague, error) {
	return s.leagueRepo.GetAll()
}

func (s *seasonLeagueService) GetActiveSeasonLeagues() ([]models.SeasonLeague, error) {
	return s.leagueRepo.GetActive()
}

func (s *seasonLeagueService) GetUpcomingSeasonLeagues() ([]models.SeasonLeague, error) {
	return s.leagueRepo.GetUpcoming()
}

func (s *seasonLeagueService) GetSeasonLeaguesByGame(gameID uuid.UUID) ([]models.SeasonLeague, error) {
	return s.leagueRepo.GetByGameID(gameID)
}

func (s *seasonLeagueService) GetSeasonLeagueByID(id uuid.UUID) (*models.SeasonLeague, error) {
	return s.leagueRepo.GetByID(id)
}

func (s *seasonLeagueService) UpdateSeasonLeague(id uuid.UUID, req *CreateSeasonLeagueRequest) error {
	league, err := s.leagueRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get season league: %w", err)
	}

	// Verify game exists if changed
	if league.GameID != req.GameID {
		_, err := s.gameRepo.GetByID(req.GameID)
		if err != nil {
			return fmt.Errorf("game not found: %w", err)
		}
	}

	// Validate dates
	if req.EndDate.Before(req.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}

	league.Name = req.Name
	league.GameID = req.GameID
	league.StartDate = req.StartDate
	league.EndDate = req.EndDate
	league.EntryFee = req.EntryFee
	league.PrizePool = req.PrizePool
	league.MaxParticipants = req.MaxParticipants

	if err := s.leagueRepo.Update(league); err != nil {
		return fmt.Errorf("failed to update season league: %w", err)
	}

	return nil
}

func (s *seasonLeagueService) UpdateSeasonLeagueStatus(id uuid.UUID, status string) error {
	league, err := s.leagueRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get season league: %w", err)
	}

	// Validate status
	validStatuses := []string{"upcoming", "active", "completed"}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid status: must be one of %v", validStatuses)
	}

	league.Status = status

	if err := s.leagueRepo.Update(league); err != nil {
		return fmt.Errorf("failed to update season league status: %w", err)
	}

	return nil
}

func (s *seasonLeagueService) DeleteSeasonLeague(id uuid.UUID) error {
	return s.leagueRepo.Delete(id)
}

func (s *seasonLeagueService) JoinSeasonLeague(leagueID, userID uuid.UUID) error {
	// Get league
	league, err := s.leagueRepo.GetByID(leagueID)
	if err != nil {
		return fmt.Errorf("season league not found: %w", err)
	}

	// Check if league is active
	if league.Status != "upcoming" && league.Status != "active" {
		return fmt.Errorf("cannot join completed season league")
	}

	// Check if league is full (this would require a participants table in a real implementation)
	// For now, we'll skip this check

	// Get user to check balance
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if user has enough balance
	if user.WalletBalance < league.EntryFee {
		return fmt.Errorf("insufficient wallet balance")
	}

	// Deduct entry fee (in a real implementation, this would be a transaction)
	user.WalletBalance -= league.EntryFee
	if err := s.userRepo.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to deduct entry fee: %w", err)
	}

	// In a real implementation, you would create a SeasonLeagueParticipant record here

	return nil
}

func (s *seasonLeagueService) GetSeasonLeagueLeaderboard(leagueID uuid.UUID) ([]SeasonLeagueEntry, error) {
	// Get league
	_, err := s.leagueRepo.GetByID(leagueID)
	if err != nil {
		return nil, fmt.Errorf("season league not found: %w", err)
	}

	// In a real implementation, this would query actual participants and their performance
	// For now, we'll return a mock leaderboard
	mockLeaderboard := []SeasonLeagueEntry{
		{
			UserID:       uuid.New(),
			Username:     "ProPlayer1",
			TotalPoints:  1500,
			Rank:         1,
			TotalPrize:   1000.0,
			Participated: time.Now().AddDate(0, -1, 0),
		},
		{
			UserID:       uuid.New(),
			Username:     "GamerPro2",
			TotalPoints:  1200,
			Rank:         2,
			TotalPrize:   500.0,
			Participated: time.Now().AddDate(0, -1, 0),
		},
	}

	return mockLeaderboard, nil
}