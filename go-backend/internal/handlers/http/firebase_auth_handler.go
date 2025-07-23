package http

import (
        "net/http"

        "esports-fantasy-backend/internal/services"

        "github.com/gin-gonic/gin"
)

type FirebaseAuthHandler struct {
        firebaseAuthService *services.FirebaseAuthService
}

func NewFirebaseAuthHandler(firebaseAuthService *services.FirebaseAuthService) *FirebaseAuthHandler {
        return &FirebaseAuthHandler{
                firebaseAuthService: firebaseAuthService,
        }
}

// GetFirebaseConfig gets Firebase configuration for client-side
func (h *FirebaseAuthHandler) GetFirebaseConfig(c *gin.Context) {
        config := h.firebaseAuthService.GetFirebaseConfig()
        
        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Firebase configuration retrieved",
                "data":    config,
        })
}

// SendOTP sends OTP via Firebase authentication
func (h *FirebaseAuthHandler) SendOTP(c *gin.Context) {
        var req services.OTPRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Invalid request format",
                        "error":   err.Error(),
                })
                return
        }

        response, err := h.firebaseAuthService.SendOTP(req.PhoneNumber)
        if err != nil {
                c.JSON(http.StatusInternalServerError, response)
                return
        }

        c.JSON(http.StatusOK, response)
}

// VerifyOTP verifies OTP and returns JWT token
func (h *FirebaseAuthHandler) VerifyOTP(c *gin.Context) {
        var req services.OTPVerifyRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "Invalid request format",
                        "error":   err.Error(),
                })
                return
        }

        response, err := h.firebaseAuthService.VerifyOTP(req.PhoneNumber, req.OTP)
        if err != nil {
                c.JSON(http.StatusUnauthorized, response)
                return
        }

        c.JSON(http.StatusOK, response)
}

// GetProfile gets user profile
func (h *FirebaseAuthHandler) GetProfile(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
                c.JSON(http.StatusUnauthorized, gin.H{
                        "success": false,
                        "message": "Authentication required",
                })
                return
        }

        profile, err := h.firebaseAuthService.GetUserProfile(userID)
        if err != nil {
                c.JSON(http.StatusNotFound, gin.H{
                        "success": false,
                        "message": "User profile not found",
                        "error":   err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Profile retrieved successfully",
                "data":    profile,
        })
}

// PromoteToAdmin promotes user to admin (admin only)
func (h *FirebaseAuthHandler) PromoteToAdmin(c *gin.Context) {
        targetUserID := c.Param("userId")
        if targetUserID == "" {
                c.JSON(http.StatusBadRequest, gin.H{
                        "success": false,
                        "message": "User ID is required",
                })
                return
        }

        if err := h.firebaseAuthService.PromoteToAdmin(targetUserID); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{
                        "success": false,
                        "message": "Failed to promote user to admin",
                        "error":   err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "User promoted to admin successfully",
        })
}