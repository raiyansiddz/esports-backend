package services

import (
        "context"
        "database/sql"
        "fmt"
        "log"
        "time"

        "esports-fantasy-backend/config"
        "esports-fantasy-backend/internal/models"
        "esports-fantasy-backend/internal/repository"

        "github.com/go-redis/redis/v8"
        "github.com/google/uuid"
        "gorm.io/gorm"
)

var ctx = context.Background()

type AnalyticsService struct {
        cfg             *config.Config
        db              *gorm.DB
        rdb             *redis.Client
        userRepo        repository.UserRepository
        contestRepo     repository.ContestRepository
        transactionRepo repository.TransactionRepository
        matchRepo       repository.MatchRepository
}

type DashboardStats struct {
        TotalUsers           int64   `json:"total_users"`
        ActiveUsers          int64   `json:"active_users"`
        TotalContests        int64   `json:"total_contests"`
        ActiveContests       int64   `json:"active_contests"`
        TotalMatches         int64   `json:"total_matches"`
        LiveMatches          int64   `json:"live_matches"`
        TotalRevenue         float64 `json:"total_revenue"`
        TodayRevenue         float64 `json:"today_revenue"`
        TotalTransactions    int64   `json:"total_transactions"`
        SuccessfulPayments   int64   `json:"successful_payments"`
        PendingPayments      int64   `json:"pending_payments"`
        FailedPayments       int64   `json:"failed_payments"`
        WalletBalance        float64 `json:"total_wallet_balance"`
        AvgContestEntry      float64 `json:"avg_contest_entry"`
        PopularGameTypes     []GameTypeStats `json:"popular_game_types"`
        RevenueGrowth        []RevenueGrowth `json:"revenue_growth"`
        UserGrowth           []UserGrowth    `json:"user_growth"`
        ContestParticipation []ContestParticipation `json:"contest_participation"`
}

type GameTypeStats struct {
        GameType string `json:"game_type"`
        Count    int64  `json:"count"`
        Revenue  float64 `json:"revenue"`
}

type RevenueGrowth struct {
        Date    string  `json:"date"`
        Revenue float64 `json:"revenue"`
}

type UserGrowth struct {
        Date  string `json:"date"`
        Count int64  `json:"count"`
}

type ContestParticipation struct {
        ContestName   string  `json:"contest_name"`
        Participants  int64   `json:"participants"`
        PrizePool     float64 `json:"prize_pool"`
        Status        string  `json:"status"`
}

type UserAnalytics struct {
        UserID              uuid.UUID             `json:"user_id"`
        PhoneNumber         string                `json:"phone_number"`
        TotalContests       int64                 `json:"total_contests"`
        WonContests         int64                 `json:"won_contests"`
        TotalWinnings       float64               `json:"total_winnings"`
        TotalSpent          float64               `json:"total_spent"`
        WinRate             float64               `json:"win_rate"`
        FavoriteGameType    string                `json:"favorite_game_type"`
        AvgContestEntry     float64               `json:"avg_contest_entry"`
        WalletBalance       float64               `json:"wallet_balance"`
        LastActivity        time.Time             `json:"last_activity"`
        ContestHistory      []ContestHistoryItem  `json:"contest_history"`
        TransactionHistory  []TransactionSummary  `json:"transaction_history"`
}

type ContestHistoryItem struct {
        ContestName string    `json:"contest_name"`
        Rank        int       `json:"rank"`
        Points      float64   `json:"points"`
        Prize       float64   `json:"prize"`
        Date        time.Time `json:"date"`
}

type TransactionSummary struct {
        Type        string    `json:"type"`
        Amount      float64   `json:"amount"`
        Status      string    `json:"status"`
        Description string    `json:"description"`
        Date        time.Time `json:"date"`
}

