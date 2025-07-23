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
	firebaseAuthHandler *http.FirebaseAuthHandler,
	userHandler *http.UserHandler,
	adminHandler *http.AdminHandler,
	contestHandler *http.ContestHandler,
	paymentHandler *http.PaymentHandler,
	phonePeHandler *http.PhonePeHandler,
	analyticsHandler *http.AnalyticsHandler,
	matchSimulationHandler *http.MatchSimulationHandler,
	autoContestHandler *http.AutoContestHandler,
	wsHandler *ws.WebSocketHandler,
	adminEnhancedHandler *http.AdminEnhancedHandler,
	userEnhancedHandler *http.UserEnhancedHandler,
	cfg *config.Config,
) {
	// Health check endpoint with enhanced info
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "esports-fantasy-backend",
			"version":   "2.0.0",
			"features": []string{
				"phonepe_payments",
				"firebase_auth",
				"auto_contest_management",
				"analytics_dashboard",
				"real_time_simulation",
			},
			"dummy_mode": cfg.Dummy,
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Firebase Configuration (public)
	v1.GET("/firebase/config", firebaseAuthHandler.GetFirebaseConfig)

	// Authentication routes
	auth := v1.Group("/auth")
	{
		// Legacy OTP auth (deprecated but kept for backward compatibility)
		auth.POST("/send-otp", authHandler.SendOTP)
		auth.POST("/verify-otp", authHandler.VerifyOTP)
		
		// Firebase Auth (recommended)
		firebase := auth.Group("/firebase")
		{
			firebase.POST("/send-otp", firebaseAuthHandler.SendOTP)
			firebase.POST("/verify-otp", firebaseAuthHandler.VerifyOTP)
		}
	}

	// WebSocket endpoints
	ws := v1.Group("/ws")
	{
		ws.GET("/leaderboard", wsHandler.HandleWebSocket)
		ws.GET("/match/:matchId/live", matchSimulationHandler.WebSocketHandler)
	}

	// Public contest information
	public := v1.Group("")
	{
		public.GET("/contests/match/:matchId", contestHandler.GetContestsByMatch)
		public.GET("/contests/:id", contestHandler.GetContestDetails)
		public.GET("/contests/:id/leaderboard", contestHandler.GetLeaderboard)
		
		// Public analytics
		public.GET("/analytics/match/:matchId", analyticsHandler.GetMatchAnalytics)
	}

	// Payment callback (public for gateway callbacks)
	payment := v1.Group("/payment")
	{
		payment.POST("/phonepe/callback", phonePeHandler.HandleCallback)
	}

	// Protected routes (authentication required)
	protected := v1.Group("")
	// protected.Use(middleware.AuthRequired(authService)) // TODO: Enable when auth is properly wired
	{
		// User profile routes
		user := protected.Group("/user")
		{
			user.GET("/profile", firebaseAuthHandler.GetProfile)
			user.PUT("/profile", userHandler.UpdateProfile)
			user.GET("/wallet", userHandler.GetWalletBalance)
			user.GET("/analytics", analyticsHandler.GetUserAnalytics) // Own analytics
		}

		// Fantasy team routes
		fantasy := protected.Group("/fantasy")
		{
			fantasy.POST("/teams", contestHandler.CreateFantasyTeam)
			fantasy.GET("/teams", contestHandler.GetUserTeams)
		}

		// Payment routes
		paymentProtected := protected.Group("/payment")
		{
			// Legacy Razorpay (deprecated)
			paymentProtected.POST("/create-order", paymentHandler.CreatePaymentOrder)
			paymentProtected.POST("/success", paymentHandler.HandlePaymentSuccess)
			
			// PhonePe (recommended)
			phonepe := paymentProtected.Group("/phonepe")
			{
				phonepe.POST("/initiate", phonePeHandler.InitiatePayment)
				phonepe.GET("/status/:txnId", phonePeHandler.CheckPaymentStatus)
			}
		}

		// Match simulation for users
		simulation := protected.Group("/simulation")
		{
			simulation.GET("/match/:matchId/events", matchSimulationHandler.GetMatchEvents)
		}
	}

	// Admin routes (authentication + admin role required)
	admin := v1.Group("/admin")
	// admin.Use(middleware.AuthRequired(authService)) // TODO: Enable when auth is properly wired
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

		// User management
		admin.PUT("/users/:userId/promote", firebaseAuthHandler.PromoteToAdmin)

		// Analytics dashboard
		analytics := admin.Group("/analytics")
		{
			analytics.GET("/dashboard", analyticsHandler.GetDashboardStats)
			analytics.GET("/users/:userId", analyticsHandler.GetUserAnalytics)
			analytics.GET("/matches/:matchId", analyticsHandler.GetMatchAnalytics)
			analytics.POST("/refresh", analyticsHandler.RefreshAnalytics)
		}

		// Match simulation management
		simAdmin := admin.Group("/simulation")
		{
			simAdmin.POST("/match/:matchId/start", matchSimulationHandler.StartSimulation)
			simAdmin.POST("/match/:matchId/stop", matchSimulationHandler.StopSimulation)
			simAdmin.GET("/active", matchSimulationHandler.GetActiveSimulations)
		}

		// Auto contest management
		autoContest := admin.Group("/auto-contest")
		{
			autoContest.GET("/status", autoContestHandler.GetSchedulerStatus)
			autoContest.POST("/contests/:contestId/distribute-prizes", autoContestHandler.ForceDistributePrizes)
			autoContest.POST("/contests/:contestId/lock", autoContestHandler.ForceLockContest)
		}
	}
}