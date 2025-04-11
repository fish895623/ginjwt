package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Config holds all configuration for the application
type Config struct {
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration
	JWTIssuer        string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSource   string // Constructed DSN
}

// Load loads configuration from environment variables
func Load(logger *zap.Logger) (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found", zap.Error(err))
	}

	// Parse JWT expiry durations
	accessExpiry, err := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m"))
	if err != nil {
		logger.Error("Invalid JWT_ACCESS_EXPIRY", zap.Error(err))
		accessExpiry = 15 * time.Minute // Default to 15 minutes
	}

	refreshExpiry, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRY", "72h"))
	if err != nil {
		logger.Error("Invalid JWT_REFRESH_EXPIRY", zap.Error(err))
		refreshExpiry = 72 * time.Hour // Default to 72 hours
	}

	// Construct DSN
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "postgres")
	dbSource := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	return &Config{
		JWTSecret:        getEnv("JWT_SECRET", "default_secret_key_change_this"),
		JWTAccessExpiry:  accessExpiry,
		JWTRefreshExpiry: refreshExpiry,
		JWTIssuer:        getEnv("JWT_ISSUER", "ginhello"),

		DBHost:     dbHost,
		DBPort:     dbPort,
		DBUser:     dbUser,
		DBPassword: dbPassword,
		DBName:     dbName,
		DBSource:   dbSource,
	}, nil
}

// Helper to get environment variable with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
