package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"example.com/ginhello/auth"
	"example.com/ginhello/config"
	"example.com/ginhello/handlers"
	"example.com/ginhello/middleware"
)

// SetupRouter configures the Gin router with all routes and middleware
func SetupRouter(cfg *config.Config, logger *zap.Logger) *gin.Engine {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	// Initialize JWT service
	jwtService := auth.NewJWTService(cfg, logger)

	// Initialize auth handler
	authHandler := handlers.NewAuthHandler(jwtService, logger)

	r := gin.New()
	r.Use(middleware.ZapLogger(logger))
	r.Use(gin.Recovery())

	// Public routes
	api := r.Group("/api")
	{
		// Health check endpoint
		api.GET("/healthcheck", handlers.HealthCheck)

		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}
	}

	// Protected routes
	protected := api.Group("/")
	protected.Use(middleware.JWTAuthMiddleware(jwtService, logger))
	{
		// User endpoints
		users := protected.Group("/users")
		{
			users.GET("", handlers.GetUsers)
			users.GET("/:id", handlers.GetUserByID)
			users.POST("", handlers.CreateUser)
		}
	}

	return r
}
