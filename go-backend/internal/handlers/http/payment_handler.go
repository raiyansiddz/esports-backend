package http

import (
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService services.PaymentService
}

func NewPaymentHandler(paymentService services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreatePaymentOrder godoc
// @Summary Create payment order
// @Description Create a Razorpay payment order for wallet top-up
// @Tags payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order body models.CreatePaymentOrderRequest true "Payment order data"
// @Success 200 {object} map[string]interface{}
// @Router /payment/create-order [post]
func (h *PaymentHandler) CreatePaymentOrder(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	userModel := user.(*models.User)

	var req models.CreatePaymentOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Currency == "" {
		req.Currency = "INR"
	}

	order, err := h.paymentService.CreatePaymentOrder(userModel.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order": order,
		"key":   "rzp_test_dummy_key_id", // Dummy key for testing
		"message": "Payment order created successfully",
		"note":    "This is a test payment order with dummy Razorpay integration",
	})
}

// HandlePaymentSuccess godoc
// @Summary Handle payment success
// @Description Handle successful payment callback
// @Tags payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment body map[string]interface{} true "Payment success data"
// @Success 200 {object} map[string]string
// @Router /payment/success [post]
func (h *PaymentHandler) HandlePaymentSuccess(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	userModel := user.(*models.User)

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	paymentID, ok := req["razorpay_payment_id"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment ID is required"})
		return
	}

	orderID, ok := req["razorpay_order_id"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	// For dummy payment, assume ₹100 
	amount := 100.0
	if amountVal, exists := req["amount"]; exists {
		if amountFloat, ok := amountVal.(float64); ok {
			amount = amountFloat
		}
	}

	if err := h.paymentService.HandlePaymentSuccess(paymentID, orderID, userModel.ID, amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment processed successfully",
		"amount":  amount,
		"status":  "success",
	})
}