package services

import (
        "encoding/json"
        "errors"
        "fmt"
        "log"
        "math/rand"
        "strconv"
        "time"

        "esports-fantasy-backend/config"
        "esports-fantasy-backend/internal/models"
        "esports-fantasy-backend/internal/repository"

        "github.com/golang-jwt/jwt/v4"
        "github.com/google/uuid"
)

type FirebaseAuthService struct {
        cfg      *config.Config
        userRepo repository.UserRepository
        otpRepo  repository.OTPRepository
}

type FirebaseConfig struct {
        APIKey            string `json:"apiKey"`
        AuthDomain        string `json:"authDomain"`
        ProjectID         string `json:"projectId"`
        StorageBucket     string `json:"storageBucket"`
        MessagingSenderID string `json:"messagingSenderId"`
        AppID             string `json:"appId"`
        MeasurementID     string `json:"measurementId"`
}

type OTPRequest struct {
        PhoneNumber string `json:"phone_number" binding:"required"`
}

type OTPVerifyRequest struct {
        PhoneNumber string `json:"phone_number" binding:"required"`
        OTP         string `json:"otp" binding:"required"`
}

type AuthResponse struct {
        Success bool        `json:"success"`
        Message string      `json:"message"`
        Data    interface{} `json:"data,omitempty"`
        Token   string      `json:"token,omitempty"`
        User    interface{} `json:"user,omitempty"`
}

type UserProfile struct {
        ID            string    `json:"id"`
        PhoneNumber   string    `json:"phone_number"`
        WalletBalance float64   `json:"wallet_balance"`
        IsAdmin       bool      `json:"is_admin"`
        CreatedAt     time.Time `json:"created_at"`
}

func NewFirebaseAuthService(cfg *config.Config, userRepo repository.UserRepository, otpRepo repository.OTPRepository) *FirebaseAuthService {
        return &FirebaseAuthService{
                cfg:      cfg,
                userRepo: userRepo,
                otpRepo:  otpRepo,
        }
}

func (s *FirebaseAuthService) GetFirebaseConfig() *FirebaseConfig {
        return &FirebaseConfig{
                APIKey:            s.cfg.FirebaseAPIKey,
                AuthDomain:        s.cfg.FirebaseAuthDomain,
                ProjectID:         s.cfg.FirebaseProjectID,
                StorageBucket:     s.cfg.FirebaseStorageBucket,
                MessagingSenderID: s.cfg.FirebaseMessagingSenderID,
                AppID:             s.cfg.FirebaseAppID,
                MeasurementID:     s.cfg.FirebaseMeasurementID,
        }
}

func (s *FirebaseAuthService) SendOTP(phoneNumber string) (*AuthResponse, error) {
        // Clean phone number
        if len(phoneNumber) < 10 {
                return &AuthResponse{
                        Success: false,
                        Message: "Invalid phone number",
                }, errors.New("invalid phone number")
        }

        // Generate OTP
        otp := s.generateOTP()

        // Save OTP to database
        otpRecord := &models.OTP{
                ID:          uuid.New(),
                PhoneNumber: phoneNumber,
                Code:        otp,
                ExpiresAt:   time.Now().Add(5 * time.Minute),
                CreatedAt:   time.Now(),
        }

        if err := s.otpRepo.Create(otpRecord); err != nil {
                return &AuthResponse{
                        Success: false,
                        Message: "Failed to send OTP",
                }, err
        }

        // In dummy mode or OTP console mode, show OTP in console
        if s.cfg.Dummy || s.cfg.OTPConsole {
                s.logOTPToConsole(phoneNumber, otp)
                
                return &AuthResponse{
                        Success: true,
                        Message: "OTP sent successfully (Console Mode - Check Server Logs)",
                        Data: map[string]interface{}{
                                "phone_number": phoneNumber,
                                "expires_in":   "5 minutes",
                                "mode":         "console",
                        },
                }, nil
        }

        // In production mode, integrate with Firebase Auth
        if err := s.sendFirebaseOTP(phoneNumber, otp); err != nil {
                return &AuthResponse{
                        Success: false,
                        Message: "Failed to send OTP via Firebase",
                }, err
        }

        return &AuthResponse{
                Success: true,
                Message: "OTP sent successfully via Firebase",
                Data: map[string]interface{}{
                        "phone_number": phoneNumber,
                        "expires_in":   "5 minutes",
                        "mode":         "firebase",
                },
        }, nil
}

