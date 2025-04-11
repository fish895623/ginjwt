package main

import (
	"go.uber.org/zap"

	"example.com/ginhello/router"
)

func main() {
	// Initialize zap logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Setup router with all routes and middleware
	r := router.SetupRouter(logger)

	// Start server
	logger.Info("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
