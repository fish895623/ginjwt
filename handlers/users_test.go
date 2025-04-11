package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"example.com/ginhello/models"
)

func TestGetUsers(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	// Create request
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Set logger in context as middleware would
	c.Set("logger", logger)

	// Call handler
	GetUsers(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.User
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 3) // We have 3 mock users

	// Check that we get expected user data
	assert.Equal(t, "user1", response[0].Username)
	assert.Equal(t, "user2", response[1].Username)
	assert.Equal(t, "user3", response[2].Username)

	// Verify passwords are not included in response
	for _, user := range response {
		assert.Empty(t, user.Password)
	}
}

func TestGetUserByID(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	// Test cases
	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "Valid user ID",
			userID:         "1",
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "Invalid user ID",
			userID:         "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
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

			// Set logger in context
			c.Set("logger", logger)

			// Call handler
			GetUserByID(c)

			// Assert
			assert.Equal(t, tc.expectedStatus, w.Code)

			if !tc.expectedError {
				var user models.User
				err := json.NewDecoder(w.Body).Decode(&user)
				assert.NoError(t, err)
				assert.Equal(t, tc.userID, user.ID)
			} else {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, "User not found", response["error"])
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	// Test cases
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Valid user",
			requestBody: `{
				"id": "4",
				"username": "newuser",
				"email": "new@example.com",
				"password": "newpassword"
			}`,
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{invalid json`,
			expectedStatus: http.StatusBadRequest,
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

			// Set logger in context
			c.Set("logger", logger)

			// Call handler
			CreateUser(c)

			// Assert status
			assert.Equal(t, tc.expectedStatus, w.Code)

			if !tc.expectedError {
				var user models.User
				err := json.NewDecoder(w.Body).Decode(&user)
				assert.NoError(t, err)
				assert.NotEmpty(t, user.ID)
				assert.NotEmpty(t, user.Username)
				assert.NotEmpty(t, user.Email)
				// Password should not be in the response
				assert.Empty(t, user.Password)
			}
		})
	}
}
