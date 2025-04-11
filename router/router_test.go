package router_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"example.com/ginhello/auth"
	"example.com/ginhello/router"
	"example.com/ginhello/testutils"
)

// Helper to perform requests with optional auth token
func performRequest(r http.Handler, method, path string, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestSetupRouter_PublicRoutes(t *testing.T) {
	// Setup
	db, logger := testutils.SetupTestDB(t)
	cfg := testutils.SetupTestConfig(t)
	routerEngine := router.SetupRouter(cfg, db, logger)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "Health check",
			method:         "GET",
			path:           "/api/healthcheck",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Login endpoint exists (requires body)",
			method:         "POST",
			path:           "/api/auth/login",
			expectedStatus: http.StatusBadRequest, // Expecting bad request without body
		},
		{
			name:           "Refresh endpoint exists (requires body)",
			method:         "POST",
			path:           "/api/auth/refresh",
			expectedStatus: http.StatusBadRequest, // Expecting bad request without body
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := performRequest(routerEngine, tc.method, tc.path, "")
			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestSetupRouter_ProtectedRoutes_Unauthorized(t *testing.T) {
	// Setup
	db, logger := testutils.SetupTestDB(t)
	cfg := testutils.SetupTestConfig(t)
	routerEngine := router.SetupRouter(cfg, db, logger)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "Get users", method: "GET", path: "/api/users"},
		{name: "Get user by ID", method: "GET", path: "/api/users/1"},
		{name: "Create user", method: "POST", path: "/api/users"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := performRequest(routerEngine, tc.method, tc.path, "")
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestSetupRouter_ProtectedRoutes_Authorized(t *testing.T) {
	// Setup
	db, logger := testutils.SetupTestDB(t)
	cfg := testutils.SetupTestConfig(t)
	routerEngine := router.SetupRouter(cfg, db, logger)
	jwtService := auth.NewJWTService(cfg, logger)

	// Create user and generate token
	testUser := testutils.CreateTestUser(t, db, "autheduser", "authed@example.com", "pw")
	tokenPair, err := jwtService.GenerateTokenPair(&testUser)
	assert.NoError(t, err)
	accessToken := tokenPair.AccessToken

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{name: "Get users", method: "GET", path: "/api/users", expectedStatus: http.StatusOK},
		{name: "Get user by ID (self)", method: "GET", path: "/api/users/" + fmt.Sprintf("%d", testUser.ID), expectedStatus: http.StatusOK},
		{name: "Get user by ID (not found)", method: "GET", path: "/api/users/9999", expectedStatus: http.StatusNotFound},
		{name: "Create user (requires body)", method: "POST", path: "/api/users", expectedStatus: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := performRequest(routerEngine, tc.method, tc.path, accessToken)
			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}
