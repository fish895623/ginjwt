package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"example.com/ginhello/auth"
	"example.com/ginhello/config"
)

func TestAuthHandler_Login(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	cfg := &config.Config{
		JWTSecret:        "test_secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 24 * time.Hour,
		JWTIssuer:        "test_issuer",
	}
	jwtService := auth.NewJWTService(cfg, logger)
	authHandler := NewAuthHandler(jwtService, logger)

	// Test cases
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Valid login",
			requestBody: map[string]interface{}{
				"username": "user1",
				"password": "password1",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "Invalid username",
			requestBody: map[string]interface{}{
				"username": "nonexistent",
				"password": "password1",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "Invalid password",
			requestBody: map[string]interface{}{
				"username": "user1",
				"password": "wrong",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "Missing username",
			requestBody: map[string]interface{}{
				"password": "password1",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Missing password",
			requestBody: map[string]interface{}{
				"username": "user1",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup request
			jsonBody, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Create context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call handler
			authHandler.Login(c)

			// Assert
			assert.Equal(t, tc.expectedStatus, w.Code)

			if !tc.expectedError {
				var response auth.TokenPair
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Greater(t, response.ExpiresIn, int64(0))
			} else {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	cfg := &config.Config{
		JWTSecret:        "test_secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 24 * time.Hour,
		JWTIssuer:        "test_issuer",
	}
	jwtService := auth.NewJWTService(cfg, logger)
	authHandler := NewAuthHandler(jwtService, logger)

	// Generate a valid refresh token
	user := users[0] // using the mock user from handlers/users.go
	tokenPair, _ := jwtService.GenerateTokenPair(&user)

	// Test cases
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Valid refresh token",
			requestBody: map[string]interface{}{
				"refresh_token": tokenPair.RefreshToken,
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "Invalid refresh token",
			requestBody: map[string]interface{}{
				"refresh_token": "invalid.token.here",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:           "Missing refresh token",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup request
			jsonBody, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Create context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call handler
			authHandler.RefreshToken(c)

			// Assert
			assert.Equal(t, tc.expectedStatus, w.Code)

			if !tc.expectedError {
				var response auth.TokenPair
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Greater(t, response.ExpiresIn, int64(0))

				// New token should be different
				assert.NotEqual(t, tokenPair.AccessToken, response.AccessToken)
				assert.NotEqual(t, tokenPair.RefreshToken, response.RefreshToken)
			} else {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}
