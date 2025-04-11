package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

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
	logger     *zap.Logger
	// In a real app, would have userService or userRepository to get users
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(jwtService *auth.JWTService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		jwtService: jwtService,
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

	// In a real app, would get user from database
	// For this example, we're using mock users
	var foundUser *models.User
	for _, user := range users {
		if user.Username == req.Username {
			foundUser = &user
			break
		}
	}

	if foundUser == nil {
		h.logger.Warn("Login attempt with non-existent user", zap.String("username", req.Username))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// In a real app, would check password hash
	// For this example, mock users have plaintext passwords
	// compareErr := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(req.Password))
	// if compareErr != nil {
	if foundUser.Password != req.Password {
		h.logger.Warn("Failed login attempt", zap.String("username", req.Username))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	tokens, err := h.jwtService.GenerateTokenPair(foundUser)
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

	// Generate new tokens
	tokens, err := h.jwtService.RefreshTokens(req.RefreshToken)
	if err != nil {
		h.logger.Error("Failed to refresh tokens", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	h.logger.Info("Token refreshed successfully")
	c.JSON(http.StatusOK, tokens)
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
