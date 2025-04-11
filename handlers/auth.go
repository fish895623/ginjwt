package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"example.com/ginhello/auth"
	"example.com/ginhello/models"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest represents the token refresh request body
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	jwtService *auth.JWTService
	db         *gorm.DB
	logger     *zap.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(jwtService *auth.JWTService, db *gorm.DB, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		jwtService: jwtService,
		db:         db,
		logger:     logger,
	}
}

// Login handles user login and token generation
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by username
	var foundUser models.User
	result := h.db.Where("username = ?", req.Username).First(&foundUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			h.logger.Warn("Login attempt with non-existent user", zap.String("username", req.Username))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		} else {
			h.logger.Error("Database error during login", zap.Error(result.Error))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Compare password hash
	compareErr := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(req.Password))
	if compareErr != nil {
		h.logger.Warn("Failed login attempt (wrong password)", zap.String("username", req.Username))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	tokens, err := h.jwtService.GenerateTokenPair(&foundUser)
	if err != nil {
		h.logger.Error("Failed to generate tokens", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	h.logger.Info("Successful login", zap.String("username", req.Username))
	c.JSON(http.StatusOK, tokens)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid refresh token request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate refresh token and get claims
	claims, err := h.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		h.logger.Warn("Invalid refresh token received", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Fetch user from DB to ensure they still exist
	var user models.User
	result := h.db.First(&user, claims.UserID)
	if result.Error != nil {
		h.logger.Error("User for refresh token not found in DB", zap.Uint("user_id", claims.UserID))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User associated with token not found"})
		return
	}

	// Generate new tokens using the user data from the DB
	newTokens, err := h.jwtService.GenerateTokenPair(&user)
	if err != nil {
		h.logger.Error("Failed to refresh tokens", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh tokens"})
		return
	}

	h.logger.Info("Token refreshed successfully", zap.Uint("user_id", claims.UserID))
	c.JSON(http.StatusOK, newTokens)
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
