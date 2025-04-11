package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"example.com/ginhello/models"
)

// Mock data - in a real app, this would come from a database
var users = []models.User{
	{ID: "1", Username: "user1", Email: "user1@example.com"},
	{ID: "2", Username: "user2", Email: "user2@example.com"},
	{ID: "3", Username: "user3", Email: "user3@example.com"},
}

// GetUsers returns all users
func GetUsers(c *gin.Context) {
	l, _ := c.Get("logger")
	logger := l.(*zap.Logger)
	logger.Info("Fetching all users")

	c.JSON(http.StatusOK, users)
}

// GetUserByID returns a user by ID
func GetUserByID(c *gin.Context) {
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
}

// CreateUser creates a new user
func CreateUser(c *gin.Context) {
	l, _ := c.Get("logger")
	logger := l.(*zap.Logger)

	var newUser models.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		logger.Error("Invalid user data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In a real app, we would save to database here
	logger.Info("Created new user", zap.String("username", newUser.Username))
	c.JSON(http.StatusCreated, newUser)
}
