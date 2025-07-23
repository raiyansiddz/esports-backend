package services

import (
        "crypto/hmac"
        "crypto/sha256"
        "encoding/base64"
        "encoding/hex"
        "encoding/json"
        "errors"
        "fmt"
        "log"
        "net/http"
        "strings"
        "time"

        "esports-fantasy-backend/config"
        "esports-fantasy-backend/internal/models"
        "esports-fantasy-backend/internal/repository"

        "github.com/google/uuid"
)

type PhonePeService struct {
        cfg         *config.Config
        userRepo    repository.UserRepository
        txnRepo     repository.TransactionRepository
        contestRepo repository.ContestRepository
}

type PhonePePaymentRequest struct {
        MerchantID          string `json:"merchantId"`
        MerchantTransactionID string `json:"merchantTransactionId"`
        Amount              int64  `json:"amount"` // Amount in paise
        MerchantUserID      string `json:"merchantUserId"`
        RedirectURL         string `json:"redirectUrl"`
        RedirectMode        string `json:"redirectMode"`
        CallbackURL         string `json:"callbackUrl"`
        MobileNumber        string `json:"mobileNumber,omitempty"`
        PaymentInstrument   struct {
                Type string `json:"type"`
        } `json:"paymentInstrument"`
}

type PhonePeRequest struct {
        Request string `json:"request"`
}

type PhonePeResponse struct {
        Success bool        `json:"success"`
        Code    string      `json:"code"`
        Message string      `json:"message"`
        Data    interface{} `json:"data"`
}

type PhonePePaymentData struct {
        MerchantID            string `json:"merchantId"`
        MerchantTransactionID string `json:"merchantTransactionId"`
        TransactionID         string `json:"transactionId"`
        Amount                int64  `json:"amount"`
        State                 string `json:"state"`
        ResponseCode          string `json:"responseCode"`
        PaymentInstrument     struct {
                Type string `json:"type"`
        } `json:"paymentInstrument"`
}

func NewPhonePeService(cfg *config.Config, userRepo repository.UserRepository, txnRepo repository.TransactionRepository, contestRepo repository.ContestRepository) *PhonePeService {
        return &PhonePeService{
                cfg:         cfg,
                userRepo:    userRepo,
                txnRepo:     txnRepo,
                contestRepo: contestRepo,
        }
}

