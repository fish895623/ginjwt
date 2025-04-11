package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"example.com/ginhello/config"
	"example.com/ginhello/models"
)

func TestJWTService_GenerateTokenPair(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	cfg := &config.Config{
		JWTSecret:        "test_secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 24 * time.Hour,
		JWTIssuer:        "test_issuer",
	}
	jwtService := NewJWTService(cfg, logger)
	user := &models.User{
		ID:       "123",
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Test
	tokenPair, err := jwtService.GenerateTokenPair(user)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, int64(cfg.JWTAccessExpiry.Seconds()), tokenPair.ExpiresIn)
}

func TestJWTService_ValidateToken(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	cfg := &config.Config{
		JWTSecret:        "test_secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 24 * time.Hour,
		JWTIssuer:        "test_issuer",
	}
	jwtService := NewJWTService(cfg, logger)
	user := &models.User{
		ID:       "123",
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Generate token
	tokenPair, err := jwtService.GenerateTokenPair(user)
	assert.NoError(t, err)

	// Test valid token
	claims, err := jwtService.ValidateToken(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.NotEmpty(t, claims.TokenID)

	// Test invalid token
	claims, err = jwtService.ValidateToken("invalid.token.string")
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrInvalidToken, err)

	// Test token with wrong signature
	tamperedToken := tokenPair.AccessToken + "tampered"
	claims, err = jwtService.ValidateToken(tamperedToken)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTService_RefreshTokens(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	cfg := &config.Config{
		JWTSecret:        "test_secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 24 * time.Hour,
		JWTIssuer:        "test_issuer",
	}
	jwtService := NewJWTService(cfg, logger)
	user := &models.User{
		ID:       "123",
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Generate initial tokens
	tokenPair, err := jwtService.GenerateTokenPair(user)
	assert.NoError(t, err)

	// Test refresh
	newTokenPair, err := jwtService.RefreshTokens(tokenPair.RefreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newTokenPair.AccessToken)
	assert.NotEmpty(t, newTokenPair.RefreshToken)
	assert.NotEqual(t, tokenPair.AccessToken, newTokenPair.AccessToken)
	assert.NotEqual(t, tokenPair.RefreshToken, newTokenPair.RefreshToken)

	// Test with invalid token
	newTokenPair, err = jwtService.RefreshTokens("invalid.token.string")
	assert.Error(t, err)
	assert.Nil(t, newTokenPair)
}
