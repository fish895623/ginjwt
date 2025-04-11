package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"example.com/ginhello/handlers"
	"example.com/ginhello/middleware"
)

// SetupRouter configures the Gin router with all routes and middleware
func SetupRouter(logger *zap.Logger) *gin.Engine {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(middleware.ZapLogger(logger))
	r.Use(gin.Recovery())

	// Define API routes
	api := r.Group("/api")
	{
		// Health check endpoint
		api.GET("/healthcheck", handlers.HealthCheck)

		// User endpoints
		api.GET("/users", handlers.GetUsers)
		api.GET("/users/:id", handlers.GetUserByID)
		api.POST("/users", handlers.CreateUser)
	}

	return r
}
