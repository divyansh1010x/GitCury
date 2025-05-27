// Package config_test tests the configuration functionality
package config_test

import (
	"GitCury/config"
	"GitCury/tests/testutils"
	"GitCury/utils"
	"os"
	"path/filepath"
	"testing"
)

// TestCriticalConfigMissing tests that LoadConfig returns an error when critical config is missing
func TestCriticalConfigMissing(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Create a sample config file missing the API key
	configPath := filepath.Join(tempDir, "config.json")
	configContent := `{
		"app_name": "GitCury",
		"version": "1.0.0",
		"root_folders": []
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Backup original HOME and env vars
	originalHome := os.Getenv("HOME")
	originalAPIKey := os.Getenv("GEMINI_API_KEY")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("GEMINI_API_KEY", originalAPIKey)
	}()

	// Clear environment variables that might interfere
	os.Setenv("GEMINI_API_KEY", "")

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

	// Load the config - should return error due to missing API key
	err := config.LoadConfig()

	// Verify error was returned
	if err == nil {
		t.Fatal("Expected an error due to missing API key, but got nil")
	}

	// Verify it's the right type of error
	configErr, ok := err.(*utils.StructuredError)
	if !ok {
		t.Fatalf("Expected *utils.StructuredError, got %T", err)
	}

	if configErr.Type != utils.ConfigError {
		t.Errorf("Expected ConfigError type, got %v", configErr.Type)
	}

	// Check that stop_execution is true in the context
	stopExecution, exists := configErr.Context["stop_execution"].(bool)
	if !exists || !stopExecution {
		t.Errorf("Expected stop_execution to be true in error context")
	}

	// Check that missing_fields includes GEMINI_API_KEY
	missingFields, exists := configErr.Context["missing_fields"].([]string)
	if !exists {
		t.Fatal("Expected missing_fields in error context")
	}

	foundAPIKey := false
	for _, field := range missingFields {
		if field == "GEMINI_API_KEY" {
			foundAPIKey = true
			break
		}
	}

	if !foundAPIKey {
		t.Errorf("Expected GEMINI_API_KEY in missing fields, got %v", missingFields)
	}
}

// TestEmptyRootFolders tests that LoadConfig returns an error when root_folders is empty
func TestEmptyRootFolders(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Create a sample config file with API key but empty root_folders
	configPath := filepath.Join(tempDir, "config.json")
	configContent := `{
		"app_name": "GitCury",
		"version": "1.0.0",
		"GEMINI_API_KEY": "test-api-key",
		"root_folders": []
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Backup original HOME
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

	// Load the config - should return error due to empty root_folders
	err := config.LoadConfig()

	// Verify error was returned
	if err == nil {
		t.Fatal("Expected an error due to empty root_folders, but got nil")
	}

	// Verify it's the right type of error
	configErr, ok := err.(*utils.StructuredError)
	if !ok {
		t.Fatalf("Expected *utils.StructuredError, got %T", err)
	}

	if configErr.Type != utils.ConfigError {
		t.Errorf("Expected ConfigError type, got %v", configErr.Type)
	}

	// Check that stop_execution is true in the context
	stopExecution, exists := configErr.Context["stop_execution"].(bool)
	if !exists || !stopExecution {
		t.Errorf("Expected stop_execution to be true in error context")
	}

	// Check that missing_fields includes root_folders
	missingFields, exists := configErr.Context["missing_fields"].([]string)
	if !exists {
		t.Fatal("Expected missing_fields in error context")
	}

	foundRootFolders := false
	for _, field := range missingFields {
		if field == "root_folders" {
			foundRootFolders = true
			break
		}
	}

	if !foundRootFolders {
		t.Errorf("Expected root_folders in missing fields, got %v", missingFields)
	}
}

// TestAPIKeyFromEnv tests that API key from environment takes precedence
func TestAPIKeyFromEnv(t *testing.T) {
	// Reset the config settings to ensure a clean test
	config.ResetConfig()

	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Create a sample config file without API key
	configPath := filepath.Join(tempDir, "config.json")
	configContent := `{
		"app_name": "GitCury",
		"version": "1.0.0",
		"root_folders": ["."]
	}`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Backup original HOME and API key env var
	originalHome := os.Getenv("HOME")
	originalAPIKey := os.Getenv("GEMINI_API_KEY")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("GEMINI_API_KEY", originalAPIKey)
	}()

	// Set API key in environment
	os.Setenv("GEMINI_API_KEY", "env-api-key")

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

	// Load the config - should succeed with API key from env
	err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned an error: %v", err)
	}

	// For debugging
	t.Logf("Config loaded. Environment API key: %s", os.Getenv("GEMINI_API_KEY"))
	t.Logf("Config value: %v", config.Get("GEMINI_API_KEY"))

	// Verify API key was taken from environment
	apiKey := config.Get("GEMINI_API_KEY")
	if apiKey != "env-api-key" {
		t.Errorf("Expected API key %q from environment, got %q", "env-api-key", apiKey)
	}
}
