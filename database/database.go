package database

import (
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"example.com/ginhello/config"
	"example.com/ginhello/models"
)

// Connect initializes the database connection using GORM
func Connect(cfg *config.Config, logger *zap.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DBSource), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
		return nil, err
	}

	logger.Info("Database connection established successfully")

	// Migrate the schema
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		logger.Fatal("Failed to migrate database schema", zap.Error(err))
		return nil, err
	}

	logger.Info("Database schema migrated successfully")
	return db, nil
}
