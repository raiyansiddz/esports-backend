package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database Configuration
	DatabaseURL string
	
	// Server Configuration
	Port    string
	GinMode string
	
	// Authentication
	JWTSecret string
	
	// Redis Configuration
	RedisURL string
	
	// PhonePe Payment Gateway
	PhonePeMerchantID  string
	PhonePeSaltKey     string
	PhonePeSaltIndex   int
	PhonePeBaseURL     string
	PhonePeRedirectURL string
	PhonePeCallbackURL string
	
	// Firebase Configuration
	FirebaseAPIKey           string
	FirebaseAuthDomain       string
	FirebaseProjectID        string
	FirebaseStorageBucket    string
	FirebaseMessagingSenderID string
	FirebaseAppID            string
	FirebaseMeasurementID    string
	
	// Development Settings
	Dummy      bool
	OTPConsole bool
	
	// Auto Contest Management
	AutoLockEnabled              bool
	AutoPrizeDistributionEnabled bool
	ContestLockMinutesBeforeMatch int
	
	// Analytics Configuration
	AnalyticsEnabled      bool
	AnalyticsRetentionDays int
	
	// Real-time Features
	WebSocketEnabled     bool
	MatchSimulationEnabled bool
	LiveScoringEnabled   bool
	
	// Legacy Razorpay (for backward compatibility)
	RazorpayKeyID       string
	RazorpaySecret      string
	RazorpayWebhook     string
	
	// Email/SMS Configuration
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	
	// Parse integer values
	phonePeSaltIndex, _ := strconv.Atoi(getEnv("PHONEPE_SALT_INDEX", "1"))
	contestLockMinutes, _ := strconv.Atoi(getEnv("CONTEST_LOCK_MINUTES_BEFORE_MATCH", "15"))
	analyticsRetentionDays, _ := strconv.Atoi(getEnv("ANALYTICS_RETENTION_DAYS", "365"))

	return &Config{
		// Database Configuration
		DatabaseURL: getEnv("DATABASE_URL", ""),
		
		// Server Configuration
		Port:    getEnv("PORT", "8001"),
		GinMode: getEnv("GIN_MODE", "debug"),
		
		// Authentication
		JWTSecret: getEnv("JWT_SECRET", "default-jwt-secret"),
		
		// Redis Configuration
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),
		
		// PhonePe Payment Gateway
		PhonePeMerchantID:  getEnv("PHONEPE_MERCHANT_ID", "UATMERCHANT"),
		PhonePeSaltKey:     getEnv("PHONEPE_SALT_KEY", ""),
		PhonePeSaltIndex:   phonePeSaltIndex,
		PhonePeBaseURL:     getEnv("PHONEPE_BASE_URL", "https://mercury-uat.phonepe.com"),
		PhonePeRedirectURL: getEnv("PHONEPE_REDIRECT_URL", "http://localhost:3000/payment/success"),
		PhonePeCallbackURL: getEnv("PHONEPE_CALLBACK_URL", "http://localhost:8001/api/v1/payment/phonepe/callback"),
		
		// Firebase Configuration
		FirebaseAPIKey:            getEnv("FIREBASE_API_KEY", ""),
		FirebaseAuthDomain:        getEnv("FIREBASE_AUTH_DOMAIN", ""),
		FirebaseProjectID:         getEnv("FIREBASE_PROJECT_ID", ""),
		FirebaseStorageBucket:     getEnv("FIREBASE_STORAGE_BUCKET", ""),
		FirebaseMessagingSenderID: getEnv("FIREBASE_MESSAGING_SENDER_ID", ""),
		FirebaseAppID:             getEnv("FIREBASE_APP_ID", ""),
		FirebaseMeasurementID:     getEnv("FIREBASE_MEASUREMENT_ID", ""),
		
		// Development Settings
		Dummy:      getEnv("DUMMY", "true") == "true",
		OTPConsole: getEnv("OTP_CONSOLE", "true") == "true",
		
		// Auto Contest Management
		AutoLockEnabled:              getEnv("AUTO_LOCK_ENABLED", "true") == "true",
		AutoPrizeDistributionEnabled: getEnv("AUTO_PRIZE_DISTRIBUTION_ENABLED", "true") == "true",
		ContestLockMinutesBeforeMatch: contestLockMinutes,
		
		// Analytics Configuration
		AnalyticsEnabled:       getEnv("ANALYTICS_ENABLED", "true") == "true",
		AnalyticsRetentionDays: analyticsRetentionDays,
		
		// Real-time Features
		WebSocketEnabled:       getEnv("WEBSOCKET_ENABLED", "true") == "true",
		MatchSimulationEnabled: getEnv("MATCH_SIMULATION_ENABLED", "true") == "true",
		LiveScoringEnabled:     getEnv("LIVE_SCORING_ENABLED", "true") == "true",
		
		// Legacy Razorpay (for backward compatibility)
		RazorpayKeyID:  getEnv("RAZORPAY_KEY_ID", ""),
		RazorpaySecret: getEnv("RAZORPAY_SECRET", ""),
		RazorpayWebhook: getEnv("RAZORPAY_WEBHOOK_SECRET", ""),
		
		// Email/SMS Configuration
		SMTPHost: getEnv("SMTP_HOST", ""),
		SMTPPort: getEnv("SMTP_PORT", ""),
		SMTPUser: getEnv("SMTP_USER", ""),
		SMTPPass: getEnv("SMTP_PASS", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}