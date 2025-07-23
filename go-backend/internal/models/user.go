package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PhoneNumber     string         `json:"phone_number" gorm:"unique;not null"`
	Name            string         `json:"name"`
	Username        string         `json:"username" gorm:"unique;not null"`
	ProfileImage    string         `json:"profile_image" gorm:"type:text"` // Base64 encoded image
	WalletBalance   float64        `json:"wallet_balance" gorm:"default:0.00"`
	IsAdmin         bool           `json:"is_admin" gorm:"default:false"`
	IsVerified      bool           `json:"is_verified" gorm:"default:false"`
	ReferralCode    string         `json:"referral_code" gorm:"unique"`
	ReferredBy      *uuid.UUID     `json:"referred_by"`
	ReferralBonus   float64        `json:"referral_bonus" gorm:"default:0.00"`
	TierLevel       string         `json:"tier_level" gorm:"default:bronze"` // bronze, silver, gold, diamond, vip
	TotalPoints     int64          `json:"total_points" gorm:"default:0"`
	ConsecutiveWins int            `json:"consecutive_wins" gorm:"default:0"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

type OTP struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PhoneNumber string    `json:"phone_number" gorm:"not null"`
	Code        string    `json:"code" gorm:"not null"`
	ExpiresAt   time.Time `json:"expires_at" gorm:"not null"`
	Used        bool      `json:"used" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
}

type Tournament struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"not null"`
	GameID    uuid.UUID `json:"game_id"` // Reference to Game
	GameType  string    `json:"game_type" gorm:"not null"` // BGMI, Valorant, etc. (kept for backward compatibility)
	StartDate time.Time `json:"start_date" gorm:"not null"`
	EndDate   time.Time `json:"end_date" gorm:"not null"`
	Status    string    `json:"status" gorm:"default:upcoming"` // upcoming, live, completed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ESportsTeam struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name     string    `json:"name" gorm:"unique;not null"`
	LogoURL  string    `json:"logo_url"`
	Players  []Player  `json:"players" gorm:"foreignKey:ESportsTeamID"`
}

type Player struct {
	ID             uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ESportsTeamID  uuid.UUID   `json:"esports_team_id"`
	ESportsTeam    ESportsTeam `json:"esports_team" gorm:"foreignKey:ESportsTeamID"`
	GameID         uuid.UUID   `json:"game_id"` // Players can play different games
	Name           string      `json:"name" gorm:"not null"`
	Role           string      `json:"role"` // rusher, assaulter, support, sniper
	CreditValue    float64     `json:"credit_value" gorm:"default:8.0"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type Match struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TournamentID uuid.UUID  `json:"tournament_id"`
	Tournament   Tournament `json:"tournament" gorm:"foreignKey:TournamentID"`
	Name         string     `json:"name"`
	MapName      string     `json:"map_name"`
	StartTime    time.Time  `json:"start_time" gorm:"not null"`
	Status       string     `json:"status" gorm:"default:upcoming"` // upcoming, locked, live, completed, cancelled
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Contest struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MatchID           uuid.UUID `json:"match_id"`
	Match             Match     `json:"match" gorm:"foreignKey:MatchID"`
	Name              string    `json:"name" gorm:"not null"`
	EntryFee          float64   `json:"entry_fee" gorm:"not null"`
	PrizePool         string    `json:"prize_pool" gorm:"type:jsonb"` // JSON structure for prize distribution
	MaxEntries        int       `json:"max_entries" gorm:"not null"`
	CurrentEntries    int       `json:"current_entries" gorm:"default:0"`
	IsPrivate         bool      `json:"is_private" gorm:"default:false"`
	InviteCode        string    `json:"invite_code" gorm:"unique"`
	Status            string    `json:"status" gorm:"default:open"` // open, locked, completed, cancelled
	LockedAt          *time.Time `json:"locked_at"`
	PrizesDistributed bool      `json:"prizes_distributed" gorm:"default:false"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type FantasyTeam struct {
	ID            uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        uuid.UUID              `json:"user_id"`
	User          User                   `json:"user" gorm:"foreignKey:UserID"`
	ContestID     uuid.UUID              `json:"contest_id"`
	Contest       Contest                `json:"contest" gorm:"foreignKey:ContestID"`
	TeamName      string                 `json:"team_name"`
	Players       []FantasyTeamPlayer    `json:"players" gorm:"foreignKey:FantasyTeamID"`
	TotalPoints   float64                `json:"total_points" gorm:"default:0.00"`
	Rank          *int                   `json:"rank"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type FantasyTeamPlayer struct {
	FantasyTeamID  uuid.UUID   `json:"fantasy_team_id" gorm:"primaryKey"`
	FantasyTeam    FantasyTeam `json:"fantasy_team" gorm:"foreignKey:FantasyTeamID"`
	PlayerID       uuid.UUID   `json:"player_id" gorm:"primaryKey"`
	Player         Player      `json:"player" gorm:"foreignKey:PlayerID"`
	IsCaptain      bool        `json:"is_captain" gorm:"default:false"`
	IsViceCaptain  bool        `json:"is_vice_captain" gorm:"default:false"`
}

