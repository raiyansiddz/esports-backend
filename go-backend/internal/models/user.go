package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PhoneNumber   string         `json:"phone_number" gorm:"unique;not null"`
	Name          string         `json:"name"`
	WalletBalance float64        `json:"wallet_balance" gorm:"default:0.00"`
	IsAdmin       bool           `json:"is_admin" gorm:"default:false"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
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
	GameType  string    `json:"game_type" gorm:"not null"` // BGMI, Valorant, etc.
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