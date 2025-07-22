package main

import (
	"context"
	"esports-fantasy-backend/config"
	"esports-fantasy-backend/internal/handlers/http"
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
// @version 1.0
// @description High-performance eSports Fantasy backend with OTP authentication
// @host localhost:8080
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
	tournamentRepo := repository.NewTournamentRepository(db)
	matchRepo := repository.NewMatchRepository(db)
	contestRepo := repository.NewContestRepository(db)
	fantasyTeamRepo := repository.NewFantasyTeamRepository(db)
	playerRepo := repository.NewPlayerRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg)
	userService := services.NewUserService(userRepo, cfg)
	tournamentService := services.NewTournamentService(tournamentRepo)
	matchService := services.NewMatchService(matchRepo)
	contestService := services.NewContestService(contestRepo)
	fantasyTeamService := services.NewFantasyTeamService(fantasyTeamRepo, playerRepo)
	playerService := services.NewPlayerService(playerRepo)
	scoringService := services.NewScoringService(db, rdb)
	paymentService := services.NewPaymentService(transactionRepo, cfg)
	leaderboardService := services.NewLeaderboardService(rdb, fantasyTeamRepo)

	// Initialize handlers
	authHandler := http.NewAuthHandler(authService, userService)
	userHandler := http.NewUserHandler(userService)
	adminHandler := http.NewAdminHandler(tournamentService, matchService, contestService, playerService, scoringService)
	contestHandler := http.NewContestHandler(contestService, fantasyTeamService, leaderboardService)
	paymentHandler := http.NewPaymentHandler(paymentService)

	// Initialize WebSocket hub
	wsHub := ws.NewHub()
	wsHandler := ws.NewWebSocketHandler(wsHub, leaderboardService)

	// Start WebSocket hub
	go wsHub.Run()

	// Initialize router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Swagger documentation
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/redoc", func(c *gin.Context) {
		c.HTML(http.StatusOK, "redoc.html", gin.H{
			"title": "eSports Fantasy API",
		})
	})

	// Setup routes
	routes.SetupRoutes(router, authHandler, userHandler, adminHandler, contestHandler, paymentHandler, wsHandler, cfg)

	// Server configuration
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Port)
		log.Printf("📚 API Documentation available at http://localhost:%s/docs", cfg.Port)
		log.Printf("📖 ReDoc available at http://localhost:%s/redoc", cfg.Port)
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited")
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
	)
}