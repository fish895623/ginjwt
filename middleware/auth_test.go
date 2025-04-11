package middleware

import (
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
	"example.com/ginhello/models"
)

func TestJWTAuthMiddleware(t *testing.T) {
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

	// Create test user and generate token
	user := &models.User{
		ID:       "123",
		Username: "testuser",
		Email:    "test@example.com",
	}
	tokenPair, _ := jwtService.GenerateTokenPair(user)

	// Test cases
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + tokenPair.AccessToken,
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header is required",
		},
		{
			name:           "Invalid authorization format",
			authHeader:     "InvalidFormat " + tokenPair.AccessToken,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header format must be Bearer <token>",
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid.token.string",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("GET", "/test", nil)

			// Set header if needed
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			c.Request = req

			// Set a handler that will be called if middleware passes
			handlerCalled := false
			nextHandler := func(c *gin.Context) {
				handlerCalled = true
				c.Status(http.StatusOK)
			}

			// Create middleware
			middleware := JWTAuthMiddleware(jwtService, logger)

			// Mock c.Next() by explicitly calling the middleware then the next handler if not aborted
			middleware(c)
			if !c.IsAborted() {
				nextHandler(c)
			}

			// Assert
			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedError != "" {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedError, response["error"])
				assert.False(t, handlerCalled)
			} else {
				assert.True(t, handlerCalled)

				// Check that user info was set in context
				userID, exists := c.Get("user_id")
				assert.True(t, exists)
				assert.Equal(t, user.ID, userID)

				username, exists := c.Get("username")
				assert.True(t, exists)
				assert.Equal(t, user.Username, username)

				_, exists = c.Get("token_id")
				assert.True(t, exists)
			}
		})
	}
}
