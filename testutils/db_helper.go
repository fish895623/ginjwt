package testutils

import (
	"testing"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"example.com/ginhello/config"
	"example.com/ginhello/database"
	"example.com/ginhello/models"
)

// SetupTestDB initializes an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) (*gorm.DB, *zap.Logger) {
	t.Helper()

	logger := zap.NewNop()

	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate test database schema: %v", err)
	}

	// Add cleanup function to close DB connection after test
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	})

	return db, logger
}

// SetupTestConfig creates a basic config for testing
func SetupTestConfig(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{
		JWTSecret:        "test_secret",
		JWTAccessExpiry:  config.DefaultJWTAccessExpiry,
		JWTRefreshExpiry: config.DefaultJWTRefreshExpiry,
		JWTIssuer:        "test_issuer",
	}
}

// CreateTestUser creates a user in the test database and returns it
func CreateTestUser(t *testing.T, db *gorm.DB, username, email, password string) models.User {
	t.Helper()

	hashedPassword, err := database.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := models.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
	}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}