type MatchAnalytics struct {
        MatchID           uuid.UUID              `json:"match_id"`
        MatchName         string                 `json:"match_name"`
        TotalContests     int64                  `json:"total_contests"`
        TotalParticipants int64                  `json:"total_participants"`
        TotalPrizePool    float64                `json:"total_prize_pool"`
        AvgPointsScored   float64                `json:"avg_points_scored"`
        TopPerformers     []PlayerPerformance    `json:"top_performers"`
        ContestBreakdown  []ContestAnalysis      `json:"contest_breakdown"`
}

type PlayerPerformance struct {
        PlayerName  string  `json:"player_name"`
        Points      float64 `json:"points"`
        Kills       int     `json:"kills"`
        Knockouts   int     `json:"knockouts"`
        IsMVP       bool    `json:"is_mvp"`
}

type ContestAnalysis struct {
        ContestName   string  `json:"contest_name"`
        Participants  int64   `json:"participants"`
        EntryFee      float64 `json:"entry_fee"`
        PrizePool     float64 `json:"prize_pool"`
        WinnerPoints  float64 `json:"winner_points"`
}

func NewAnalyticsService(cfg *config.Config, db *gorm.DB, rdb *redis.Client, userRepo repository.UserRepository, contestRepo repository.ContestRepository, transactionRepo repository.TransactionRepository, matchRepo repository.MatchRepository) *AnalyticsService {
        return &AnalyticsService{
                cfg:             cfg,
                db:              db,
                rdb:             rdb,
                userRepo:        userRepo,
                contestRepo:     contestRepo,
                transactionRepo: transactionRepo,
                matchRepo:       matchRepo,
        }
}

func (s *AnalyticsService) GetDashboardStats() (*DashboardStats, error) {
        stats := &DashboardStats{}

        // User statistics
        s.db.Model(&models.User{}).Count(&stats.TotalUsers)
        s.db.Model(&models.User{}).Where("updated_at > ?", time.Now().AddDate(0, 0, -7)).Count(&stats.ActiveUsers)

        // Contest statistics  
        s.db.Model(&models.Contest{}).Count(&stats.TotalContests)
        s.db.Model(&models.Contest{}).Where("status IN ?", []string{"OPEN", "LOCKED"}).Count(&stats.ActiveContests)

        // Match statistics
        s.db.Model(&models.Match{}).Count(&stats.TotalMatches)
        s.db.Model(&models.Match{}).Where("status = ?", "LIVE").Count(&stats.LiveMatches)

        // Transaction statistics
        s.db.Model(&models.Transaction{}).Count(&stats.TotalTransactions)
        s.db.Model(&models.Transaction{}).Where("status = ?", "SUCCESS").Count(&stats.SuccessfulPayments)
        s.db.Model(&models.Transaction{}).Where("status = ?", "PENDING").Count(&stats.PendingPayments)
        s.db.Model(&models.Transaction{}).Where("status = ?", "FAILED").Count(&stats.FailedPayments)

        // Revenue statistics
        var totalRevenue, todayRevenue, walletBalance sql.NullFloat64
        s.db.Model(&models.Transaction{}).
                Where("status = ? AND type IN ?", "SUCCESS", []string{"DEPOSIT", "CONTEST_ENTRY"}).
                Select("COALESCE(SUM(amount), 0)").Scan(&totalRevenue)
        stats.TotalRevenue = totalRevenue.Float64

        s.db.Model(&models.Transaction{}).
                Where("status = ? AND type IN ? AND DATE(created_at) = DATE(?)", 
                        "SUCCESS", []string{"DEPOSIT", "CONTEST_ENTRY"}, time.Now()).
                Select("COALESCE(SUM(amount), 0)").Scan(&todayRevenue)
        stats.TodayRevenue = todayRevenue.Float64

        s.db.Model(&models.User{}).
                Select("COALESCE(SUM(wallet_balance), 0)").Scan(&walletBalance)
        stats.WalletBalance = walletBalance.Float64

        // Average contest entry
        var avgEntry sql.NullFloat64
        s.db.Model(&models.Contest{}).
                Select("COALESCE(AVG(entry_fee), 0)").Scan(&avgEntry)
        stats.AvgContestEntry = avgEntry.Float64

        // Popular game types
        stats.PopularGameTypes = s.getPopularGameTypes()

        // Revenue growth (last 7 days)
        stats.RevenueGrowth = s.getRevenueGrowth(7)

        // User growth (last 7 days)
        stats.UserGrowth = s.getUserGrowth(7)

        // Contest participation
        stats.ContestParticipation = s.getContestParticipation()

        return stats, nil
}

