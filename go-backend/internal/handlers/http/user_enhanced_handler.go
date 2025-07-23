package http

import (
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/services"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserEnhancedHandler struct {
	userService     services.UserService
	usernameService services.UsernameService
}

func NewUserEnhancedHandler(userService services.UserService, usernameService services.UsernameService) *UserEnhancedHandler {
	return &UserEnhancedHandler{
		userService:     userService,
		usernameService: usernameService,
	}
}

// GenerateUsername godoc
// @Summary Generate username for user
// @Description User generates a username after verification
// @Tags user-enhanced
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UsernameGenerationRequest true "Username generation data"
// @Success 200 {object} map[string]string
// @Router /user/generate-username [post]
func (h *UserEnhancedHandler) GenerateUsername(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	userModel := user.(*models.User)

	var req models.UsernameGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Generate username
	username, err := h.usernameService.GenerateUsername(req.PrefixID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user's username
	if err := h.usernameService.UpdateUsername(userModel.ID, username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update username"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Username generated successfully",
		"username": username,
	})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update authenticated user's profile information
// @Tags user-enhanced
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body models.UpdateProfileRequest true "Profile data"
// @Success 200 {object} map[string]string
// @Router /user/profile [put]
func (h *UserEnhancedHandler) UpdateProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	userModel := user.(*models.User)

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate and update username if it's different
	if userModel.Username != req.Username {
		if err := h.usernameService.UpdateUsername(userModel.ID, req.Username); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Validate profile image format if provided
	if req.ProfileImage != "" {
		if err := h.validateProfileImage(req.ProfileImage); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Update profile through user service
	if err := h.userService.UpdateUserProfile(userModel.ID, req.Name, req.Username, req.ProfileImage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get authenticated user's complete profile
// @Tags user-enhanced
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Router /user/profile [get]
func (h *UserEnhancedHandler) GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	userModel := user.(*models.User)

	// Get complete user profile
	profile, err := h.userService.GetProfile(userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// CheckUsernameAvailability godoc
// @Summary Check username availability
// @Description Check if a username is available
// @Tags user-enhanced
// @Produce json
// @Param username query string true "Username to check"
// @Success 200 {object} map[string]bool
// @Router /user/check-username [get]
func (h *UserEnhancedHandler) CheckUsernameAvailability(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username parameter is required"})
		return
	}

	isAvailable := h.usernameService.IsUsernameAvailable(username)

	c.JSON(http.StatusOK, gin.H{
		"username":  username,
		"available": isAvailable,
	})
}

// GetUsernamePrefixes godoc
// @Summary Get available username prefixes
// @Description Get all active username prefixes for selection
// @Tags user-enhanced
// @Produce json
// @Success 200 {array} models.UsernamePrefix
// @Router /user/username-prefixes [get]
func (h *UserEnhancedHandler) GetUsernamePrefixes(c *gin.Context) {
	prefixes, err := h.usernameService.GetUsernamePrefixes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get prefixes"})
		return
	}

	// Filter only active prefixes for users
	var activePrefixes []models.UsernamePrefix
	for _, prefix := range prefixes {
		if prefix.IsActive {
			activePrefixes = append(activePrefixes, prefix)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"prefixes": activePrefixes,
		"count":    len(activePrefixes),
	})
}

// UploadProfileImage godoc
// @Summary Upload profile image
// @Description Upload profile image as base64 (PNG, JPEG, JPG supported)
// @Tags user-enhanced
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param image body map[string]string true "Base64 image data"
// @Success 200 {object} map[string]string
// @Router /user/upload-image [post]
func (h *UserEnhancedHandler) UploadProfileImage(c *gin.Context) {
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

	imageData, exists := req["image"]
	if !exists || imageData == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image data is required"})
		return
	}

	// Validate image format
	if err := h.validateProfileImage(imageData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update profile image
	if err := h.userService.UpdateProfileImage(userModel.ID, imageData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile image uploaded successfully"})
}

// Helper function to validate profile image
func (h *UserEnhancedHandler) validateProfileImage(imageData string) error {
	if imageData == "" {
		return nil // Empty is allowed
	}

	// Check if it's a valid base64 data URL
	if !strings.HasPrefix(imageData, "data:image/") {
		return fmt.Errorf("invalid image format: must be a base64 data URL")
	}

	// Check supported formats
	supportedFormats := []string{"data:image/png;base64,", "data:image/jpeg;base64,", "data:image/jpg;base64,"}
	isValidFormat := false
	for _, format := range supportedFormats {
		if strings.HasPrefix(imageData, format) {
			isValidFormat = true
			break
		}
	}

	if !isValidFormat {
		return fmt.Errorf("unsupported image format: only PNG, JPEG, and JPG are supported")
	}

	// Check file size (base64 encoded data is ~33% larger than original)
	// Limiting to ~2MB original size (2.5MB base64)
	if len(imageData) > 2500000 {
		return fmt.Errorf("image size too large: maximum 2MB allowed")
	}

	return nil
}