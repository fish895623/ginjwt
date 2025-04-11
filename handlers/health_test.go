package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create request
	req := httptest.NewRequest("GET", "/api/healthcheck", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Call handler
	HealthCheck(c)

	// Assert status
	assert.Equal(t, http.StatusOK, w.Code)

	// Decode response
	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	// Check response fields
	assert.Equal(t, "ok", response["status"])
	assert.NotEmpty(t, response["time"])

	// Validate time format
	timeValue := response["time"]
	parsedTime, err := time.Parse(time.RFC3339, timeValue)
	assert.NoError(t, err)

	// Time should be recent
	assert.WithinDuration(t, time.Now(), parsedTime, 5*time.Second)
}
