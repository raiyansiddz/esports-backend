package services

import (
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type ContestService interface {
	CreateContest(contest *models.Contest) error
	GetContestsByMatch(matchID uuid.UUID) ([]models.Contest, error)
	GetContestByID(id uuid.UUID) (*models.Contest, error)
	JoinContest(userID, contestID uuid.UUID) error
}

type contestService struct {
	contestRepo repository.ContestRepository
	userRepo    repository.UserRepository
}

func NewContestService(contestRepo repository.ContestRepository) ContestService {
	return &contestService{
		contestRepo: contestRepo,
	}
}

func (s *contestService) CreateContest(contest *models.Contest) error {
	if err := s.contestRepo.CreateContest(contest); err != nil {
		return fmt.Errorf("failed to create contest: %w", err)
	}
	return nil
}

func (s *contestService) GetContestsByMatch(matchID uuid.UUID) ([]models.Contest, error) {
	contests, err := s.contestRepo.GetContestsByMatchID(matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contests: %w", err)
	}
	return contests, nil
}

func (s *contestService) GetContestByID(id uuid.UUID) (*models.Contest, error) {
	contest, err := s.contestRepo.GetContestByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get contest: %w", err)
	}
	return contest, nil
}

func (s *contestService) JoinContest(userID, contestID uuid.UUID) error {
	// Get contest details
	contest, err := s.contestRepo.GetContestByID(contestID)
	if err != nil {
		return fmt.Errorf("contest not found: %w", err)
	}

	// Check if contest is full
	if contest.CurrentEntries >= contest.MaxEntries {
		return fmt.Errorf("contest is full")
	}

	// Increment contest entries
	if err := s.contestRepo.IncrementEntries(contestID); err != nil {
		return fmt.Errorf("failed to join contest: %w", err)
	}

	return nil
}