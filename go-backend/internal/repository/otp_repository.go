package repository

import (
        "esports-fantasy-backend/internal/models"
        "gorm.io/gorm"
)

type OTPRepository interface {
        Create(otp *models.OTP) error
        GetByPhoneNumber(phoneNumber string) (*models.OTP, error)
        Delete(id string) error
}

type otpRepository struct {
        db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) OTPRepository {
        return &otpRepository{db: db}
}

func (r *otpRepository) Create(otp *models.OTP) error {
        // Delete any existing OTP for this phone number first
        r.db.Where("phone_number = ?", otp.PhoneNumber).Delete(&models.OTP{})
        
        return r.db.Create(otp).Error
}

func (r *otpRepository) GetByPhoneNumber(phoneNumber string) (*models.OTP, error) {
        var otp models.OTP
        err := r.db.Where("phone_number = ?", phoneNumber).First(&otp).Error
        if err != nil {
                return nil, err
        }
        return &otp, nil
}

func (r *otpRepository) Delete(id string) error {
        return r.db.Delete(&models.OTP{}, "id = ?", id).Error
}