func (s *PhonePeService) InitiatePayment(userID string, amount float64, contestID, purpose string) (*PhonePeResponse, error) {
        // Create transaction record
        txnID := uuid.New().String()
        
        transaction := &models.Transaction{
                ID:             txnID,
                UserID:         userID,
                Amount:         amount,
                Type:           "DEPOSIT",
                Status:         "PENDING",
                PaymentGateway: "PHONEPE",
                PaymentID:      "",
                Description:    purpose,
                ContestID:      contestID,
                CreatedAt:      time.Now(),
                UpdatedAt:      time.Now(),
        }

        if err := s.txnRepo.Create(transaction); err != nil {
                return nil, fmt.Errorf("failed to create transaction: %w", err)
        }

        // Get user details
        user, err := s.userRepo.GetByID(userID)
        if err != nil {
                return nil, fmt.Errorf("failed to get user: %w", err)
        }

        // Create PhonePe payment request
        paymentReq := PhonePePaymentRequest{
                MerchantID:            s.cfg.PhonePeMerchantID,
                MerchantTransactionID: txnID,
                Amount:                int64(amount * 100), // Convert to paise
                MerchantUserID:        userID,
                RedirectURL:           s.cfg.PhonePeRedirectURL,
                RedirectMode:          "REDIRECT",
                CallbackURL:           s.cfg.PhonePeCallbackURL,
                MobileNumber:          user.PhoneNumber,
        }
        paymentReq.PaymentInstrument.Type = "PAY_PAGE"

        // Convert to JSON and encode in base64
        paymentJSON, err := json.Marshal(paymentReq)
        if err != nil {
                return nil, fmt.Errorf("failed to marshal payment request: %w", err)
        }

        base64Payload := base64.StdEncoding.EncodeToString(paymentJSON)
        
        // Create checksum
        checksum := s.createChecksum(base64Payload, "/pg/v1/pay")

        // Log for development (dummy mode)
        if s.cfg.Dummy {
                log.Printf("🔥 PHONEPE PAYMENT INITIATED (DUMMY MODE)")
                log.Printf("📱 Transaction ID: %s", txnID)
                log.Printf("💰 Amount: ₹%.2f", amount)
                log.Printf("👤 User: %s", user.PhoneNumber)
                log.Printf("🎯 Purpose: %s", purpose)
                log.Printf("🔗 Redirect URL: %s", s.cfg.PhonePeRedirectURL)
                log.Printf("===================================")

                // In dummy mode, return a mock response
                return &PhonePeResponse{
                        Success: true,
                        Code:    "PAYMENT_INITIATED",
                        Message: "Payment initiated successfully (DUMMY MODE)",
                        Data: map[string]interface{}{
                                "merchantId":            s.cfg.PhonePeMerchantID,
                                "merchantTransactionId": txnID,
                                "instrumentResponse": map[string]interface{}{
                                        "type": "PAY_PAGE",
                                        "redirectInfo": map[string]interface{}{
                                                "url":    fmt.Sprintf("%s/payment/test?txnId=%s&amount=%.2f", s.cfg.PhonePeRedirectURL, txnID, amount),
                                                "method": "GET",
                                        },
                                },
                        },
                }, nil
        }

        // Make actual API call to PhonePe (production mode)
        return s.makePhonePeAPICall(base64Payload, checksum, "/pg/v1/pay")
}

func (s *PhonePeService) HandleCallback(base64Response, checksum string) error {
        // Verify checksum
        if !s.verifyChecksum(base64Response, "/pg/v1/status", checksum) {
                return errors.New("invalid checksum")
        }

        // Decode response
        responseBytes, err := base64.StdEncoding.DecodeString(base64Response)
        if err != nil {
                return fmt.Errorf("failed to decode response: %w", err)
        }

        var callbackData PhonePePaymentData
        if err := json.Unmarshal(responseBytes, &callbackData); err != nil {
                return fmt.Errorf("failed to unmarshal callback data: %w", err)
        }

        // Get transaction
        transaction, err := s.txnRepo.GetByID(callbackData.MerchantTransactionID)
        if err != nil {
                return fmt.Errorf("transaction not found: %w", err)
        }

        // Update transaction based on payment state
        transaction.PaymentID = callbackData.TransactionID
        transaction.UpdatedAt = time.Now()

        switch callbackData.State {
        case "COMPLETED":
                transaction.Status = "SUCCESS"
                
                // Update user wallet
                user, err := s.userRepo.GetByID(transaction.UserID)
                if err != nil {
                        return fmt.Errorf("failed to get user: %w", err)
                }
                
                user.WalletBalance += transaction.Amount
                if err := s.userRepo.Update(user); err != nil {
                        return fmt.Errorf("failed to update user wallet: %w", err)
                }

                // If this was for contest entry, handle contest entry
                if transaction.ContestID != "" {
                        s.handleContestEntry(transaction)
                }

                log.Printf("✅ Payment successful for transaction %s, amount: ₹%.2f", transaction.ID, transaction.Amount)
                
        case "FAILED":
                transaction.Status = "FAILED"
                log.Printf("❌ Payment failed for transaction %s", transaction.ID)
                
        default:
                transaction.Status = "PENDING"
        }

        return s.txnRepo.Update(transaction)
}

