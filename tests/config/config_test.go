// Package config_test tests the configuration functionality
package config_test

import (
	"GitCury/config"
	"GitCury/tests/testutils"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadConfig tests loading the configuration
func TestLoadConfig(t *testing.T) {
	// Setup test environment with config and API key
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()

	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Create a sample config file with all required fields
	configPath := filepath.Join(tempDir, "config.json")
	configContent := `{
		"GEMINI_API_KEY": "test-api-key",
		"output_file_path": "test-output.json",
		"logLevel": "debug",
		"maxConcurrent": 4,
		"app_name": "GitCury",
		"version": "1.0.0",
		"root_folders": ["."],
		"numFilesToCommit": 5
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Backup original HOME and config
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set HOME to point to temp directory
	os.Setenv("HOME", tempDir)

	// Create .gitcury directory and move config there
	gitcuryDir := filepath.Join(tempDir, ".gitcury")
	if err := os.MkdirAll(gitcuryDir, 0755); err != nil {
		t.Fatalf("Failed to create .gitcury directory: %v", err)
	}

	finalConfigPath := filepath.Join(gitcuryDir, "config.json")
	if err := os.Rename(configPath, finalConfigPath); err != nil {
		t.Fatalf("Failed to move config file: %v", err)
	}

	// Load the config
	err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned an error: %v", err)
	}

	// Verify config values using Get()
	if apiKey := config.Get("GEMINI_API_KEY"); apiKey != "test-api-key" {
		t.Errorf("Expected API key %q, got %q", "test-api-key", apiKey)
	}

	if outputFile := config.Get("output_file_path"); outputFile != "test-output.json" {
		t.Errorf("Expected output file %q, got %q", "test-output.json", outputFile)
	}

	if logLevel := config.Get("logLevel"); logLevel != "debug" {
		t.Errorf("Expected log level %q, got %q", "debug", logLevel)
	}

	if maxConcurrent := config.Get("maxConcurrent"); maxConcurrent != float64(4) { // JSON numbers are float64
		t.Errorf("Expected max concurrent %v, got %v", 4, maxConcurrent)
	}
}

// TestGetConfigPath tests getting the config directory path
func TestGetConfigPath(t *testing.T) {
	// The config directory should be available
	configDir := config.Get("config_dir")
	if configDir == nil || configDir == "" {
		t.Error("Expected non-empty config directory path")
	}
}

// TestSetConfig tests setting configuration values
func TestSetConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Backup original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set HOME to point to temp directory
	os.Setenv("HOME", tempDir)

	// Create .gitcury directory
	gitcuryDir := filepath.Join(tempDir, ".gitcury")
	if err := os.MkdirAll(gitcuryDir, 0755); err != nil {
		t.Fatalf("Failed to create .gitcury directory: %v", err)
	}

	// Test setting various config values
	testKey := "test_key"
	testValue := "test_value"

	config.Set(testKey, testValue)

	// Verify the value was set
	retrievedValue := config.Get(testKey)
	if retrievedValue != testValue {
		t.Errorf("Expected %q, got %q", testValue, retrievedValue)
	}

	// Test setting numeric value
	numericKey := "numeric_test"
	numericValue := 42

	config.Set(numericKey, numericValue)

	retrievedNumeric := config.Get(numericKey)
	if retrievedNumeric != numericValue {
		t.Errorf("Expected %d, got %v", numericValue, retrievedNumeric)
	}

	// Verify the config file was created
	configPath := filepath.Join(gitcuryDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

// TestLoadConfigDefaults tests that default values are set correctly
func TestLoadConfigDefaults(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Backup original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set HOME to point to temp directory (no existing config)
	os.Setenv("HOME", tempDir)

	// Load the config (should create defaults)
	err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned an error: %v", err)
	}

	// Verify default values are set
	appName := config.Get("app_name")
	if appName != "GitCury" {
		t.Errorf("Expected app_name %q, got %q", "GitCury", appName)
	}

	version := config.Get("version")
	if version != "1.0.0" {
		t.Errorf("Expected version %q, got %q", "1.0.0", version)
	}

	// Check that root_folders has a default value
	rootFolders := config.Get("root_folders")
	if rootFolders == nil {
		t.Error("Expected non-nil root_folders")
	}
}
