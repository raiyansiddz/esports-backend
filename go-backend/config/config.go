package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL     string
	JWTSecret      string
	RedisURL       string
	Port           string
	GinMode        string
	RazorpayKeyID  string
	RazorpaySecret string
	RazorpayWebhook string
	SMTPHost       string
	SMTPPort       string
	SMTPUser       string
	SMTPPass       string
	OTPConsole     bool
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		JWTSecret:      getEnv("JWT_SECRET", "default-jwt-secret"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		Port:           getEnv("PORT", "8080"),
		GinMode:        getEnv("GIN_MODE", "debug"),
		RazorpayKeyID:  getEnv("RAZORPAY_KEY_ID", ""),
		RazorpaySecret: getEnv("RAZORPAY_SECRET", ""),
		RazorpayWebhook: getEnv("RAZORPAY_WEBHOOK_SECRET", ""),
		SMTPHost:       getEnv("SMTP_HOST", ""),
		SMTPPort:       getEnv("SMTP_PORT", ""),
		SMTPUser:       getEnv("SMTP_USER", ""),
		SMTPPass:       getEnv("SMTP_PASS", ""),
		OTPConsole:     getEnv("OTP_CONSOLE", "true") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}