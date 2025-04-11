package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"example.com/ginhello/config"
	"example.com/ginhello/models"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

// TokenPair contains access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Expiry in seconds
}

// TokenClaims contains the claims for JWT
type TokenClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TokenID  string `json:"token_id"` // Used for tracking refresh tokens
	jwt.RegisteredClaims
}

// JWTService handles JWT operations
type JWTService struct {
	config *config.Config
	logger *zap.Logger
}

// NewJWTService creates a new JWT service
func NewJWTService(config *config.Config, logger *zap.Logger) *JWTService {
	return &JWTService{
		config: config,
		logger: logger,
	}
}

// GenerateTokenPair generates an access token and refresh token
func (s *JWTService) GenerateTokenPair(user *models.User) (*TokenPair, error) {
	// Generate tokens with a unique token ID
	tokenID := uuid.NewString()

	// Generate access token
	accessToken, _, err := s.generateToken(user, tokenID, s.config.JWTAccessExpiry)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, _, err := s.generateToken(user, tokenID, s.config.JWTRefreshExpiry)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.JWTAccessExpiry.Seconds()),
	}, nil
}

// ValidateToken validates the JWT token
func (s *JWTService) ValidateToken(tokenString string) (*TokenClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(
		tokenString,
		&TokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(s.config.JWTSecret), nil
		},
	)

	if err != nil {
		// Check if the error is because the token is expired
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// Extract claims
	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshTokens generates new tokens using a refresh token
func (s *JWTService) RefreshTokens(refreshToken string) (*TokenPair, error) {
	// Validate refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Create user object for token generation
	user := &models.User{
		ID:       claims.UserID,
		Username: claims.Username,
	}

	// Generate new token pair with a new token ID
	return s.GenerateTokenPair(user)
}

// Helper to generate a token
func (s *JWTService) generateToken(user *models.User, tokenID string, expiry time.Duration) (string, time.Time, error) {
	// Set expiration time
	expiryTime := time.Now().Add(expiry)

	// Create claims
	claims := &TokenClaims{
		UserID:   user.ID,
		Username: user.Username,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.JWTIssuer,
			Subject:   user.ID,
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		s.logger.Error("Failed to sign token", zap.Error(err))
		return "", time.Time{}, err
	}

	return tokenString, expiryTime, nil
}
