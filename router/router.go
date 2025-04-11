package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"example.com/ginhello/auth"
	"example.com/ginhello/config"
	"example.com/ginhello/handlers"
	"example.com/ginhello/middleware"
)

// SetupRouter configures the Gin router with all routes and middleware
func SetupRouter(cfg *config.Config, db *gorm.DB, logger *zap.Logger) *gin.Engine {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	// Initialize JWT service
	jwtService := auth.NewJWTService(cfg, logger)

	// Initialize handlers with DB dependency
	authHandler := handlers.NewAuthHandler(jwtService, db, logger)
	userHandler := handlers.NewUserHandler(db, logger)

	r := gin.New()
	r.Use(middleware.ZapLogger(logger))
	r.Use(gin.Recovery())

	// Add Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

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
	protected.POST("/users", userHandler.CreateUser)
	protected.Use(middleware.JWTAuthMiddleware(jwtService, logger))
	{
		// User endpoints
		users := protected.Group("/users")
		{
			users.GET("", userHandler.GetUsers)
			users.GET("/:id", userHandler.GetUserByID)
		}
	}

	return r
}
