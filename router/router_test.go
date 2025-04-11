package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"example.com/ginhello/config"
)

func TestSetupRouter(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	cfg := &config.Config{
		JWTSecret:        "test_secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 24 * time.Hour,
		JWTIssuer:        "test_issuer",
	}

	// Initialize router
	router := SetupRouter(cfg, logger)

	// Test cases for various endpoints
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "Health check endpoint",
			method:         "GET",
			path:           "/api/healthcheck",
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Login endpoint exists",
			method: "POST",
			path:   "/api/auth/login",
			// Will be 400 Bad Request because we're not providing credentials
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Refresh token endpoint exists",
			method: "POST",
			path:   "/api/auth/refresh",
			// Will be 400 Bad Request because we're not providing a refresh token
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Protected user list endpoint requires auth",
			method:         "GET",
			path:           "/api/users",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Protected user by ID endpoint requires auth",
			method:         "GET",
			path:           "/api/users/1",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Protected create user endpoint requires auth",
			method:         "POST",
			path:           "/api/users",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Assert status
			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}
