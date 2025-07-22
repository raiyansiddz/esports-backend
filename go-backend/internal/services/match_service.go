package services

import (
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type MatchService interface {
	CreateMatch(match *models.Match) error
	GetMatches() ([]models.Match, error)
	GetMatchByID(id uuid.UUID) (*models.Match, error)
	UpdateMatch(match *models.Match) error
	UpdateMatchStatus(id uuid.UUID, status string) error
	GetUpcomingMatches() ([]models.Match, error)
	LockExpiredMatches() error
}

type matchService struct {
	matchRepo repository.MatchRepository
}

func NewMatchService(matchRepo repository.MatchRepository) MatchService {
	return &matchService{
		matchRepo: matchRepo,
	}
}

func (s *matchService) CreateMatch(match *models.Match) error {
	if err := s.matchRepo.CreateMatch(match); err != nil {
		return fmt.Errorf("failed to create match: %w", err)
	}
	return nil
}

func (s *matchService) GetMatches() ([]models.Match, error) {
	matches, err := s.matchRepo.GetMatches()
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}
	return matches, nil
}

func (s *matchService) GetMatchByID(id uuid.UUID) (*models.Match, error) {
	match, err := s.matchRepo.GetMatchByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get match: %w", err)
	}
	return match, nil
}

func (s *matchService) UpdateMatch(match *models.Match) error {
	if err := s.matchRepo.UpdateMatch(match); err != nil {
		return fmt.Errorf("failed to update match: %w", err)
	}
	return nil
}

func (s *matchService) UpdateMatchStatus(id uuid.UUID, status string) error {
	if err := s.matchRepo.UpdateMatchStatus(id, status); err != nil {
		return fmt.Errorf("failed to update match status: %w", err)
	}
	return nil
}

func (s *matchService) GetUpcomingMatches() ([]models.Match, error) {
	matches, err := s.matchRepo.GetUpcomingMatches()
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming matches: %w", err)
	}
	return matches, nil
}

func (s *matchService) LockExpiredMatches() error {
	matches, err := s.matchRepo.GetMatchesNeedingLock()
	if err != nil {
		return fmt.Errorf("failed to get matches needing lock: %w", err)
	}

	for _, match := range matches {
		if err := s.matchRepo.UpdateMatchStatus(match.ID, "locked"); err != nil {
			return fmt.Errorf("failed to lock match %s: %w", match.ID, err)
		}
	}

	return nil
}