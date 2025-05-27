package api

import (
	"os"
	"sync"
)

var (
	// APIConfig holds common settings for API interactions
	APIConfig struct {
		MaxRetries    int
		RetryDelay    int
		GeminiAPIKey  string
		MaxConcurrent int
		Timeout       int
	}

	// configMutex protects concurrent access to the API configuration
	configMutex sync.RWMutex
)

// SetRetryConfig sets the retry configuration for API calls
func SetRetryConfig(maxRetries, retryDelay int) {
	configMutex.Lock()
	defer configMutex.Unlock()

	// Ensure we have sensible defaults
	if maxRetries <= 0 {
		maxRetries = 3 // Default to 3 retries
	}

	if retryDelay <= 0 {
		retryDelay = 5 // Default to 5 seconds
	}

	APIConfig.MaxRetries = maxRetries
	APIConfig.RetryDelay = retryDelay
}

// SetAPIKey sets the Gemini API key
func SetAPIKey(apiKey string) {
	configMutex.Lock()
	defer configMutex.Unlock()
	APIConfig.GeminiAPIKey = apiKey
}

// GetAPIKey returns the Gemini API key
func GetAPIKey() string {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return APIConfig.GeminiAPIKey
}

// GetRetryConfig returns the current max retries and retry delay settings
func GetRetryConfig() (int, int) {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return APIConfig.MaxRetries, APIConfig.RetryDelay
}

// SetConcurrencyConfig sets the max concurrent API calls and timeout
func SetConcurrencyConfig(maxConcurrent, timeout int) {
	configMutex.Lock()
	defer configMutex.Unlock()

	// Ensure we have sensible defaults
	if maxConcurrent <= 0 {
		maxConcurrent = 5 // Default to 5 concurrent calls
	}

	if timeout <= 0 {
		timeout = 30 // Default to 30 seconds timeout
	}

	APIConfig.MaxConcurrent = maxConcurrent
	APIConfig.Timeout = timeout
}

// GetConcurrencyConfig returns the current max concurrent calls and timeout settings
func GetConcurrencyConfig() (int, int) {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return APIConfig.MaxConcurrent, APIConfig.Timeout
}

// Init initializes the API configuration with defaults
func Init() {
	SetRetryConfig(3, 5)
	SetConcurrencyConfig(5, 30)
}

func init() {
	// Set default values when the package is imported
	Init()
}

// LoadConfig loads the API configuration from the given settings map
func LoadConfig(settings map[string]interface{}) {
	// Initialize retry configuration
	retriesValue, ok := settings["retries"].(int)
	if !ok {
		// Try to convert from float64 (JSON numbers)
		if retriesFloat, ok := settings["retries"].(float64); ok {
			retriesValue = int(retriesFloat)
		} else {
			// Default if not found or invalid
			retriesValue = 3
		}
	}

	// Initialize timeout configuration
	timeoutValue, ok := settings["timeout"].(int)
	if !ok {
		// Try to convert from float64 (JSON numbers)
		if timeoutFloat, ok := settings["timeout"].(float64); ok {
			timeoutValue = int(timeoutFloat)
		} else {
			// Default if not found or invalid
			timeoutValue = 30
		}
	}

	// Get max concurrent value
	maxConcurrentValue, ok := settings["maxConcurrent"].(int)
	if !ok {
		// Try to convert from float64 (JSON numbers)
		if maxConcurrentFloat, ok := settings["maxConcurrent"].(float64); ok {
			maxConcurrentValue = int(maxConcurrentFloat)
		} else {
			// Default if not found or invalid
			maxConcurrentValue = 5
		}
	}

	// Get API key
	apiKey, _ := settings["GEMINI_API_KEY"].(string)
	if apiKey == "" {
		// Try environment variable
		apiKey = os.Getenv("GEMINI_API_KEY")
	}

	// Update the API configuration
	SetRetryConfig(retriesValue, timeoutValue)
	SetConcurrencyConfig(maxConcurrentValue, timeoutValue)
	SetAPIKey(apiKey)
}
