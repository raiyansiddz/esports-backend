package services

import (
	"esports-fantasy-backend/config"
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/repository"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type PaymentService interface {
	CreatePaymentOrder(userID uuid.UUID, req *models.CreatePaymentOrderRequest) (map[string]interface{}, error)
	HandlePaymentSuccess(paymentID, orderID string, userID uuid.UUID, amount float64) error
	ProcessContestEntry(userID, contestID uuid.UUID, amount float64) error
}

type paymentService struct {
	transactionRepo repository.TransactionRepository
	userRepo        repository.UserRepository
	config          *config.Config
}

func NewPaymentService(transactionRepo repository.TransactionRepository, userRepo repository.UserRepository, config *config.Config) PaymentService {
	return &paymentService{
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
		config:          config,
	}
}

func (s *paymentService) CreatePaymentOrder(userID uuid.UUID, req *models.CreatePaymentOrderRequest) (map[string]interface{}, error) {
	// For now, create a dummy Razorpay order response
	// In production, this would integrate with actual Razorpay API
	
	orderID := fmt.Sprintf("order_%s", uuid.New().String()[:8])
	
	// Create transaction record
	transaction := &models.Transaction{
		ID:              uuid.New(),
		UserID:          userID,
		Amount:          req.Amount,
		Type:            "deposit",
		Status:          "pending",
		RelatedEntityID: &req.ContestID,
		OrderID:         orderID,
	}

	if err := s.transactionRepo.CreateTransaction(transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Dummy Razorpay order response
	orderResponse := map[string]interface{}{
		"id":                orderID,
		"entity":           "order",
		"amount":           int(req.Amount * 100), // Convert to paise
		"amount_paid":      0,
		"amount_due":       int(req.Amount * 100),
		"currency":         "INR",
		"status":           "created",
		"attempts":         0,
		"created_at":       1234567890,
	}

	log.Printf("💳 Payment order created: %s for user %s (₹%.2f)", orderID, userID, req.Amount)

	return orderResponse, nil
}

func (s *paymentService) HandlePaymentSuccess(paymentID, orderID string, userID uuid.UUID, amount float64) error {
	// Find transaction by order ID
	transaction, err := s.transactionRepo.GetTransactionByPaymentID(paymentID)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	// Update transaction
	transaction.Status = "completed"
	transaction.PaymentID = paymentID

	if err := s.transactionRepo.UpdateTransaction(transaction); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// Update user wallet balance
	if err := s.userRepo.UpdateWalletBalance(userID, amount); err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	log.Printf("✅ Payment successful: %s for user %s (₹%.2f)", paymentID, userID, amount)

	return nil
}

func (s *paymentService) ProcessContestEntry(userID, contestID uuid.UUID, amount float64) error {
	// Create transaction for contest entry
	transaction := &models.Transaction{
		ID:              uuid.New(),
		UserID:          userID,
		Amount:          -amount, // Negative for debit
		Type:            "contest_entry",
		Status:          "completed",
		RelatedEntityID: &contestID,
	}

	if err := s.transactionRepo.CreateTransaction(transaction); err != nil {
		return fmt.Errorf("failed to create contest entry transaction: %w", err)
	}

	// Deduct from wallet
	if err := s.userRepo.UpdateWalletBalance(userID, -amount); err != nil {
		return fmt.Errorf("failed to deduct from wallet: %w", err)
	}

	log.Printf("💰 Contest entry processed: User %s paid ₹%.2f for contest %s", userID, amount, contestID)

	return nil
}