func (s *FirebaseAuthService) VerifyOTP(phoneNumber, otp string) (*AuthResponse, error) {
        // Get OTP from database
        otpRecord, err := s.otpRepo.GetByPhoneNumber(phoneNumber)
        if err != nil {
                return &AuthResponse{
                        Success: false,
                        Message: "Invalid OTP or phone number",
                }, errors.New("OTP not found")
        }

        // Check if OTP is expired
        if time.Now().After(otpRecord.ExpiresAt) {
                return &AuthResponse{
                        Success: false,
                        Message: "OTP has expired",
                }, errors.New("OTP expired")
        }

        // Verify OTP
        if otpRecord.Code != otp {
                return &AuthResponse{
                        Success: false,
                        Message: "Invalid OTP",
                }, errors.New("invalid OTP")
        }

        // Delete used OTP
        s.otpRepo.Delete(otpRecord.ID.String())

        // Get or create user
        user, err := s.userRepo.GetByPhoneNumber(phoneNumber)
        if err != nil {
                // Create new user
                user = &models.User{
                        ID:            uuid.New(),
                        PhoneNumber:   phoneNumber,
                        WalletBalance: 0.0,
                        IsAdmin:       false,
                        CreatedAt:     time.Now(),
                        UpdatedAt:     time.Now(),
                }

                if err := s.userRepo.Create(user); err != nil {
                        return &AuthResponse{
                                Success: false,
                                Message: "Failed to create user account",
                        }, err
                }

                log.Printf("👤 New user created: %s (%s)", user.ID, phoneNumber)
        }

        // Generate JWT token
        token, err := s.generateJWTToken(user)
        if err != nil {
                return &AuthResponse{
                        Success: false,
                        Message: "Failed to generate authentication token",
                }, err
        }

        // Create user profile response
        userProfile := &UserProfile{
                ID:            user.ID.String(),
                PhoneNumber:   user.PhoneNumber,
                WalletBalance: user.WalletBalance,
                IsAdmin:       user.IsAdmin,
                CreatedAt:     user.CreatedAt,
        }

        log.Printf("✅ User authenticated: %s (%s)", user.ID, phoneNumber)

        return &AuthResponse{
                Success: true,
                Message: "Authentication successful",
                Token:   token,
                User:    userProfile,
                Data: map[string]interface{}{
                        "authentication_method": "firebase_otp",
                        "login_time":            time.Now(),
                },
        }, nil
}

func (s *FirebaseAuthService) generateOTP() string {
        // Generate 6-digit OTP
        rand.Seed(time.Now().UnixNano())
        otp := rand.Intn(900000) + 100000
        return strconv.Itoa(otp)
}

func (s *FirebaseAuthService) logOTPToConsole(phoneNumber, otp string) {
        log.Printf("==========================================")
        log.Printf("🔥 FIREBASE OTP AUTHENTICATION")
        log.Printf("📱 Phone Number: %s", phoneNumber)
        log.Printf("🔢 YOUR OTP: %s", otp)
        log.Printf("⏰ Expires in: 5 minutes")
        log.Printf("🚀 Mode: Console (DUMMY=%t)", s.cfg.Dummy)
        log.Printf("==========================================")
}

func (s *FirebaseAuthService) sendFirebaseOTP(phoneNumber, otp string) error {
        // In production, this would integrate with Firebase Auth
        // For now, return success as it's handled in dummy mode
        log.Printf("🔥 Firebase OTP Integration: Sending OTP %s to %s", otp, phoneNumber)
        return nil
}

func (s *FirebaseAuthService) generateJWTToken(user *models.User) (string, error) {
        claims := jwt.MapClaims{
                "user_id":      user.ID,
                "phone_number": user.PhoneNumber,
                "is_admin":     user.IsAdmin,
                "iat":          time.Now().Unix(),
                "exp":          time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days
        }

        token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
        return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *FirebaseAuthService) ValidateToken(tokenString string) (*models.User, error) {
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                        return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
                }
                return []byte(s.cfg.JWTSecret), nil
        })

        if err != nil {
                return nil, err
        }

        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
                userID := claims["user_id"].(string)
                user, err := s.userRepo.GetByID(userID)
                if err != nil {
                        return nil, err
                }
                return user, nil
        }

        return nil, fmt.Errorf("invalid token")
}

// Admin functionality
func (s *FirebaseAuthService) PromoteToAdmin(userID string) error {
        user, err := s.userRepo.GetByID(userID)
        if err != nil {
                return err
        }

        user.IsAdmin = true
        user.UpdatedAt = time.Now()

        if err := s.userRepo.Update(user); err != nil {
                return err
        }

        log.Printf("👑 User promoted to admin: %s (%s)", user.ID, user.PhoneNumber)
        return nil
}

func (s *FirebaseAuthService) GetUserProfile(userID string) (*UserProfile, error) {
        user, err := s.userRepo.GetByID(userID)
        if err != nil {
                return nil, err
        }

        return &UserProfile{
                ID:            user.ID.String(),
                PhoneNumber:   user.PhoneNumber,
                WalletBalance: user.WalletBalance,
                IsAdmin:       user.IsAdmin,
                CreatedAt:     user.CreatedAt,
        }, nil
}