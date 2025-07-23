package services

import (
        "fmt"
        "log"
        "time"

        "esports-fantasy-backend/config"
        "esports-fantasy-backend/internal/models"
        "esports-fantasy-backend/internal/repository"

        "github.com/robfig/cron/v3"
)

type AutoContestService struct {
        cfg                *config.Config
        contestRepo        repository.ContestRepository
        matchRepo          repository.MatchRepository
        fantasyTeamRepo    repository.FantasyTeamRepository
        transactionRepo    repository.TransactionRepository
        userRepo           repository.UserRepository
        leaderboardService LeaderboardService
        cron               *cron.Cron
}

func NewAutoContestService(
        cfg *config.Config,
        contestRepo repository.ContestRepository,
        matchRepo repository.MatchRepository,
        fantasyTeamRepo repository.FantasyTeamRepository,
        transactionRepo repository.TransactionRepository,
        userRepo repository.UserRepository,
        leaderboardService *LeaderboardService,
) *AutoContestService {
        return &AutoContestService{
                cfg:                cfg,
                contestRepo:        contestRepo,
                matchRepo:          matchRepo,
                fantasyTeamRepo:    fantasyTeamRepo,
                transactionRepo:    transactionRepo,
                userRepo:           userRepo,
                leaderboardService: leaderboardService,
                cron:               cron.New(),
        }
}

func (s *AutoContestService) StartScheduler() error {
        if !s.cfg.AutoLockEnabled && !s.cfg.AutoPrizeDistributionEnabled {
                log.Println("📅 Auto Contest Management is disabled")
                return nil
        }

        // Schedule contest auto-lock (runs every minute)
        if s.cfg.AutoLockEnabled {
                s.cron.AddFunc("@every 1m", s.autoLockContests)
                log.Println("🔒 Auto Contest Lock scheduler started")
        }

        // Schedule prize distribution (runs every 5 minutes)
        if s.cfg.AutoPrizeDistributionEnabled {
                s.cron.AddFunc("@every 5m", s.autoPrizeDistribution)
                log.Println("🏆 Auto Prize Distribution scheduler started")
        }

        // Schedule match status updates (runs every 2 minutes)
        s.cron.AddFunc("@every 2m", s.autoUpdateMatchStatus)
        log.Println("⚽ Match Status Auto-Update scheduler started")

        // Schedule leaderboard refresh (runs every 30 seconds)
        s.cron.AddFunc("@every 30s", s.autoRefreshLeaderboards)
        log.Println("📊 Leaderboard Auto-Refresh scheduler started")

        s.cron.Start()
        log.Println("🚀 Auto Contest Management Service started successfully!")
        return nil
}

func (s *AutoContestService) StopScheduler() {
        if s.cron != nil {
                s.cron.Stop()
                log.Println("⏹️ Auto Contest Management Service stopped")
        }
}

func (s *AutoContestService) autoLockContests() {
        log.Println("🔍 Checking contests for auto-lock...")

        // Get all open contests
        contests, err := s.contestRepo.GetContestsByStatus("OPEN")
        if err != nil {
                log.Printf("❌ Error fetching open contests: %v", err)
                return
        }

        lockTime := time.Duration(s.cfg.ContestLockMinutesBeforeMatch) * time.Minute
        lockedCount := 0

        for _, contest := range contests {
                // Get match details
                match, err := s.matchRepo.GetByID(contest.MatchID.String())
                if err != nil {
                        log.Printf("❌ Error fetching match %s: %v", contest.MatchID, err)
                        continue
                }

                // Check if contest should be locked
                timeUntilMatch := time.Until(match.StartTime)
                if timeUntilMatch <= lockTime {
                        // Lock the contest
                        contest.Status = "LOCKED"
                        now := time.Now()
                        contest.LockedAt = &now
                        contest.UpdatedAt = time.Now()

                        if err := s.contestRepo.Update(contest); err != nil {
                                log.Printf("❌ Error locking contest %s: %v", contest.ID, err)
                                continue
                        }

                        lockedCount++
                        log.Printf("🔒 Contest locked: %s (%s) - Match starts in %.0f minutes", 
                                contest.Name, contest.ID[:8], timeUntilMatch.Minutes())

                        // Notify leaderboard service
                        if s.leaderboardService != nil {
                                s.leaderboardService.InitializeContestLeaderboard(contest.ID)
                        }
                }
        }

        if lockedCount > 0 {
                log.Printf("✅ Auto-locked %d contests", lockedCount)
        }
}