type PlayerMatchStats struct {
	ID                   uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	PlayerID            uuid.UUID `json:"player_id"`
	Player              Player    `json:"player" gorm:"foreignKey:PlayerID"`
	MatchID             uuid.UUID `json:"match_id"`
	Match               Match     `json:"match" gorm:"foreignKey:MatchID"`
	Kills               int       `json:"kills" gorm:"default:0"`
	Revives             int       `json:"revives" gorm:"default:0"`
	Knockouts           int       `json:"knockouts" gorm:"default:0"`
	SurvivalTimeMinutes int       `json:"survival_time_minutes" gorm:"default:0"`
	IsMVP               bool      `json:"is_mvp" gorm:"default:false"`
	TeamKillPenalty     int       `json:"team_kill_penalty" gorm:"default:0"`
	TotalPoints         float64   `json:"total_points" gorm:"default:0.00"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type Transaction struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID          uuid.UUID `json:"user_id"`
	User            User      `json:"user" gorm:"foreignKey:UserID"`
	Amount          float64   `json:"amount" gorm:"not null"`
	Type            string    `json:"type" gorm:"not null"` // deposit, withdrawal, contest_entry, winnings
	Status          string    `json:"status" gorm:"default:pending"` // pending, completed, failed
	RelatedEntityID *uuid.UUID `json:"related_entity_id"` // contest_id or other reference
	PaymentID       string    `json:"payment_id"` // Razorpay payment ID
	OrderID         string    `json:"order_id"`   // Razorpay order ID
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Request/Response DTOs
type LoginRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

type VerifyOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	OTP         string `json:"otp" binding:"required"`
}

type CreateTeamRequest struct {
	ContestID   uuid.UUID   `json:"contest_id" binding:"required"`
	TeamName    string      `json:"team_name" binding:"required"`
	PlayerIDs   []uuid.UUID `json:"player_ids" binding:"required"`
	CaptainID   uuid.UUID   `json:"captain_id" binding:"required"`
	ViceCaptainID uuid.UUID `json:"vice_captain_id" binding:"required"`
}

type UpdateStatsRequest struct {
	Kills               int  `json:"kills"`
	Revives             int  `json:"revives"`
	Knockouts           int  `json:"knockouts"`
	SurvivalTimeMinutes int  `json:"survival_time_minutes"`
	IsMVP               bool `json:"is_mvp"`
	TeamKillPenalty     int  `json:"team_kill_penalty"`
}

type CreatePaymentOrderRequest struct {
	Amount    float64 `json:"amount" binding:"required"`
	Currency  string  `json:"currency"`
	ContestID uuid.UUID `json:"contest_id"`
}

// === ADMIN-CONTROLLED MODELS ===

// UsernamePrefix - Admin can manage username prefixes
type UsernamePrefix struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Prefix    string    `json:"prefix" gorm:"unique;not null"` // e.g., "PLAYER_", "GAMER_", "PRO_"
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Game - Admin can add/manage different games
type Game struct {
	ID                uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name              string        `json:"name" gorm:"unique;not null"` // e.g., "BGMI", "Valorant", "Apex Legends"
	DisplayName       string        `json:"display_name" gorm:"not null"`
	Icon              string        `json:"icon" gorm:"type:text"` // Base64 encoded icon
	IsActive          bool          `json:"is_active" gorm:"default:true"`
	MaxPlayersPerTeam int           `json:"max_players_per_team" gorm:"default:6"`
	MinPlayersPerTeam int           `json:"min_players_per_team" gorm:"default:4"`
	ScoringRules      string        `json:"scoring_rules" gorm:"type:jsonb"` // JSON structure for scoring
	PlayerRoles       string        `json:"player_roles" gorm:"type:jsonb"` // JSON array of available roles
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
	Tournaments       []Tournament  `json:"tournaments" gorm:"foreignKey:GameID"`
}

// GameScoringRule - Flexible scoring system per game
type GameScoringRule struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	GameID       uuid.UUID `json:"game_id"`
	Game         Game      `json:"game" gorm:"foreignKey:GameID"`
	ActionType   string    `json:"action_type" gorm:"not null"` // kill, assist, revive, death, etc.
	Points       float64   `json:"points" gorm:"not null"`
	Description  string    `json:"description"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Achievement - Gamification achievements
type Achievement struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"unique;not null"`
	Description string    `json:"description" gorm:"not null"`
	Icon        string    `json:"icon" gorm:"type:text"` // Base64 encoded badge icon
	Category    string    `json:"category" gorm:"not null"` // wins, kills, participation, etc.
	Criteria    string    `json:"criteria" gorm:"type:jsonb"` // JSON criteria for achievement
	Points      int       `json:"points" gorm:"default:0"` // Points awarded for achievement
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserAchievement - Track user achievements
type UserAchievement struct {
	ID            uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        uuid.UUID   `json:"user_id"`
	User          User        `json:"user" gorm:"foreignKey:UserID"`
	AchievementID uuid.UUID   `json:"achievement_id"`
	Achievement   Achievement `json:"achievement" gorm:"foreignKey:AchievementID"`
	UnlockedAt    time.Time   `json:"unlocked_at"`
	CreatedAt     time.Time   `json:"created_at"`
}

// ContestTemplate - Admin can create contest templates
type ContestTemplate struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name           string    `json:"name" gorm:"not null"`
	GameID         uuid.UUID `json:"game_id"`
	Game           Game      `json:"game" gorm:"foreignKey:GameID"`
	EntryFee       float64   `json:"entry_fee" gorm:"not null"`
	PrizeStructure string    `json:"prize_structure" gorm:"type:jsonb"` // JSON prize distribution
	MaxEntries     int       `json:"max_entries" gorm:"not null"`
	IsVIP          bool      `json:"is_vip" gorm:"default:false"`
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PlayerAnalytics - Enhanced player analytics
type PlayerAnalytics struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PlayerID        uuid.UUID `json:"player_id"`
	Player          Player    `json:"player" gorm:"foreignKey:PlayerID"`
	GameID          uuid.UUID `json:"game_id"`
	Game            Game      `json:"game" gorm:"foreignKey:GameID"`
	TotalMatches    int       `json:"total_matches" gorm:"default:0"`
	TotalKills      int       `json:"total_kills" gorm:"default:0"`
	TotalDeaths     int       `json:"total_deaths" gorm:"default:0"`
	TotalAssists    int       `json:"total_assists" gorm:"default:0"`
	AvgPerformance  float64   `json:"avg_performance" gorm:"default:0.00"`
	Form            string    `json:"form" gorm:"default:stable"` // hot, cold, stable
	InjuryStatus    string    `json:"injury_status" gorm:"default:fit"` // fit, injured, doubtful
	LastUpdated     time.Time `json:"last_updated"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// SeasonLeague - Multi-tournament season leagues
type SeasonLeague struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name          string    `json:"name" gorm:"not null"`
	GameID        uuid.UUID `json:"game_id"`
	Game          Game      `json:"game" gorm:"foreignKey:GameID"`
	StartDate     time.Time `json:"start_date" gorm:"not null"`
	EndDate       time.Time `json:"end_date" gorm:"not null"`
	EntryFee      float64   `json:"entry_fee" gorm:"not null"`
	PrizePool     float64   `json:"prize_pool" gorm:"not null"`
	MaxParticipants int     `json:"max_participants" gorm:"not null"`
	Status        string    `json:"status" gorm:"default:upcoming"` // upcoming, active, completed
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// === REQUEST/RESPONSE DTOs FOR NEW FEATURES ===

// UpdateProfileRequest - Enhanced profile update
type UpdateProfileRequest struct {
	Name         string `json:"name" binding:"required"`
	Username     string `json:"username" binding:"required"`
	ProfileImage string `json:"profile_image"` // Base64 encoded image
}

// UsernameGenerationRequest - For generating username
type UsernameGenerationRequest struct {
	PrefixID uuid.UUID `json:"prefix_id" binding:"required"`
}

// CreateGameRequest - Admin creates new game
type CreateGameRequest struct {
	Name              string   `json:"name" binding:"required"`
	DisplayName       string   `json:"display_name" binding:"required"`
	Icon              string   `json:"icon"` // Base64 encoded
	MaxPlayersPerTeam int      `json:"max_players_per_team" binding:"required"`
	MinPlayersPerTeam int      `json:"min_players_per_team" binding:"required"`
	PlayerRoles       []string `json:"player_roles" binding:"required"`
}

// CreateScoringRuleRequest - Admin creates scoring rules
type CreateScoringRuleRequest struct {
	GameID      uuid.UUID `json:"game_id" binding:"required"`
	ActionType  string    `json:"action_type" binding:"required"`
	Points      float64   `json:"points" binding:"required"`
	Description string    `json:"description"`
}

// LiveMatchUpdateRequest - Admin updates match data live
type LiveMatchUpdateRequest struct {
	PlayerID            uuid.UUID `json:"player_id" binding:"required"`
	ActionType          string    `json:"action_type" binding:"required"` // kill, death, assist, etc.
	Value               int       `json:"value" gorm:"default:1"`
	Timestamp           time.Time `json:"timestamp"`
}

// ReferralRequest - User referral
type ReferralRequest struct {
	ReferralCode string `json:"referral_code" binding:"required"`
}