package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLoad_WithDefaultValues(t *testing.T) {
	// Setup
	logger := zap.NewNop()

	// Clear environment variables to test defaults
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("JWT_ACCESS_EXPIRY")
	os.Unsetenv("JWT_REFRESH_EXPIRY")
	os.Unsetenv("JWT_ISSUER")

	// Test
	cfg, err := Load(logger)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "default_secret_key_change_this", cfg.JWTSecret)
	assert.Equal(t, 15*time.Minute, cfg.JWTAccessExpiry)
	assert.Equal(t, 72*time.Hour, cfg.JWTRefreshExpiry)
	assert.Equal(t, "ginhello", cfg.JWTIssuer)
}

func TestLoad_WithEnvironmentValues(t *testing.T) {
	// Setup
	logger := zap.NewNop()

	// Set environment variables
	os.Setenv("JWT_SECRET", "custom_secret")
	os.Setenv("JWT_ACCESS_EXPIRY", "30m")
	os.Setenv("JWT_REFRESH_EXPIRY", "48h")
	os.Setenv("JWT_ISSUER", "custom_issuer")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_ACCESS_EXPIRY")
		os.Unsetenv("JWT_REFRESH_EXPIRY")
		os.Unsetenv("JWT_ISSUER")
	}()

	// Test
	cfg, err := Load(logger)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "custom_secret", cfg.JWTSecret)
	assert.Equal(t, 30*time.Minute, cfg.JWTAccessExpiry)
	assert.Equal(t, 48*time.Hour, cfg.JWTRefreshExpiry)
	assert.Equal(t, "custom_issuer", cfg.JWTIssuer)
}

func TestLoad_WithInvalidDurations(t *testing.T) {
	// Setup
	logger := zap.NewNop()

	// Set invalid duration
	os.Setenv("JWT_ACCESS_EXPIRY", "invalid")
	os.Setenv("JWT_REFRESH_EXPIRY", "also-invalid")
	defer func() {
		os.Unsetenv("JWT_ACCESS_EXPIRY")
		os.Unsetenv("JWT_REFRESH_EXPIRY")
	}()

	// Test
	cfg, err := Load(logger)

	// Assert - should use defaults for invalid durations
	assert.NoError(t, err)
	assert.Equal(t, 15*time.Minute, cfg.JWTAccessExpiry)
	assert.Equal(t, 72*time.Hour, cfg.JWTRefreshExpiry)
}

func TestGetEnv(t *testing.T) {
	// Test with existing environment variable
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	result := getEnv("TEST_KEY", "default_value")
	assert.Equal(t, "test_value", result)

	// Test with non-existent environment variable
	result = getEnv("NON_EXISTENT_KEY", "default_value")
	assert.Equal(t, "default_value", result)
}
