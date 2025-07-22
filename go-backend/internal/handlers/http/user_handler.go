package http

import (
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update authenticated user's profile information
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body map[string]string true "Profile data"
// @Success 200 {object} map[string]string
// @Router /user/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	userModel := user.(*models.User)

	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	name := req["name"]
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	if err := h.userService.UpdateProfile(userModel.ID, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// GetWalletBalance godoc
// @Summary Get user wallet balance
// @Description Get the current wallet balance for authenticated user
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]float64
// @Router /user/wallet [get]
func (h *UserHandler) GetWalletBalance(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	userModel := user.(*models.User)

	balance, err := h.userService.GetWalletBalance(userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get wallet balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": balance,
		"user_id": userModel.ID,
	})
}