package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestZapLogger(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create a logger that records logs for testing
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	// Setup router with middleware
	r := gin.New()
	r.Use(ZapLogger(logger))

	// Add a test route
	r.GET("/test", func(c *gin.Context) {
		// Verify that logger was set in context
		l, exists := c.Get("logger")
		assert.True(t, exists)
		assert.NotNil(t, l)

		// Use the logger from context
		loggerFromCtx := l.(*zap.Logger)
		loggerFromCtx.Info("Test log from handler")

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create test request
	req := httptest.NewRequest("GET", "/test?param=value", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:1234"

	w := httptest.NewRecorder()

	// Serve request
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Check logs
	logEntries := logs.All()
	assert.GreaterOrEqual(t, len(logEntries), 2) // At least request log and handler log

	// Find request log entry
	var requestLog *observer.LoggedEntry
	for _, entry := range logEntries {
		if entry.Message == "request" {
			requestLog = &entry
			break
		}
	}

	assert.NotNil(t, requestLog)
	if requestLog != nil {
		// Just check that some basic fields exist without asserting specific values
		fields := make(map[string]zap.Field)
		for _, field := range requestLog.Context {
			fields[field.Key] = field
		}

		assert.Contains(t, fields, "status")
		assert.Contains(t, fields, "method")
		assert.Contains(t, fields, "path")
		assert.Contains(t, fields, "query")
		assert.Contains(t, fields, "ip")
		assert.Contains(t, fields, "user-agent")
		assert.Contains(t, fields, "latency")
	}

	// Find handler log entry
	var handlerLog *observer.LoggedEntry
	for _, entry := range logEntries {
		if entry.Message == "Test log from handler" {
			handlerLog = &entry
			break
		}
	}

	assert.NotNil(t, handlerLog)
}