func (s *AutoContestService) autoPrizeDistribution() {
        log.Println("🏆 Checking contests for prize distribution...")

        // Get all completed contests that haven't distributed prizes
        contests, err := s.contestRepo.GetContestsByStatus("COMPLETED")
        if err != nil {
                log.Printf("❌ Error fetching completed contests: %v", err)
                return
        }

        distributedCount := 0

        for _, contest := range contests {
                // Check if prizes already distributed
                if contest.PrizesDistributed {
                        continue
                }

                // Get match to ensure it's really completed
                match, err := s.matchRepo.GetByID(contest.MatchID.String())
                if err != nil {
                        log.Printf("❌ Error fetching match %s: %v", contest.MatchID, err)
                        continue
                }

                if match.Status != "COMPLETED" {
                        continue
                }

                // Distribute prizes
                if err := s.distributePrizes(contest); err != nil {
                        log.Printf("❌ Error distributing prizes for contest %s: %v", contest.ID, err)
                        continue
                }

                distributedCount++
                log.Printf("💰 Prizes distributed for contest: %s (%s)", contest.Name, contest.ID[:8])
        }

        if distributedCount > 0 {
                log.Printf("✅ Distributed prizes for %d contests", distributedCount)
        }
}

func (s *AutoContestService) distributePrizes(contest *models.Contest) error {
        // Get final leaderboard
        leaderboard, err := s.leaderboardService.GetContestLeaderboard(contest.ID, 100)
        if err != nil {
                return fmt.Errorf("failed to get leaderboard: %w", err)
        }

        if len(leaderboard) == 0 {
                log.Printf("⚠️ No participants in contest %s", contest.ID)
                contest.PrizesDistributed = true
                return s.contestRepo.Update(contest)
        }

        // Parse prize distribution
        prizeDistribution := make(map[int]float64)
        // In a real implementation, you'd parse contest.PrizePool JSON
        // For now, simple distribution based on total prize pool
        totalPrizePool := float64(contest.EntryFee) * float64(len(leaderboard)) * 0.9 // 90% of entry fees

        // Simple prize distribution: Winner gets 50%, 2nd gets 30%, 3rd gets 20%
        if len(leaderboard) >= 1 {
                prizeDistribution[1] = totalPrizePool * 0.5
        }
        if len(leaderboard) >= 2 {
                prizeDistribution[2] = totalPrizePool * 0.3
        }
        if len(leaderboard) >= 3 {
                prizeDistribution[3] = totalPrizePool * 0.2
        }

        // Distribute prizes
        for rank, amount := range prizeDistribution {
                if rank <= len(leaderboard) {
                        entry := leaderboard[rank-1]
                        
                        // Create prize transaction
                        transaction := &models.Transaction{
                                UserID:         entry.UserID,
                                Amount:         amount,
                                Type:           "winnings",
                                Status:         "completed",
                                RelatedEntityID: &contest.ID,
                                CreatedAt:      time.Now(),
                                UpdatedAt:      time.Now(),
                        }

                        if err := s.transactionRepo.Create(transaction); err != nil {
                                log.Printf("❌ Error creating prize transaction: %v", err)
                                continue
                        }

                        // Update user wallet
                        user, err := s.userRepo.GetByID(entry.UserID)
                        if err != nil {
                                log.Printf("❌ Error fetching user %s: %v", entry.UserID, err)
                                continue
                        }

                        user.WalletBalance += amount
                        user.UpdatedAt = time.Now()

                        if err := s.userRepo.Update(user); err != nil {
                                log.Printf("❌ Error updating user wallet: %v", err)
                                continue
                        }

                        log.Printf("💰 Prize distributed: ₹%.2f to user %s (Rank %d)", 
                                amount, user.PhoneNumber, rank)
                }
        }

        // Mark contest as prizes distributed
        contest.PrizesDistributed = true
        contest.UpdatedAt = time.Now()
        return s.contestRepo.Update(contest)
}

