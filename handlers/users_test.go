package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"example.com/ginhello/handlers"
	"example.com/ginhello/models"
	"example.com/ginhello/testutils"
)

func TestGetUsers(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db, logger := testutils.SetupTestDB(t)
	userHandler := handlers.NewUserHandler(db, logger)

	// Create some test users
	u1 := testutils.CreateTestUser(t, db, "getuser1", "get1@example.com", "pw1")
	u2 := testutils.CreateTestUser(t, db, "getuser2", "get2@example.com", "pw2")

	// Create request
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Call handler
	userHandler.GetUsers(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.PublicUser
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)

	// Check usernames (order might not be guaranteed)
	foundUsernames := map[string]bool{}
	for _, u := range response {
		foundUsernames[u.Username] = true
	}
	assert.True(t, foundUsernames[u1.Username])
	assert.True(t, foundUsernames[u2.Username])
}

func TestGetUserByID(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db, logger := testutils.SetupTestDB(t)
	userHandler := handlers.NewUserHandler(db, logger)

	// Create a test user
	testUser := testutils.CreateTestUser(t, db, "getbyiduser", "getbyid@example.com", "pw")

	// Test cases
	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		expectedError  bool
		expectedUser   *models.User
	}{
		{
			name:           "Valid user ID",
			userID:         fmt.Sprintf("%d", testUser.ID),
			expectedStatus: http.StatusOK,
			expectedError:  false,
			expectedUser:   &testUser,
		},
		{
			name:           "Invalid user ID (not found)",
			userID:         "9999",
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
			expectedUser:   nil,
		},
		{
			name:           "Invalid user ID format (non-numeric)",
			userID:         "not-a-number",
			expectedStatus: http.StatusNotFound, // Expecting not found as the ID won't parse/match
			expectedError:  true,
			expectedUser:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup request
			req := httptest.NewRequest("GET", "/api/users/"+tc.userID, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = []gin.Param{{Key: "id", Value: tc.userID}}

			// Call handler
			userHandler.GetUserByID(c)

			// Assert
			assert.Equal(t, tc.expectedStatus, w.Code)

			if !tc.expectedError {
				var user models.PublicUser
				err := json.NewDecoder(w.Body).Decode(&user)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUser.ID, user.ID)
				assert.Equal(t, tc.expectedUser.Username, user.Username)
				assert.Equal(t, tc.expectedUser.Email, user.Email)
			} else {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db, logger := testutils.SetupTestDB(t)
	userHandler := handlers.NewUserHandler(db, logger)

	// Create an existing user for conflict test
	_ = testutils.CreateTestUser(t, db, "existinguser", "existing@example.com", "pw")

	// Test cases
	tests := []struct {
		name             string
		requestBody      string
		expectedStatus   int
		expectedError    bool
		expectedUsername string // Only check username on success
	}{
		{
			name: "Valid user",
			requestBody: `{
				"username": "newuser",
				"email": "new@example.com",
				"password": "newpassword"
			}`,
			expectedStatus:   http.StatusCreated,
			expectedError:    false,
			expectedUsername: "newuser",
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{invalid json`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Missing required fields (username)",
			requestBody: `{
				"email": "missing@example.com",
				"password": "password"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Duplicate username",
			requestBody: `{
				"username": "existinguser",
				"email": "duplicate@example.com",
				"password": "password"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  true,
		},
		{
			name: "Duplicate email",
			requestBody: `{
				"username": "anotheruser",
				"email": "existing@example.com",
				"password": "password"
			}`,
			expectedStatus: http.StatusConflict,
			expectedError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup request
			req := httptest.NewRequest("POST", "/api/users", strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call handler
			userHandler.CreateUser(c)

			// Assert status
			assert.Equal(t, tc.expectedStatus, w.Code)

			if !tc.expectedError {
				var user models.PublicUser
				err := json.NewDecoder(w.Body).Decode(&user)
				assert.NoError(t, err)
				assert.Greater(t, user.ID, uint(0))
				assert.Equal(t, tc.expectedUsername, user.Username)
				assert.NotEmpty(t, user.CreatedAt)
			} else {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}
