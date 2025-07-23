package services

import (
	"crypto/rand"
	"esports-fantasy-backend/config"
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/repository"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
)

type ReferralService interface {
	GenerateReferralCode(userID uuid.UUID) (string, error)
	ApplyReferralCode(userID uuid.UUID, referralCode string) error
	GetReferralStats(userID uuid.UUID) (*ReferralStats, error)
	GetReferralLeaderboard(limit int) ([]ReferralEntry, error)
}

type ReferralStats struct {
	ReferralCode      string  `json:"referral_code"`
	TotalReferrals    int     `json:"total_referrals"`
	TotalBonusEarned  float64 `json:"total_bonus_earned"`
	ReferredUsers     []ReferredUser `json:"referred_users"`
}

type ReferredUser struct {
	Username     string  `json:"username"`
	JoinedDate   string  `json:"joined_date"`
	BonusEarned  float64 `json:"bonus_earned"`
	IsActive     bool    `json:"is_active"`
}

type ReferralEntry struct {
	UserID         uuid.UUID `json:"user_id"`
	Username       string    `json:"username"`
	ReferralCode   string    `json:"referral_code"`
	TotalReferrals int       `json:"total_referrals"`
	TotalBonus     float64   `json:"total_bonus"`
	Rank           int       `json:"rank"`
}

type referralService struct {
	userRepo repository.UserRepository
	config   *config.Config
}

func NewReferralService(userRepo repository.UserRepository, config *config.Config) ReferralService {
	return &referralService{
		userRepo: userRepo,
		config:   config,
	}
}

func (s *referralService) GenerateReferralCode(userID uuid.UUID) (string, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	// If user already has a referral code, return it
	if user.ReferralCode != "" {
		return user.ReferralCode, nil
	}

	// Generate a new referral code
	referralCode, err := s.generateUniqueReferralCode()
	if err != nil {
		return "", fmt.Errorf("failed to generate referral code: %w", err)
	}

	// Update user with referral code
	user.ReferralCode = referralCode
	if err := s.userRepo.UpdateUser(user); err != nil {
		return "", fmt.Errorf("failed to save referral code: %w", err)
	}

	return referralCode, nil
}

func (s *referralService) ApplyReferralCode(userID uuid.UUID, referralCode string) error {
	// Get the user applying the referral code
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if user already used a referral code
	if user.ReferredBy != nil {
		return fmt.Errorf("user has already used a referral code")
	}

	// Find the user who owns the referral code
	referrer, err := s.userRepo.GetByReferralCode(referralCode)
	if err != nil {
		return fmt.Errorf("invalid referral code")
	}

	// Can't refer yourself
	if referrer.ID == userID {
		return fmt.Errorf("cannot use your own referral code")
	}

	// Calculate referral bonus (configurable)
	referralBonus := 50.0 // ₹50 bonus for referrer
	newUserBonus := 25.0  // ₹25 bonus for new user

	// Update referred user
	user.ReferredBy = &referrer.ID
	user.WalletBalance += newUserBonus
	if err := s.userRepo.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update referred user: %w", err)
	}

	// Update referrer
	referrer.ReferralBonus += referralBonus
	referrer.WalletBalance += referralBonus
	if err := s.userRepo.UpdateUser(referrer); err != nil {
		return fmt.Errorf("failed to update referrer: %w", err)
	}

	return nil
}

func (s *referralService) GetReferralStats(userID uuid.UUID) (*ReferralStats, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get all users referred by this user
	referredUsers, err := s.userRepo.GetReferredUsers(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get referred users: %w", err)
	}

	// Convert to response format
	var referredUsersList []ReferredUser
	for _, referredUser := range referredUsers {
		referredUsersList = append(referredUsersList, ReferredUser{
			Username:    referredUser.Username,
			JoinedDate:  referredUser.CreatedAt.Format("2006-01-02"),
			BonusEarned: 50.0, // Fixed bonus per referral
			IsActive:    true, // You can add logic to determine activity
		})
	}

	return &ReferralStats{
		ReferralCode:     user.ReferralCode,
		TotalReferrals:   len(referredUsers),
		TotalBonusEarned: user.ReferralBonus,
		ReferredUsers:    referredUsersList,
	}, nil
}

func (s *referralService) GetReferralLeaderboard(limit int) ([]ReferralEntry, error) {
	users, err := s.userRepo.GetTopReferrers(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top referrers: %w", err)
	}

	var leaderboard []ReferralEntry
	for i, user := range users {
		// Get referral count for each user
		referredUsers, _ := s.userRepo.GetReferredUsers(user.ID)
		
		leaderboard = append(leaderboard, ReferralEntry{
			UserID:         user.ID,
			Username:       user.Username,
			ReferralCode:   user.ReferralCode,
			TotalReferrals: len(referredUsers),
			TotalBonus:     user.ReferralBonus,
			Rank:           i + 1,
		})
	}

	return leaderboard, nil
}

func (s *referralService) generateUniqueReferralCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	for attempts := 0; attempts < 10; attempts++ {
		// Generate random string
		b := make([]byte, length)
		for i := range b {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			b[i] = charset[n.Int64()]
		}

		code := string(b)

		// Check if code already exists
		_, err := s.userRepo.GetByReferralCode(code)
		if err != nil {
			// Code doesn't exist, we can use it
			return code, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique referral code after 10 attempts")
}