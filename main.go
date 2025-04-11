package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// User represents a mock user entity
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// ZapLogger creates a gin middleware for logging HTTP requests using Zap
func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Store logger in context for handlers to use
		c.Set("logger", logger)

		// Process request
		c.Next()

		// Log after request is processed
		end := time.Now()
		latency := end.Sub(start)

		logger.Info("request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
			zap.String("error", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}

func main() {
	// Initialize zap logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	// Create a Gin router with no middleware by default
	r := gin.New()

	// Use Zap logger middleware
	r.Use(ZapLogger(logger))

	// Use recovery middleware to handle panics
	r.Use(gin.Recovery())

	// Mock data
	users := []User{
		{ID: "1", Username: "user1", Email: "user1@example.com"},
		{ID: "2", Username: "user2", Email: "user2@example.com"},
		{ID: "3", Username: "user3", Email: "user3@example.com"},
	}

	// Define API routes
	api := r.Group("/api")
	{
		// GET /api/healthcheck - Simple health check endpoint
		api.GET("/healthcheck", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"time":   time.Now().Format(time.RFC3339),
			})
		})

		// GET /api/users - Get all users
		api.GET("/users", func(c *gin.Context) {
			l, _ := c.Get("logger")
			logger := l.(*zap.Logger)
			logger.Info("Fetching all users")

			c.JSON(http.StatusOK, users)
		})

		// GET /api/users/:id - Get user by ID
		api.GET("/users/:id", func(c *gin.Context) {
			id := c.Param("id")
			l, _ := c.Get("logger")
			logger := l.(*zap.Logger)
			logger.Info("Fetching user by ID", zap.String("id", id))

			for _, user := range users {
				if user.ID == id {
					c.JSON(http.StatusOK, user)
					return
				}
			}

			logger.Warn("User not found", zap.String("id", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		})

		// POST /api/users - Create new user (mock)
		api.POST("/users", func(c *gin.Context) {
			l, _ := c.Get("logger")
			logger := l.(*zap.Logger)

			var newUser User
			if err := c.ShouldBindJSON(&newUser); err != nil {
				logger.Error("Invalid user data", zap.Error(err))
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// In a real app, we would save to database here
			logger.Info("Created new user", zap.String("username", newUser.Username))
			c.JSON(http.StatusCreated, newUser)
		})
	}

	// Start server
	logger.Info("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