func (s *AnalyticsService) GetUserAnalytics(userID string) (*UserAnalytics, error) {
        user, err := s.userRepo.GetByID(userID)
        if err != nil {
                return nil, fmt.Errorf("user not found: %w", err)
        }

        analytics := &UserAnalytics{
                UserID:         user.ID,
                PhoneNumber:    user.PhoneNumber,
                WalletBalance:  user.WalletBalance,
                LastActivity:   user.UpdatedAt,
        }

        // Get contest statistics
        var contestCount int64
        s.db.Model(&models.FantasyTeam{}).Where("user_id = ?", userID).Count(&contestCount)
        analytics.TotalContests = contestCount

        // Get winnings and spending
        var totalWinnings, totalSpent sql.NullFloat64
        s.db.Model(&models.Transaction{}).
                Where("user_id = ? AND type = ? AND status = ?", userID, "PRIZE", "SUCCESS").
                Select("COALESCE(SUM(amount), 0)").Scan(&totalWinnings)
        analytics.TotalWinnings = totalWinnings.Float64

        s.db.Model(&models.Transaction{}).
                Where("user_id = ? AND type IN ? AND status = ?", 
                        userID, []string{"DEPOSIT", "CONTEST_ENTRY"}, "SUCCESS").
                Select("COALESCE(SUM(amount), 0)").Scan(&totalSpent)
        analytics.TotalSpent = totalSpent.Float64

        // Calculate win rate and other metrics
        if analytics.TotalContests > 0 {
                analytics.WinRate = float64(analytics.WonContests) / float64(analytics.TotalContests) * 100
                analytics.AvgContestEntry = analytics.TotalSpent / float64(analytics.TotalContests)
        }

        // Get contest history
        analytics.ContestHistory = s.getUserContestHistory(userID)

        // Get transaction history
        analytics.TransactionHistory = s.getUserTransactionHistory(userID)

        return analytics, nil
}

func (s *AnalyticsService) GetMatchAnalytics(matchID string) (*MatchAnalytics, error) {
        match, err := s.matchRepo.GetByID(matchID)
        if err != nil {
                return nil, fmt.Errorf("match not found: %w", err)
        }

        analytics := &MatchAnalytics{
                MatchID:   match.ID,
                MatchName: match.Name,
        }

        // Get contest statistics for this match
        var contestCount int64
        s.db.Model(&models.Contest{}).Where("match_id = ?", matchID).Count(&contestCount)
        analytics.TotalContests = contestCount

        // Get total participants and prize pool
        var participants int64
        var prizePool sql.NullFloat64
        
        s.db.Table("fantasy_teams").
                Joins("JOIN contests ON fantasy_teams.contest_id = contests.id").
                Where("contests.match_id = ?", matchID).
                Count(&participants)
        analytics.TotalParticipants = participants

        s.db.Model(&models.Contest{}).
                Where("match_id = ?", matchID).
                Select("COALESCE(SUM(entry_fee * max_entries), 0)").Scan(&prizePool)
        analytics.TotalPrizePool = prizePool.Float64

        // Get contest breakdown
        analytics.ContestBreakdown = s.getMatchContestBreakdown(matchID)

        return analytics, nil
}

// Helper methods
func (s *AnalyticsService) getPopularGameTypes() []GameTypeStats {
        var results []GameTypeStats
        
        s.db.Table("tournaments").
                Select("game_type, COUNT(*) as count, COALESCE(SUM(entry_fee), 0) as revenue").
                Joins("LEFT JOIN contests ON tournaments.id = contests.tournament_id").
                Group("game_type").
                Order("count DESC").
                Limit(5).
                Scan(&results)
        
        return results
}

