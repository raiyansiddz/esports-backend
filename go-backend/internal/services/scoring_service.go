package services

import (
	"context"
	"encoding/json"
	"esports-fantasy-backend/internal/models"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ScoringService interface {
	UpdatePlayerStats(matchID, playerID uuid.UUID, stats *models.UpdateStatsRequest) error
	CalculatePlayerPoints(stats *models.PlayerMatchStats) float64
	RecalculateFantasyTeamScores(matchID uuid.UUID) error
}

type scoringService struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewScoringService(db *gorm.DB, rdb *redis.Client) ScoringService {
	return &scoringService{
		db:  db,
		rdb: rdb,
	}
}

// Scoring rules for eSports fantasy
const (
	PointsPerKill           = 10.0
	PointsPerKnockout       = 6.0
	PointsPerRevive         = 5.0
	PointsPerSurvivalMinute = 1.0
	PointsNotKnocked        = 15.0
	PointsMVP               = 20.0
	PointsTeamKillPenalty   = -2.0
	CaptainMultiplier       = 2.0
	ViceCaptainMultiplier   = 1.5
)

func (s *scoringService) UpdatePlayerStats(matchID, playerID uuid.UUID, stats *models.UpdateStatsRequest) error {
	// Start database transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find or create player match stats
	var playerStats models.PlayerMatchStats
	err := tx.Where("player_id = ? AND match_id = ?", playerID, matchID).First(&playerStats).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			playerStats = models.PlayerMatchStats{
				PlayerID: playerID,
				MatchID:  matchID,
			}
		} else {
			tx.Rollback()
			return fmt.Errorf("failed to get player stats: %w", err)
		}
	}

	// Update stats
	playerStats.Kills = stats.Kills
	playerStats.Revives = stats.Revives
	playerStats.Knockouts = stats.Knockouts
	playerStats.SurvivalTimeMinutes = stats.SurvivalTimeMinutes
	playerStats.IsMVP = stats.IsMVP
	playerStats.TeamKillPenalty = stats.TeamKillPenalty

	// Calculate points
	playerStats.TotalPoints = s.CalculatePlayerPoints(&playerStats)

	// Save player stats
	if err := tx.Save(&playerStats).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to save player stats: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Recalculate fantasy team scores asynchronously
	go func() {
		if err := s.RecalculateFantasyTeamScores(matchID); err != nil {
			log.Printf("Error recalculating fantasy team scores: %v", err)
		}
	}()

	log.Printf("✅ Updated stats for player %s in match %s: %.2f points", playerID, matchID, playerStats.TotalPoints)

	return nil
}

func (s *scoringService) CalculatePlayerPoints(stats *models.PlayerMatchStats) float64 {
	points := 0.0

	// Kill points
	points += float64(stats.Kills) * PointsPerKill

	// Knockout points
	points += float64(stats.Knockouts) * PointsPerKnockout

	// Revive points
	points += float64(stats.Revives) * PointsPerRevive

	// Survival points
	points += float64(stats.SurvivalTimeMinutes) * PointsPerSurvivalMinute

	// MVP bonus
	if stats.IsMVP {
		points += PointsMVP
	}

	// Team kill penalty
	points += float64(stats.TeamKillPenalty) * PointsTeamKillPenalty

	// Not knocked bonus (if survival time is significant, assume not knocked)
	if stats.SurvivalTimeMinutes > 20 { // Arbitrary threshold for "not knocked"
		points += PointsNotKnocked
	}

	return points
}

func (s *scoringService) RecalculateFantasyTeamScores(matchID uuid.UUID) error {
	// Get all fantasy teams for contests related to this match
	var fantasyTeams []models.FantasyTeam
	err := s.db.Joins("JOIN contests ON fantasy_teams.contest_id = contests.id").
		Where("contests.match_id = ?", matchID).
		Preload("Players.Player").
		Find(&fantasyTeams).Error
	if err != nil {
		return fmt.Errorf("failed to get fantasy teams: %w", err)
	}

	// Update scores for each fantasy team
	for _, team := range fantasyTeams {
		totalPoints := 0.0

		for _, fantasyPlayer := range team.Players {
			// Get player stats for this match
			var playerStats models.PlayerMatchStats
			err := s.db.Where("player_id = ? AND match_id = ?", fantasyPlayer.PlayerID, matchID).
				First(&playerStats).Error
			if err != nil {
				if err != gorm.ErrRecordNotFound {
					log.Printf("Error getting player stats for %s: %v", fantasyPlayer.PlayerID, err)
				}
				continue
			}

			points := playerStats.TotalPoints

			// Apply multipliers
			if fantasyPlayer.IsCaptain {
				points *= CaptainMultiplier
			} else if fantasyPlayer.IsViceCaptain {
				points *= ViceCaptainMultiplier
			}

			totalPoints += points
		}

		// Update fantasy team total points
		if err := s.db.Model(&team).Update("total_points", totalPoints).Error; err != nil {
			log.Printf("Error updating fantasy team %s points: %v", team.ID, err)
			continue
		}

		// Update leaderboard in Redis
		ctx := context.Background()
		leaderboardKey := fmt.Sprintf("leaderboard:%s", team.ContestID)
		if err := s.rdb.ZAdd(ctx, leaderboardKey, &redis.Z{
			Score:  totalPoints,
			Member: team.ID.String(),
		}).Err(); err != nil {
			log.Printf("Error updating Redis leaderboard: %v", err)
		}

		log.Printf("🏆 Updated fantasy team %s (%s): %.2f points", team.ID, team.TeamName, totalPoints)
	}

	// Publish leaderboard update event
	updateEvent := map[string]interface{}{
		"type":     "leaderboard_update",
		"match_id": matchID.String(),
	}
	eventData, _ := json.Marshal(updateEvent)
	
	ctx := context.Background()
	if err := s.rdb.Publish(ctx, "leaderboard_updates", eventData).Err(); err != nil {
		log.Printf("Error publishing leaderboard update: %v", err)
	}

	return nil
}