func (s *AutoContestService) autoUpdateMatchStatus() {
        log.Println("⚽ Checking match status updates...")

        // Get all live and upcoming matches
        liveMatches, _ := s.matchRepo.GetMatchesByStatus("LIVE")
        upcomingMatches, _ := s.matchRepo.GetMatchesByStatus("UPCOMING")

        allMatches := append(liveMatches, upcomingMatches...)
        updatedCount := 0

        for _, match := range allMatches {
                updated := false

                // Check if upcoming match should be live
                if match.Status == "UPCOMING" && time.Now().After(match.StartTime) {
                        match.Status = "LIVE"
                        updated = true
                        log.Printf("🔴 Match is now LIVE: %s", match.Name)
                }

                // Check if live match should be completed (2 hours after start)
                if match.Status == "LIVE" && time.Now().After(match.StartTime.Add(2*time.Hour)) {
                        match.Status = "COMPLETED"
                        updated = true
                        log.Printf("✅ Match completed: %s", match.Name)

                        // Update associated contests
                        s.updateContestsForCompletedMatch(match.ID)
                }

                if updated {
                        match.UpdatedAt = time.Now()
                        if err := s.matchRepo.Update(match); err != nil {
                                log.Printf("❌ Error updating match status: %v", err)
                        } else {
                                updatedCount++
                        }
                }
        }

        if updatedCount > 0 {
                log.Printf("✅ Updated status for %d matches", updatedCount)
        }
}

func (s *AutoContestService) updateContestsForCompletedMatch(matchID string) {
        contests, err := s.contestRepo.GetContestsByMatchID(matchID)
        if err != nil {
                log.Printf("❌ Error fetching contests for match %s: %v", matchID, err)
                return
        }

        for _, contest := range contests {
                if contest.Status == "LOCKED" {
                        contest.Status = "COMPLETED"
                        contest.UpdatedAt = time.Now()

                        if err := s.contestRepo.Update(contest); err != nil {
                                log.Printf("❌ Error updating contest status: %v", err)
                        } else {
                                log.Printf("🏁 Contest completed: %s", contest.Name)
                        }
                }
        }
}

func (s *AutoContestService) autoRefreshLeaderboards() {
        // Get all active contests (locked or live)
        lockedContests, _ := s.contestRepo.GetContestsByStatus("LOCKED")
        
        for _, contest := range lockedContests {
                // Refresh leaderboard
                if s.leaderboardService != nil {
                        s.leaderboardService.RefreshContestLeaderboard(contest.ID)
                }
        }
}

// Manual methods for immediate actions
func (s *AutoContestService) ForceDistributePrizes(contestID string) error {
        contest, err := s.contestRepo.GetByID(contestID)
        if err != nil {
                return fmt.Errorf("contest not found: %w", err)
        }

        return s.distributePrizes(contest)
}

func (s *AutoContestService) ForceLockContest(contestID string) error {
        contest, err := s.contestRepo.GetByID(contestID)
        if err != nil {
                return fmt.Errorf("contest not found: %w", err)
        }

        contest.Status = "LOCKED"
        contest.LockedAt = time.Now()
        contest.UpdatedAt = time.Now()

        return s.contestRepo.Update(contest)
}

func (s *AutoContestService) GetSchedulerStatus() map[string]interface{} {
        return map[string]interface{}{
                "auto_lock_enabled":              s.cfg.AutoLockEnabled,
                "auto_prize_distribution_enabled": s.cfg.AutoPrizeDistributionEnabled,
                "contest_lock_minutes_before_match": s.cfg.ContestLockMinutesBeforeMatch,
                "scheduler_running":              s.cron != nil,
                "next_runs":                      s.cron.Entries(),
        }
}