func (s *PhonePeService) CheckPaymentStatus(txnID string) (*PhonePeResponse, error) {
        if s.cfg.Dummy {
                // In dummy mode, simulate successful payment after some time
                transaction, err := s.txnRepo.GetByID(txnID)
                if err != nil {
                        return nil, fmt.Errorf("transaction not found: %w", err)
                }

                // Simulate processing time
                if time.Since(transaction.CreatedAt) > 10*time.Second {
                        // Auto-complete the payment in dummy mode
                        transaction.Status = "SUCCESS"
                        transaction.PaymentID = "dummy_" + txnID
                        transaction.UpdatedAt = time.Now()
                        
                        // Update user wallet
                        user, err := s.userRepo.GetByID(transaction.UserID)
                        if err == nil {
                                user.WalletBalance += transaction.Amount
                                s.userRepo.Update(user)
                        }
                        
                        if err := s.txnRepo.Update(transaction); err == nil {
                                log.Printf("🤖 DUMMY: Auto-completed payment for transaction %s", txnID)
                        }
                }

                return &PhonePeResponse{
                        Success: true,
                        Code:    "PAYMENT_SUCCESS",
                        Message: "Payment status retrieved (DUMMY MODE)",
                        Data: map[string]interface{}{
                                "merchantId":            s.cfg.PhonePeMerchantID,
                                "merchantTransactionId": txnID,
                                "transactionId":         "dummy_" + txnID,
                                "amount":                int64(transaction.Amount * 100),
                                "state":                 strings.ToUpper(transaction.Status),
                                "responseCode":          "SUCCESS",
                        },
                }, nil
        }

        // Create checksum for status check
        endpoint := fmt.Sprintf("/pg/v1/status/%s/%s", s.cfg.PhonePeMerchantID, txnID)
        checksum := s.createChecksum("", endpoint)

        // Make actual API call to PhonePe
        return s.makePhonePeAPICall("", checksum, endpoint)
}

func (s *PhonePeService) createChecksum(payload, endpoint string) string {
        data := payload + endpoint + s.cfg.PhonePeSaltKey
        hash := hmac.New(sha256.New, []byte(s.cfg.PhonePeSaltKey))
        hash.Write([]byte(data))
        return hex.EncodeToString(hash.Sum(nil)) + "###" + fmt.Sprintf("%d", s.cfg.PhonePeSaltIndex)
}

func (s *PhonePeService) verifyChecksum(payload, endpoint, receivedChecksum string) bool {
        expectedChecksum := s.createChecksum(payload, endpoint)
        return expectedChecksum == receivedChecksum
}

func (s *PhonePeService) makePhonePeAPICall(payload, checksum, endpoint string) (*PhonePeResponse, error) {
        url := s.cfg.PhonePeBaseURL + endpoint
        
        var reqBody interface{}
        if payload != "" {
                reqBody = PhonePeRequest{Request: payload}
        }

        reqJSON, err := json.Marshal(reqBody)
        if err != nil {
                return nil, fmt.Errorf("failed to marshal request: %w", err)
        }

        req, err := http.NewRequest("POST", url, strings.NewReader(string(reqJSON)))
        if err != nil {
                return nil, fmt.Errorf("failed to create request: %w", err)
        }

        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-VERIFY", checksum)
        req.Header.Set("X-MERCHANT-ID", s.cfg.PhonePeMerchantID)

        client := &http.Client{Timeout: 30 * time.Second}
        resp, err := client.Do(req)
        if err != nil {
                return nil, fmt.Errorf("failed to make request: %w", err)
        }
        defer resp.Body.Close()

        var phonePeResp PhonePeResponse
        if err := json.NewDecoder(resp.Body).Decode(&phonePeResp); err != nil {
                return nil, fmt.Errorf("failed to decode response: %w", err)
        }

        return &phonePeResp, nil
}

func (s *PhonePeService) handleContestEntry(transaction *models.Transaction) error {
        // Logic to handle contest entry after successful payment
        // This would involve creating a contest entry record
        log.Printf("🎮 Processing contest entry for transaction %s, contest %s", transaction.ID, transaction.ContestID)
        return nil
}