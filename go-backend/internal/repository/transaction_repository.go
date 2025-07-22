package repository

import (
	"esports-fantasy-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionRepository interface {
	CreateTransaction(transaction *models.Transaction) error
	GetUserTransactions(userID uuid.UUID) ([]models.Transaction, error)
	GetTransactionByID(id uuid.UUID) (*models.Transaction, error)
	GetTransactionByPaymentID(paymentID string) (*models.Transaction, error)
	UpdateTransaction(transaction *models.Transaction) error
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) CreateTransaction(transaction *models.Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *transactionRepository) GetUserTransactions(userID uuid.UUID) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) GetTransactionByID(id uuid.UUID) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.First(&transaction, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) GetTransactionByPaymentID(paymentID string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.Where("payment_id = ?", paymentID).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) UpdateTransaction(transaction *models.Transaction) error {
	return r.db.Save(transaction).Error
}