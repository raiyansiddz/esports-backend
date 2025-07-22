package routes

import (
	"esports-fantasy-backend/config"
	"esports-fantasy-backend/internal/handlers/http"
	"esports-fantasy-backend/internal/handlers/ws"
	"esports-fantasy-backend/internal/middleware"
	"esports-fantasy-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	router *gin.Engine,
	authHandler *http.AuthHandler,
	userHandler *http.UserHandler,
	adminHandler *http.AdminHandler,
	contestHandler *http.ContestHandler,
	paymentHandler *http.PaymentHandler,
	wsHandler *ws.WebSocketHandler,
	cfg *config.Config,
) {
	// Create auth service for middleware (this would be better injected)
	// For now, we'll assume it's available through dependency injection
	
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "esports-fantasy-backend",
			"version": "1.0.0",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Public routes (no authentication required)
	auth := v1.Group("/auth")
	{
		auth.POST("/send-otp", authHandler.SendOTP)
		auth.POST("/verify-otp", authHandler.VerifyOTP)
	}

	// WebSocket endpoint (no auth for simplicity in MVP)
	v1.GET("/ws", wsHandler.HandleWebSocket)

	// Public contest information
	public := v1.Group("")
	{
		public.GET("/contests/match/:matchId", contestHandler.GetContestsByMatch)
		public.GET("/contests/:id", contestHandler.GetContestDetails)
		public.GET("/contests/:id/leaderboard", contestHandler.GetLeaderboard)
		// Add more public endpoints as needed
	}

	// Protected routes (authentication required)
	protected := v1.Group("")
	// protected.Use(middleware.AuthRequired(authService)) // TODO: Fix dependency injection
	{
		// User profile routes
		user := protected.Group("/user")
		{
			user.GET("/profile", authHandler.GetProfile)
			user.PUT("/profile", userHandler.UpdateProfile)
			user.GET("/wallet", userHandler.GetWalletBalance)
		}

		// Fantasy team routes
		fantasy := protected.Group("/fantasy")
		{
			fantasy.POST("/teams", contestHandler.CreateFantasyTeam)
			fantasy.GET("/teams", contestHandler.GetUserTeams)
		}

		// Payment routes
		payment := protected.Group("/payment")
		{
			payment.POST("/create-order", paymentHandler.CreatePaymentOrder)
			payment.POST("/success", paymentHandler.HandlePaymentSuccess)
		}
	}

	// Admin routes (authentication + admin role required)
	admin := v1.Group("/admin")
	// admin.Use(middleware.AuthRequired(authService)) // TODO: Fix dependency injection
	// admin.Use(middleware.AdminRequired())
	{
		// Tournament management
		admin.POST("/tournaments", adminHandler.CreateTournament)
		admin.GET("/tournaments", adminHandler.GetTournaments)

		// Match management
		admin.POST("/matches", adminHandler.CreateMatch)
		admin.GET("/matches", adminHandler.GetMatches)
		admin.PUT("/matches/:id/status", adminHandler.UpdateMatchStatus)

		// Contest management
		admin.POST("/contests", adminHandler.CreateContest)

		// Player management
		admin.POST("/players", adminHandler.CreatePlayer)
		admin.POST("/esports-teams", adminHandler.CreateESportsTeam)
		admin.GET("/esports-teams", adminHandler.GetESportsTeams)

		// Stats management
		admin.PUT("/stats/match/:matchId/player/:playerId", adminHandler.UpdatePlayerStats)
	}
}

// SetupAuthenticatedRoutes sets up routes with authentication
func SetupAuthenticatedRoutes(
	router *gin.Engine,
	authService services.AuthService,
	authHandler *http.AuthHandler,
	userHandler *http.UserHandler,
	adminHandler *http.AdminHandler,
	contestHandler *http.ContestHandler,
	paymentHandler *http.PaymentHandler,
	wsHandler *ws.WebSocketHandler,
) {
	// API v1 routes
	v1 := router.Group("/api/v1")

	// Protected routes (authentication required)
	protected := v1.Group("")
	protected.Use(middleware.AuthRequired(authService))
	{
		// User profile routes
		user := protected.Group("/user")
		{
			user.GET("/profile", authHandler.GetProfile)
			user.PUT("/profile", userHandler.UpdateProfile)
			user.GET("/wallet", userHandler.GetWalletBalance)
		}

		// Fantasy team routes
		fantasy := protected.Group("/fantasy")
		{
			fantasy.POST("/teams", contestHandler.CreateFantasyTeam)
			fantasy.GET("/teams", contestHandler.GetUserTeams)
		}

		// Payment routes
		payment := protected.Group("/payment")
		{
			payment.POST("/create-order", paymentHandler.CreatePaymentOrder)
			payment.POST("/success", paymentHandler.HandlePaymentSuccess)
		}
	}

	// Admin routes (authentication + admin role required)
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthRequired(authService))
	admin.Use(middleware.AdminRequired())
	{
		// Tournament management
		admin.POST("/tournaments", adminHandler.CreateTournament)
		admin.GET("/tournaments", adminHandler.GetTournaments)

		// Match management
		admin.POST("/matches", adminHandler.CreateMatch)
		admin.GET("/matches", adminHandler.GetMatches)
		admin.PUT("/matches/:id/status", adminHandler.UpdateMatchStatus)

		// Contest management
		admin.POST("/contests", adminHandler.CreateContest)

		// Player management
		admin.POST("/players", adminHandler.CreatePlayer)
		admin.POST("/esports-teams", adminHandler.CreateESportsTeam)
		admin.GET("/esports-teams", adminHandler.GetESportsTeams)

		// Stats management
		admin.PUT("/stats/match/:matchId/player/:playerId", adminHandler.UpdatePlayerStats)
	}
}