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

type UsernameService interface {
	GenerateUsername(prefixID uuid.UUID) (string, error)
	IsUsernameAvailable(username string) bool
	UpdateUsername(userID uuid.UUID, newUsername string) error
	GetUsernamePrefixes() ([]models.UsernamePrefix, error)
	CreateUsernamePrefix(prefix string) (*models.UsernamePrefix, error)
	UpdateUsernamePrefix(id uuid.UUID, prefix string, isActive bool) error
	DeleteUsernamePrefix(id uuid.UUID) error
}

type usernameService struct {
	userRepo   repository.UserRepository
	prefixRepo repository.UsernamePrefixRepository
	config     *config.Config
}

func NewUsernameService(userRepo repository.UserRepository, prefixRepo repository.UsernamePrefixRepository, config *config.Config) UsernameService {
	return &usernameService{
		userRepo:   userRepo,
		prefixRepo: prefixRepo,
		config:     config,
	}
}

func (s *usernameService) GenerateUsername(prefixID uuid.UUID) (string, error) {
	// Get the prefix
	prefix, err := s.prefixRepo.GetByID(prefixID)
	if err != nil {
		return "", fmt.Errorf("failed to get prefix: %w", err)
	}

	if !prefix.IsActive {
		return "", fmt.Errorf("prefix is not active")
	}

	// Generate a random 5-digit number
	max := big.NewInt(99999)
	min := big.NewInt(10000)
	n, err := rand.Int(rand.Reader, max.Sub(max, min))
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}
	
	randomNumber := n.Add(n, min).Int64()
	
	// Create username with prefix and number
	username := fmt.Sprintf("%s%d", prefix.Prefix, randomNumber)
	
	// Check if username is available, if not, try again (max 5 attempts)
	for i := 0; i < 5; i++ {
		if s.IsUsernameAvailable(username) {
			return username, nil
		}
		
		// Generate new random number
		n, err := rand.Int(rand.Reader, max.Sub(max, min))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		randomNumber = n.Add(n, min).Int64()
		username = fmt.Sprintf("%s%d", prefix.Prefix, randomNumber)
	}
	
	return "", fmt.Errorf("failed to generate unique username after 5 attempts")
}

func (s *usernameService) IsUsernameAvailable(username string) bool {
	user, err := s.userRepo.GetByUsername(username)
	return err != nil || user == nil
}

func (s *usernameService) UpdateUsername(userID uuid.UUID, newUsername string) error {
	// Check if username is available
	if !s.IsUsernameAvailable(newUsername) {
		return fmt.Errorf("username is already taken")
	}

	// Validate username format (alphanumeric and underscores only, 3-20 characters)
	if len(newUsername) < 3 || len(newUsername) > 20 {
		return fmt.Errorf("username must be between 3 and 20 characters")
	}

	for _, char := range newUsername {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || char == '_') {
			return fmt.Errorf("username can only contain letters, numbers, and underscores")
		}
	}

	// Update username
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.Username = newUsername
	if err := s.userRepo.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update username: %w", err)
	}

	return nil
}

func (s *usernameService) GetUsernamePrefixes() ([]models.UsernamePrefix, error) {
	return s.prefixRepo.GetAll()
}

func (s *usernameService) CreateUsernamePrefix(prefix string) (*models.UsernamePrefix, error) {
	// Validate prefix format
	prefix = strings.TrimSpace(prefix)
	if len(prefix) < 2 || len(prefix) > 10 {
		return nil, fmt.Errorf("prefix must be between 2 and 10 characters")
	}

	// Check if prefix already exists
	existing, err := s.prefixRepo.GetByPrefix(prefix)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("prefix already exists")
	}

	usernamePrefix := &models.UsernamePrefix{
		Prefix:   prefix,
		IsActive: true,
	}

	if err := s.prefixRepo.Create(usernamePrefix); err != nil {
		return nil, fmt.Errorf("failed to create prefix: %w", err)
	}

	return usernamePrefix, nil
}

func (s *usernameService) UpdateUsernamePrefix(id uuid.UUID, prefix string, isActive bool) error {
	existingPrefix, err := s.prefixRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get prefix: %w", err)
	}

	// Validate prefix format if it's being changed
	if existingPrefix.Prefix != prefix {
		prefix = strings.TrimSpace(prefix)
		if len(prefix) < 2 || len(prefix) > 10 {
			return fmt.Errorf("prefix must be between 2 and 10 characters")
		}

		// Check if new prefix already exists
		existing, err := s.prefixRepo.GetByPrefix(prefix)
		if err == nil && existing != nil && existing.ID != id {
			return fmt.Errorf("prefix already exists")
		}
	}

	existingPrefix.Prefix = prefix
	existingPrefix.IsActive = isActive

	if err := s.prefixRepo.Update(existingPrefix); err != nil {
		return fmt.Errorf("failed to update prefix: %w", err)
	}

	return nil
}

func (s *usernameService) DeleteUsernamePrefix(id uuid.UUID) error {
	return s.prefixRepo.Delete(id)
}