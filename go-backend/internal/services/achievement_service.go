package services

import (
	"encoding/json"
	"esports-fantasy-backend/config"
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AchievementService interface {
	CreateAchievement(req *CreateAchievementRequest) (*models.Achievement, error)
	GetAllAchievements() ([]models.Achievement, error)
	GetActiveAchievements() ([]models.Achievement, error)
	GetAchievementByID(id uuid.UUID) (*models.Achievement, error)
	UpdateAchievement(id uuid.UUID, req *CreateAchievementRequest) error
	ToggleAchievementStatus(id uuid.UUID) error
	DeleteAchievement(id uuid.UUID) error
	
	// User achievements
	GetUserAchievements(userID uuid.UUID) ([]models.UserAchievement, error)
	UnlockAchievementForUser(userID, achievementID uuid.UUID) error
	CheckAndUnlockAchievements(userID uuid.UUID) error
}

type CreateAchievementRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description" binding:"required"`
	Icon        string                 `json:"icon"` // Base64 encoded icon
	Category    string                 `json:"category" binding:"required"`
	Criteria    map[string]interface{} `json:"criteria" binding:"required"` // JSON criteria
	Points      int                    `json:"points"`
}

type achievementService struct {
	achievementRepo     repository.AchievementRepository
	userAchievementRepo repository.UserAchievementRepository
	userRepo            repository.UserRepository
	config              *config.Config
}

func NewAchievementService(
	achievementRepo repository.AchievementRepository,
	userAchievementRepo repository.UserAchievementRepository,
	userRepo repository.UserRepository,
	config *config.Config,
) AchievementService {
	return &achievementService{
		achievementRepo:     achievementRepo,
		userAchievementRepo: userAchievementRepo,
		userRepo:            userRepo,
		config:              config,
	}
}

func (s *achievementService) CreateAchievement(req *CreateAchievementRequest) (*models.Achievement, error) {
	// Convert criteria to JSON
	criteriaJSON, err := json.Marshal(req.Criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize criteria: %w", err)
	}

	achievement := &models.Achievement{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		Category:    req.Category,
		Criteria:    string(criteriaJSON),
		Points:      req.Points,
		IsActive:    true,
	}

	if err := s.achievementRepo.Create(achievement); err != nil {
		return nil, fmt.Errorf("failed to create achievement: %w", err)
	}

	return achievement, nil
}

func (s *achievementService) GetAllAchievements() ([]models.Achievement, error) {
	return s.achievementRepo.GetAll()
}

func (s *achievementService) GetActiveAchievements() ([]models.Achievement, error) {
	return s.achievementRepo.GetActive()
}

func (s *achievementService) GetAchievementByID(id uuid.UUID) (*models.Achievement, error) {
	return s.achievementRepo.GetByID(id)
}

func (s *achievementService) UpdateAchievement(id uuid.UUID, req *CreateAchievementRequest) error {
	achievement, err := s.achievementRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get achievement: %w", err)
	}

	// Convert criteria to JSON
	criteriaJSON, err := json.Marshal(req.Criteria)
	if err != nil {
		return fmt.Errorf("failed to serialize criteria: %w", err)
	}

	achievement.Name = req.Name
	achievement.Description = req.Description
	achievement.Icon = req.Icon
	achievement.Category = req.Category
	achievement.Criteria = string(criteriaJSON)
	achievement.Points = req.Points

	if err := s.achievementRepo.Update(achievement); err != nil {
		return fmt.Errorf("failed to update achievement: %w", err)
	}

	return nil
}

func (s *achievementService) ToggleAchievementStatus(id uuid.UUID) error {
	achievement, err := s.achievementRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get achievement: %w", err)
	}

	achievement.IsActive = !achievement.IsActive

	if err := s.achievementRepo.Update(achievement); err != nil {
		return fmt.Errorf("failed to update achievement status: %w", err)
	}

	return nil
}

func (s *achievementService) DeleteAchievement(id uuid.UUID) error {
	return s.achievementRepo.Delete(id)
}

func (s *achievementService) GetUserAchievements(userID uuid.UUID) ([]models.UserAchievement, error) {
	return s.userAchievementRepo.GetByUserID(userID)
}

func (s *achievementService) UnlockAchievementForUser(userID, achievementID uuid.UUID) error {
	// Check if already unlocked
	if s.userAchievementRepo.HasUserUnlockedAchievement(userID, achievementID) {
		return fmt.Errorf("achievement already unlocked")
	}

	// Get achievement to add points to user
	achievement, err := s.achievementRepo.GetByID(achievementID)
	if err != nil {
		return fmt.Errorf("failed to get achievement: %w", err)
	}

	// Create user achievement record
	userAchievement := &models.UserAchievement{
		UserID:        userID,
		AchievementID: achievementID,
		UnlockedAt:    time.Now(),
	}

	if err := s.userAchievementRepo.Create(userAchievement); err != nil {
		return fmt.Errorf("failed to unlock achievement: %w", err)
	}

	// Update user's total points
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.TotalPoints += int64(achievement.Points)
	if err := s.userRepo.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update user points: %w", err)
	}

	return nil
}

func (s *achievementService) CheckAndUnlockAchievements(userID uuid.UUID) error {
	// Get all active achievements
	achievements, err := s.achievementRepo.GetActive()
	if err != nil {
		return fmt.Errorf("failed to get achievements: %w", err)
	}

	// Get user data
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Check each achievement
	for _, achievement := range achievements {
		// Skip if already unlocked
		if s.userAchievementRepo.HasUserUnlockedAchievement(userID, achievement.ID) {
			continue
		}

		// Parse criteria and check if user meets it
		var criteria map[string]interface{}
		if err := json.Unmarshal([]byte(achievement.Criteria), &criteria); err != nil {
			continue // Skip malformed criteria
		}

		if s.checkAchievementCriteria(user, criteria) {
			// Unlock achievement
			s.UnlockAchievementForUser(userID, achievement.ID)
		}
	}

	return nil
}

func (s *achievementService) checkAchievementCriteria(user *models.User, criteria map[string]interface{}) bool {
	// Basic criteria checking logic
	if totalPointsRequired, ok := criteria["total_points"].(float64); ok {
		if user.TotalPoints < int64(totalPointsRequired) {
			return false
		}
	}

	if consecutiveWinsRequired, ok := criteria["consecutive_wins"].(float64); ok {
		if user.ConsecutiveWins < int(consecutiveWinsRequired) {
			return false
		}
	}

	// Add more criteria checks as needed
	return true
}