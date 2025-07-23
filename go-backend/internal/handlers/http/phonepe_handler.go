package http

import (
        "net/http"

        "esports-fantasy-backend/internal/services"

        "github.com/gin-gonic/gin"
)

type PhonePeHandler struct {
        phonePeService *services.PhonePeService
}

func NewPhonePeHandler(phonePeService *services.PhonePeService) *PhonePeHandler {
        return &PhonePeHandler{
                phonePeService: phonePeService,
        }
}

type PaymentRequest struct {
        Amount    float64 `json:"amount" binding:"required,gt=0"`
        ContestID string  `json:"contest_id,omitempty"`
        Purpose   string  `json:"purpose" binding:"required"`
}

// InitiatePayment initiates PhonePe payment
func (h *PhonePeHandler) InitiatePayment(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
                c.JSON(http.StatusUnauthorized, gin.H{
                        "success": false,
                        "message": "Authentication required",
                })
                return
        }

        var req PaymentRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Invalid request format",
                        "error":   err.Error(),
                })
                return
        }

        response, err := h.phonePeService.InitiatePayment(userID, req.Amount, req.ContestID, req.Purpose)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{
                        "success": false,
                        "message": "Failed to initiate payment",
                        "error":   err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, response)
}

// HandleCallback handles PhonePe payment callback
func (h *PhonePeHandler) HandleCallback(c *gin.Context) {
        base64Response := c.PostForm("response")
        checksum := c.GetHeader("X-VERIFY")

        if base64Response == "" || checksum == "" {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Missing response or checksum",
                })
                return
        }

        if err := h.phonePeService.HandleCallback(base64Response, checksum); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Failed to process callback",
                        "error":   err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Payment callback processed successfully",
        })
}

// CheckPaymentStatus checks payment status
func (h *PhonePeHandler) CheckPaymentStatus(c *gin.Context) {
        txnID := c.Param("txnId")
        if txnID == "" {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Transaction ID is required",
                })
                return
        }

        response, err := h.phonePeService.CheckPaymentStatus(txnID)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{
                        "success": false,
                        "message": "Failed to check payment status",
                        "error":   err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, response)
}