package utils_test

import (
	"GitCury/utils"
	"fmt"
	"strings"
	"testing"
)

func TestFileContextInErrors(t *testing.T) {
	// Test NewGitError with file information
	testFile := "src/main.go"
	err := utils.NewGitError(
		"Test error with file context",
		fmt.Errorf("underlying error"),
		map[string]interface{}{
			"key": "value",
		},
		testFile,
	)

	// Verify the file information is correctly stored
	if err.ProcessedFile != testFile {
		t.Errorf("Expected ProcessedFile to be %s, got %s", testFile, err.ProcessedFile)
	}

	// Test the error string formatting includes file information
	errorString := err.Error()
	if !strings.Contains(errorString, testFile) {
		t.Errorf("Error string '%s' does not contain expected file information '%s'", errorString, testFile)
	}
}

func TestValidationErrorWithFile(t *testing.T) {
	testFile := "config.json"
	err := utils.NewValidationError(
		"Invalid configuration",
		fmt.Errorf("missing required field"),
		map[string]interface{}{
			"field": "api_key",
		},
		testFile,
	)

	if err.ProcessedFile != testFile {
		t.Errorf("Expected ProcessedFile to be %s, got %s", testFile, err.ProcessedFile)
	}

	errorString := err.Error()
	if !strings.Contains(errorString, testFile) {
		t.Errorf("Validation error string '%s' does not contain file information '%s'", errorString, testFile)
	}
}

func TestAPIErrorWithFile(t *testing.T) {
	testFile := "batch_files.txt"
	err := utils.NewAPIError(
		"API request failed",
		fmt.Errorf("timeout"),
		map[string]interface{}{
			"endpoint": "/api/commit",
		},
		testFile,
	)

	if err.ProcessedFile != testFile {
		t.Errorf("Expected ProcessedFile to be %s, got %s", testFile, err.ProcessedFile)
	}

	errorString := err.Error()
	if !strings.Contains(errorString, testFile) {
		t.Errorf("API error string '%s' does not contain file information '%s'", errorString, testFile)
	}
}
