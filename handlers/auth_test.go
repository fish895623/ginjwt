package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"example.com/ginhello/auth"
	"example.com/ginhello/handlers"

	// "example.com/ginhello/models" // Removed as not directly used here
	"example.com/ginhello/testutils"
)

func TestAuthHandler_Login(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db, logger := testutils.SetupTestDB(t)
	cfg := testutils.SetupTestConfig(t)
	jwtService := auth.NewJWTService(cfg, logger)
	authHandler := handlers.NewAuthHandler(jwtService, db, logger)

	// Create a test user
	testUser := testutils.CreateTestUser(t, db, "loginuser", "login@example.com", "password123")

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
				"username": testUser.Username,
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "Invalid username",
			requestBody: map[string]interface{}{
				"username": "nonexistent",
				"password": "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "Invalid password",
			requestBody: map[string]interface{}{
				"username": testUser.Username,
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "Missing username",
			requestBody: map[string]interface{}{
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Missing password",
			requestBody: map[string]interface{}{
				"username": testUser.Username,
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
	db, logger := testutils.SetupTestDB(t)
	cfg := testutils.SetupTestConfig(t)
	jwtService := auth.NewJWTService(cfg, logger)
	authHandler := handlers.NewAuthHandler(jwtService, db, logger)

	// Create a test user and generate a refresh token
	testUser := testutils.CreateTestUser(t, db, "refreshuser", "refresh@example.com", "password123")
	tokenPair, _ := jwtService.GenerateTokenPair(&testUser)

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
		{
			name: "Refresh token for non-existent user",
			// Generate token first, then delete user
			requestBody: func() map[string]interface{} {
				tempUser := testutils.CreateTestUser(t, db, "tempuser", "temp@example.com", "pw")
				tempTokenPair, _ := jwtService.GenerateTokenPair(&tempUser)
				db.Delete(&tempUser)
				return map[string]interface{}{"refresh_token": tempTokenPair.RefreshToken}
			}(),
			expectedStatus: http.StatusUnauthorized,
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
