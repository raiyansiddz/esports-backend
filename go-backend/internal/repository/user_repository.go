package repository

import (
	"esports-fantasy-backend/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByPhone(phoneNumber string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	UpdateUser(user *models.User) error
	CreateOTP(otp *models.OTP) error
	GetValidOTP(phoneNumber, code string) (*models.OTP, error)
	MarkOTPAsUsed(otpID uuid.UUID) error
	UpdateWalletBalance(userID uuid.UUID, amount float64) error
	
	// New methods for enhanced features
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByPhoneNumber(phoneNumber string) (*models.User, error)
	Update(user *models.User) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetUserByPhone(phoneNumber string) (*models.User, error) {
	var user models.User
	err := r.db.Where("phone_number = ?", phoneNumber).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) CreateOTP(otp *models.OTP) error {
	return r.db.Create(otp).Error
}

func (r *userRepository) GetValidOTP(phoneNumber, code string) (*models.OTP, error) {
	var otp models.OTP
	err := r.db.Where("phone_number = ? AND code = ? AND used = false AND expires_at > ?", 
		phoneNumber, code, time.Now()).First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *userRepository) MarkOTPAsUsed(otpID uuid.UUID) error {
	return r.db.Model(&models.OTP{}).Where("id = ?", otpID).Update("used", true).Error
}

func (r *userRepository) UpdateWalletBalance(userID uuid.UUID, amount float64) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Update("wallet_balance", gorm.Expr("wallet_balance + ?", amount)).Error
}

// New methods for enhanced features
func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByPhoneNumber(phoneNumber string) (*models.User, error) {
	var user models.User
	err := r.db.Where("phone_number = ?", phoneNumber).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}