// Package api_test tests the API configuration functionality
package api_test

import (
	"GitCury/api"
	"GitCury/tests/testutils"
	"testing"
	"time"
)

// TestAPIConfig tests API configuration functions
func TestAPIConfig(t *testing.T) {
	// Setup test environment
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()

	// Test SetAPIKey
	api.SetAPIKey("test-api-key")

	// Test GetAPIKey
	apiKey := api.GetAPIKey()
	if apiKey != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", apiKey)
	}

	// Test retry configuration
	retryConfig := api.GetRetryConfig()
	if retryConfig.MaxRetries == 0 {
		t.Error("Expected non-zero MaxRetries in default config")
	}

	// Test SetRetryConfig
	newRetryConfig := struct {
		MaxRetries int
		BaseDelay  time.Duration
	}{
		MaxRetries: 5,
		BaseDelay:  time.Second * 2,
	}
	api.SetRetryConfig(newRetryConfig.MaxRetries, newRetryConfig.BaseDelay)

	updatedRetryConfig := api.GetRetryConfig()
	if updatedRetryConfig.MaxRetries != newRetryConfig.MaxRetries {
		t.Errorf("Expected MaxRetries %d, got %d", newRetryConfig.MaxRetries, updatedRetryConfig.MaxRetries)
	}

	// Test concurrency configuration
	api.SetConcurrencyConfig(8, time.Second*30)
	concurrencyConfig := api.GetConcurrencyConfig()
	if concurrencyConfig.MaxConcurrency != 8 {
		t.Errorf("Expected MaxConcurrency 8, got %d", concurrencyConfig.MaxConcurrency)
	}

	// Test LoadConfig (should not panic)
	err := api.LoadConfig()
	if err != nil {
		t.Logf("API LoadConfig returned error: %v", err)
	}

	t.Log("API configuration tests completed successfully")
}

// TestAPIInit tests the API initialization
func TestAPIInit(t *testing.T) {
	// Setup test environment
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()

	// Test Init function
	api.Init()

	// Verify default configuration is loaded
	retryConfig := api.GetRetryConfig()
	if retryConfig.MaxRetries == 0 {
		t.Error("Expected non-zero MaxRetries after Init")
	}

	concurrencyConfig := api.GetConcurrencyConfig()
	if concurrencyConfig.MaxConcurrency == 0 {
		t.Error("Expected non-zero MaxConcurrency after Init")
	}

	t.Log("API initialization tests completed successfully")
}
