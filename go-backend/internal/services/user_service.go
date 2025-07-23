package services

import (
	"esports-fantasy-backend/config"
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type UserService interface {
	GetProfile(userID uuid.UUID) (*models.User, error)
	UpdateProfile(userID uuid.UUID, name string) error
	UpdateUserProfile(userID uuid.UUID, name, username, profileImage string) error
	UpdateProfileImage(userID uuid.UUID, profileImage string) error
	GetWalletBalance(userID uuid.UUID) (float64, error)
}

type userService struct {
	userRepo repository.UserRepository
	config   *config.Config
}

func NewUserService(userRepo repository.UserRepository, config *config.Config) UserService {
	return &userService{
		userRepo: userRepo,
		config:   config,
	}
}

func (s *userService) GetProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	return user, nil
}

func (s *userService) UpdateProfile(userID uuid.UUID, name string) error {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.Name = name
	if err := s.userRepo.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *userService) GetWalletBalance(userID uuid.UUID) (float64, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user: %w", err)
	}
	return user.WalletBalance, nil
}

func (s *userService) UpdateUserProfile(userID uuid.UUID, name, username, profileImage string) error {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.Name = name
	user.Username = username
	user.ProfileImage = profileImage
	
	if err := s.userRepo.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
}

func (s *userService) UpdateProfileImage(userID uuid.UUID, profileImage string) error {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.ProfileImage = profileImage
	
	if err := s.userRepo.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update profile image: %w", err)
	}

	return nil
}