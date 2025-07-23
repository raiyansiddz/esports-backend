package main

import (
	"context"
	"esports-fantasy-backend/config"
	httphandlers "esports-fantasy-backend/internal/handlers/http"
	"esports-fantasy-backend/internal/handlers/ws"
	"esports-fantasy-backend/internal/middleware"
	"esports-fantasy-backend/internal/models"
	"esports-fantasy-backend/internal/repository"
	"esports-fantasy-backend/internal/routes"
	"esports-fantasy-backend/internal/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title eSports Fantasy API
// @version 2.0
// @description Advanced eSports Fantasy backend with PhonePe payments, Firebase auth, auto-management, analytics and real-time features
// @host localhost:8001
// @BasePath /api/v1
func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize database
	db, err := initDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models
	if err := autoMigrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize Redis
	rdb := initRedis(cfg.RedisURL)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	otpRepo := repository.NewOTPRepository(db)
	tournamentRepo := repository.NewTournamentRepository(db)
	matchRepo := repository.NewMatchRepository(db)
	contestRepo := repository.NewContestRepository(db)
	fantasyTeamRepo := repository.NewFantasyTeamRepository(db)
	playerRepo := repository.NewPlayerRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	
	// Initialize enhanced repositories
	usernamePrefixRepo := repository.NewUsernamePrefixRepository(db)
	gameRepo := repository.NewGameRepository(db)
	gameScoringRuleRepo := repository.NewGameScoringRuleRepository(db)

	// Initialize core services
	authService := services.NewAuthService(userRepo, cfg)
	firebaseAuthService := services.NewFirebaseAuthService(cfg, userRepo, otpRepo)
	userService := services.NewUserService(userRepo, cfg)
	tournamentService := services.NewTournamentService(tournamentRepo)
	matchService := services.NewMatchService(matchRepo)
	contestService := services.NewContestService(contestRepo)
	fantasyTeamService := services.NewFantasyTeamService(fantasyTeamRepo, playerRepo)
	playerService := services.NewPlayerService(playerRepo)
	scoringService := services.NewScoringService(db, rdb)
	leaderboardService := services.NewLeaderboardService(rdb, fantasyTeamRepo)
	
	// Initialize enhanced services
	usernameService := services.NewUsernameService(userRepo, usernamePrefixRepo, cfg)
	gameService := services.NewGameService(gameRepo, gameScoringRuleRepo, cfg)

	// Initialize advanced services
	phonePeService := services.NewPhonePeService(cfg, userRepo, transactionRepo, contestRepo)
	paymentService := services.NewPaymentService(transactionRepo, userRepo, cfg)
	analyticsService := services.NewAnalyticsService(cfg, db, rdb, userRepo, contestRepo, transactionRepo, matchRepo)
	matchSimulationService := services.NewMatchSimulationService(cfg, matchRepo, playerRepo, scoringService, leaderboardService, rdb)
	autoContestService := services.NewAutoContestService(cfg, contestRepo, matchRepo, fantasyTeamRepo, transactionRepo, userRepo, leaderboardService)

	// Initialize handlers
	authHandler := httphandlers.NewAuthHandler(authService, userService)
	firebaseAuthHandler := httphandlers.NewFirebaseAuthHandler(firebaseAuthService)
	userHandler := httphandlers.NewUserHandler(userService)
	adminHandler := httphandlers.NewAdminHandler(tournamentService, matchService, contestService, playerService, scoringService)
	contestHandler := httphandlers.NewContestHandler(contestService, fantasyTeamService, leaderboardService)
	paymentHandler := httphandlers.NewPaymentHandler(paymentService)
	phonePeHandler := httphandlers.NewPhonePeHandler(phonePeService)
	analyticsHandler := httphandlers.NewAnalyticsHandler(analyticsService)
	matchSimulationHandler := httphandlers.NewMatchSimulationHandler(matchSimulationService)
	autoContestHandler := httphandlers.NewAutoContestHandler(autoContestService)
	
	// Initialize enhanced handlers
	adminEnhancedHandler := httphandlers.NewAdminEnhancedHandler(usernameService, gameService)
	userEnhancedHandler := httphandlers.NewUserEnhancedHandler(userService, usernameService)

	// Initialize WebSocket hub
	wsHub := ws.NewHub()
	wsHandler := ws.NewWebSocketHandler(wsHub, leaderboardService)

	// Start services
	go wsHub.Run()
	
	// Start auto contest management
	if err := autoContestService.StartScheduler(); err != nil {
		log.Printf("❌ Failed to start auto contest service: %v", err)
	}
	
	// Start analytics caching
	analyticsService.StartAnalyticsCaching()

	// Initialize router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Swagger documentation
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/redoc", func(c *gin.Context) {
		c.HTML(http.StatusOK, "redoc.html", gin.H{
			"title": "eSports Fantasy API v2.0",
		})
	})

	// Setup routes
	routes.SetupRoutes(router, authHandler, firebaseAuthHandler, userHandler, adminHandler, contestHandler, paymentHandler, phonePeHandler, analyticsHandler, matchSimulationHandler, autoContestHandler, wsHandler, adminEnhancedHandler, userEnhancedHandler, cfg)

	// Server configuration
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Enhanced eSports Fantasy Backend v2.0 starting on port %s", cfg.Port)
		log.Printf("📚 API Documentation: http://localhost:%s/docs", cfg.Port)
		log.Printf("📖 ReDoc: http://localhost:%s/redoc", cfg.Port)
		log.Printf("🔥 Features: PhonePe Payments, Firebase Auth, Auto-Management, Analytics, Real-time Simulation")
		log.Printf("💳 Payment Gateway: PhonePe (Dummy Mode: %t)", cfg.Dummy)
		log.Printf("🔐 Authentication: Firebase + JWT (Console OTP: %t)", cfg.OTPConsole)
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Stop auto contest service
	autoContestService.StopScheduler()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited gracefully")
}

func initDatabase(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Enable UUID extension
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
	
	return db, nil
}

func initRedis(redisURL string) *redis.Client {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("Warning: Failed to parse Redis URL, using defaults: %v", err)
		opt = &redis.Options{
			Addr: "localhost:6379",
		}
	}

	rdb := redis.NewClient(opt)
	
	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	}

	return rdb
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.OTP{},
		&models.Tournament{},
		&models.ESportsTeam{},
		&models.Player{},
		&models.Match{},
		&models.Contest{},
		&models.FantasyTeam{},
		&models.FantasyTeamPlayer{},
		&models.PlayerMatchStats{},
		&models.Transaction{},
		// New enhanced models
		&models.UsernamePrefix{},
		&models.Game{},
		&models.GameScoringRule{},
		&models.Achievement{},
		&models.UserAchievement{},
		&models.ContestTemplate{},
		&models.PlayerAnalytics{},
		&models.SeasonLeague{},
	)
}