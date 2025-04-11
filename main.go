package main

import (
	"go.uber.org/zap"

	"example.com/ginhello/config"
	"example.com/ginhello/database"
	"example.com/ginhello/router"
)

func main() {
	// Initialize zap logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load(logger)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Connect to database
	db, err := database.Connect(cfg, logger)
	if err != nil {
		// Error is already logged in Connect
		return
	}

	// Inject DB connection into router setup
	r := router.SetupRouter(cfg, db, logger)

	// Start server
	logger.Info("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
