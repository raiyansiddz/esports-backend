package services

import (
	"context"
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/repository"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type LeaderboardService interface {
	GetLeaderboard(contestID uuid.UUID, limit int) ([]LeaderboardEntry, error)
	UpdateTeamScore(contestID, teamID uuid.UUID, points float64) error
	GetTeamRank(contestID, teamID uuid.UUID) (int, error)
}

type LeaderboardEntry struct {
	TeamID     uuid.UUID `json:"team_id"`
	TeamName   string    `json:"team_name"`
	UserName   string    `json:"user_name"`
	Points     float64   `json:"points"`
	Rank       int       `json:"rank"`
}

type leaderboardService struct {
	rdb             *redis.Client
	fantasyTeamRepo repository.FantasyTeamRepository
}

func NewLeaderboardService(rdb *redis.Client, fantasyTeamRepo repository.FantasyTeamRepository) LeaderboardService {
	return &leaderboardService{
		rdb:             rdb,
		fantasyTeamRepo: fantasyTeamRepo,
	}
}

func (s *leaderboardService) GetLeaderboard(contestID uuid.UUID, limit int) ([]LeaderboardEntry, error) {
	ctx := context.Background()
	leaderboardKey := fmt.Sprintf("leaderboard:%s", contestID)

	// Get top teams from Redis sorted set
	results, err := s.rdb.ZRevRangeWithScores(ctx, leaderboardKey, 0, int64(limit-1)).Result()
	if err != nil {
		// Fallback to database if Redis fails
		return s.getLeaderboardFromDB(contestID, limit)
	}

	if len(results) == 0 {
		// No Redis data, fall back to database
		return s.getLeaderboardFromDB(contestID, limit)
	}

	var entries []LeaderboardEntry
	for i, result := range results {
		teamIDStr, ok := result.Member.(string)
		if !ok {
			continue
		}

		teamID, err := uuid.Parse(teamIDStr)
		if err != nil {
			continue
		}

		// Get team details from database
		team, err := s.fantasyTeamRepo.GetTeamByID(teamID)
		if err != nil {
			continue
		}

		entries = append(entries, LeaderboardEntry{
			TeamID:   teamID,
			TeamName: team.TeamName,
			UserName: team.User.Name,
			Points:   result.Score,
			Rank:     i + 1,
		})
	}

	return entries, nil
}

func (s *leaderboardService) getLeaderboardFromDB(contestID uuid.UUID, limit int) ([]LeaderboardEntry, error) {
	teams, err := s.fantasyTeamRepo.GetLeaderboard(contestID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard from database: %w", err)
	}

	var entries []LeaderboardEntry
	for i, team := range teams {
		entries = append(entries, LeaderboardEntry{
			TeamID:   team.ID,
			TeamName: team.TeamName,
			UserName: team.User.Name,
			Points:   team.TotalPoints,
			Rank:     i + 1,
		})
	}

	return entries, nil
}

func (s *leaderboardService) UpdateTeamScore(contestID, teamID uuid.UUID, points float64) error {
	ctx := context.Background()
	leaderboardKey := fmt.Sprintf("leaderboard:%s", contestID)

	// Update score in Redis sorted set
	return s.rdb.ZAdd(ctx, leaderboardKey, &redis.Z{
		Score:  points,
		Member: teamID.String(),
	}).Err()
}

func (s *leaderboardService) GetTeamRank(contestID, teamID uuid.UUID) (int, error) {
	ctx := context.Background()
	leaderboardKey := fmt.Sprintf("leaderboard:%s", contestID)

	// Get rank from Redis (0-indexed, so add 1)
	rank, err := s.rdb.ZRevRank(ctx, leaderboardKey, teamID.String()).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("team not found in leaderboard")
		}
		return 0, err
	}

	return int(rank) + 1, nil
}