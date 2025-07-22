package services

import (
	"esports-fantasy-backend/config"
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/repository"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService interface {
	SendOTP(phoneNumber string) error
	VerifyOTP(phoneNumber, otpCode string) (*models.User, string, error)
	ValidateToken(tokenString string) (*models.User, error)
}

type authService struct {
	userRepo repository.UserRepository
	config   *config.Config
}

func NewAuthService(userRepo repository.UserRepository, config *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		config:   config,
	}
}

func (s *authService) SendOTP(phoneNumber string) error {
	// Generate 6-digit OTP
	otpCode := s.generateOTP()
	
	// Create OTP record
	otp := &models.OTP{
		ID:          uuid.New(),
		PhoneNumber: phoneNumber,
		Code:        otpCode,
		ExpiresAt:   time.Now().Add(5 * time.Minute), // 5 minutes expiry
		Used:        false,
	}

	if err := s.userRepo.CreateOTP(otp); err != nil {
		return fmt.Errorf("failed to create OTP: %w", err)
	}

	// For development - show OTP in console
	if s.config.OTPConsole {
		log.Printf("🔐 OTP for %s: %s (expires in 5 minutes)", phoneNumber, otpCode)
		fmt.Printf("\n==========================================\n")
		fmt.Printf("📱 OTP SENT TO: %s\n", phoneNumber)
		fmt.Printf("🔢 YOUR OTP: %s\n", otpCode)
		fmt.Printf("⏰ Expires in: 5 minutes\n")
		fmt.Printf("==========================================\n\n")
	}

	return nil
}

func (s *authService) VerifyOTP(phoneNumber, otpCode string) (*models.User, string, error) {
	// Validate OTP
	otp, err := s.userRepo.GetValidOTP(phoneNumber, otpCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, "", fmt.Errorf("invalid or expired OTP")
		}
		return nil, "", fmt.Errorf("failed to validate OTP: %w", err)
	}

	// Mark OTP as used
	if err := s.userRepo.MarkOTPAsUsed(otp.ID); err != nil {
		return nil, "", fmt.Errorf("failed to mark OTP as used: %w", err)
	}

	// Get or create user
	user, err := s.userRepo.GetUserByPhone(phoneNumber)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new user
			user = &models.User{
				ID:            uuid.New(),
				PhoneNumber:   phoneNumber,
				WalletBalance: 0.0,
				IsAdmin:       false,
			}
			if err := s.userRepo.CreateUser(user); err != nil {
				return nil, "", fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			return nil, "", fmt.Errorf("failed to get user: %w", err)
		}
	}

	// Generate JWT token
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	log.Printf("✅ User authenticated: %s (ID: %s)", phoneNumber, user.ID)

	return user, token, nil
}

func (s *authService) ValidateToken(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_id in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id format")
	}

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

func (s *authService) generateOTP() string {
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(900000) + 100000 // Generate 6-digit number
	return strconv.Itoa(otp)
}

func (s *authService) generateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":      user.ID.String(),
		"phone_number": user.PhoneNumber,
		"is_admin":     user.IsAdmin,
		"exp":          time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
		"iat":          time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}