func (s *AnalyticsService) getRevenueGrowth(days int) []RevenueGrowth {
        var results []RevenueGrowth
        
        s.db.Table("transactions").
                Select("DATE(created_at) as date, COALESCE(SUM(amount), 0) as revenue").
                Where("status = ? AND type IN ? AND created_at >= ?", 
                        "SUCCESS", []string{"DEPOSIT", "CONTEST_ENTRY"}, 
                        time.Now().AddDate(0, 0, -days)).
                Group("DATE(created_at)").
                Order("date DESC").
                Scan(&results)
        
        return results
}

func (s *AnalyticsService) getUserGrowth(days int) []UserGrowth {
        var results []UserGrowth
        
        s.db.Table("users").
                Select("DATE(created_at) as date, COUNT(*) as count").
                Where("created_at >= ?", time.Now().AddDate(0, 0, -days)).
                Group("DATE(created_at)").
                Order("date DESC").
                Scan(&results)
        
        return results
}

func (s *AnalyticsService) getContestParticipation() []ContestParticipation {
        var results []ContestParticipation
        
        s.db.Table("contests").
                Select("name as contest_name, max_entries as participants, entry_fee * max_entries as prize_pool, status").
                Where("status IN ?", []string{"OPEN", "LOCKED", "COMPLETED"}).
                Order("created_at DESC").
                Limit(10).
                Scan(&results)
        
        return results
}

func (s *AnalyticsService) getUserContestHistory(userID string) []ContestHistoryItem {
        var results []ContestHistoryItem
        
        // This would require more complex joins to get actual contest results
        // For now, return empty slice
        return results
}

func (s *AnalyticsService) getUserTransactionHistory(userID string) []TransactionSummary {
        var results []TransactionSummary
        
        s.db.Table("transactions").
                Select("type, amount, status, description, created_at as date").
                Where("user_id = ?", userID).
                Order("created_at DESC").
                Limit(20).
                Scan(&results)
        
        return results
}

func (s *AnalyticsService) getMatchContestBreakdown(matchID string) []ContestAnalysis {
        var results []ContestAnalysis
        
        s.db.Table("contests").
                Select("name as contest_name, max_entries as participants, entry_fee, entry_fee * max_entries as prize_pool, 0 as winner_points").
                Where("match_id = ?", matchID).
                Scan(&results)
        
        return results
}

// Real-time analytics caching
func (s *AnalyticsService) CacheStats() error {
        if !s.cfg.AnalyticsEnabled {
                return nil
        }

        stats, err := s.GetDashboardStats()
        if err != nil {
                return err
        }

        // Cache for 5 minutes
        return s.rdb.Set(ctx, "analytics:dashboard", stats, 5*time.Minute).Err()
}

func (s *AnalyticsService) GetCachedStats() (*DashboardStats, error) {
        var stats DashboardStats
        
        result := s.rdb.Get(ctx, "analytics:dashboard")
        if result.Err() == redis.Nil {
                // Cache miss, generate fresh stats
                return s.GetDashboardStats()
        }
        
        if err := result.Scan(&stats); err != nil {
                return s.GetDashboardStats()
        }
        
        return &stats, nil
}

func (s *AnalyticsService) StartAnalyticsCaching() {
        if !s.cfg.AnalyticsEnabled {
                return
        }

        // Cache stats every 5 minutes
        ticker := time.NewTicker(5 * time.Minute)
        go func() {
                for {
                        select {
                        case <-ticker.C:
                                if err := s.CacheStats(); err != nil {
                                        log.Printf("❌ Error caching analytics: %v", err)
                                } else {
                                        log.Println("📊 Analytics stats cached successfully")
                                }
                        }
                }
        }()

        log.Println("📊 Analytics caching service started")
}