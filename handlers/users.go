package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"example.com/ginhello/models"
)

// UserHandler handles user-related requests
type UserHandler struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(db *gorm.DB, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		db:     db,
		logger: logger,
	}
}

// GetUsers returns all users
func (h *UserHandler) GetUsers(c *gin.Context) {
	h.logger.Info("Fetching all users")

	var users []models.User
	result := h.db.Find(&users)
	if result.Error != nil {
		h.logger.Error("Database error fetching users", zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Convert to public representation
	publicUsers := make([]models.PublicUser, len(users))
	for i, user := range users {
		publicUsers[i] = models.PublicUser{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	c.JSON(http.StatusOK, publicUsers)
}

// GetUserByID returns a user by ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	h.logger.Info("Fetching user by ID", zap.String("id", idStr))

	// Validate that ID is numeric before querying DB
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid user ID format", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var user models.User
	result := h.db.First(&user, uint(id))
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			h.logger.Warn("User not found in DB", zap.Uint64("id", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			h.logger.Error("Database error fetching user by ID", zap.Uint64("id", id), zap.Error(result.Error))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Convert to public representation
	publicUser := models.PublicUser{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, publicUser)
}

// CreateUserRequest represents the body for creating a user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid user creation request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("Failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	// Create user model
	newUser := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	// Save to database
	result := h.db.Create(&newUser)
	if result.Error != nil {
		// Check for unique constraint violation (adjust error message check for broader compatibility)
		errMsg := strings.ToLower(result.Error.Error())
		if strings.Contains(errMsg, "unique constraint") || strings.Contains(errMsg, "duplicate key value") {
			h.logger.Warn("Attempted to create user with existing username or email", zap.String("username", req.Username), zap.String("email", req.Email))
			c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		} else {
			h.logger.Error("Failed to create user in database", zap.Error(result.Error))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		}
		return
	}

	h.logger.Info("Created new user", zap.String("username", newUser.Username), zap.Uint("user_id", newUser.ID))

	// Convert to public representation
	publicUser := models.PublicUser{
		ID:        newUser.ID,
		Username:  newUser.Username,
		Email:     newUser.Email,
		CreatedAt: newUser.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	c.JSON(http.StatusCreated, publicUser)
}
