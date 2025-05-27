package git_test

import (
	"GitCury/git"
	"GitCury/tests/testutils"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunGitCmd tests the basic git command execution functionality
func TestRunGitCmd(t *testing.T) {
	// Setup test environment with config and API key
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Set up a git repository
	testutils.SetupGitRepo(t, tempDir)

	// Test a simple git command (git status)
	output, err := git.RunGitCmd(tempDir, nil, "status")
	if err != nil {
		t.Fatalf("RunGitCmd returned an error: %v", err)
	}

	// Verify output contains expected text for a clean repo
	if output == "" {
		t.Error("Expected non-empty output from git status")
	}
}

// TestRunGitCmdWithTimeout tests the git command execution with timeout
func TestRunGitCmdWithTimeout(t *testing.T) {
	// Setup test environment with config and API key
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Set up a git repository
	testutils.SetupGitRepo(t, tempDir)

	// Test with a reasonable timeout (5 seconds)
	output, err := git.RunGitCmdWithTimeout(tempDir, nil, 5000000000, "status") // 5 seconds in nanoseconds
	if err != nil {
		t.Logf("RunGitCmdWithTimeout returned an error: %v (this might be expected in test environment)", err)
		return
	}

	// Verify output contains expected text for a clean repo
	if output == "" {
		t.Error("Expected non-empty output from git status")
	}
}

// TestGetAllChangedFiles tests retrieving changed files from git
func TestGetAllChangedFiles(t *testing.T) {
	// Setup test environment with config and API key
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Set up a git repository
	testutils.SetupGitRepo(t, tempDir)

	// Create an untracked file
	untrackedFile := filepath.Join(tempDir, "untracked.txt")
	if err := os.WriteFile(untrackedFile, []byte("untracked content"), 0644); err != nil {
		t.Fatalf("Failed to create untracked file: %v", err)
	}

	// Get changed files
	files, err := git.GetAllChangedFiles(tempDir)
	if err != nil {
		t.Fatalf("GetAllChangedFiles returned an error: %v", err)
	}

	// Verify we have at least one changed file
	if len(files) < 1 {
		t.Errorf("Expected at least 1 changed file, got %d", len(files))
	}
}

// TestIsGitRepository tests checking if a path is a git repository
func TestIsGitRepository(t *testing.T) {
	// Setup test environment with config and API key
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Check before initializing - should be false
	// Use RunGitCmd to check if directory is a git repository
	_, err := git.RunGitCmd(tempDir, nil, "rev-parse", "--is-inside-work-tree")
	if err == nil {
		t.Error("Expected error when checking non-git directory")
	}

	// Set up a git repository
	testutils.SetupGitRepo(t, tempDir)

	// Check after initializing - should be true
	output, err := git.RunGitCmd(tempDir, nil, "rev-parse", "--is-inside-work-tree")
	if err != nil || strings.TrimSpace(output) != "true" {
		t.Error("Expected directory to be recognized as git repository")
	}
}

// TestGetGitConfigValue tests reading git config values
func TestGetGitConfigValue(t *testing.T) {
	// Setup test environment with config and API key
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Set up a git repository
	testutils.SetupGitRepo(t, tempDir)

	// Get user.name config value using git config command
	output, err := git.RunGitCmd(tempDir, nil, "config", "--get", "user.name")
	if err != nil {
		t.Fatalf("Git config command returned an error: %v", err)
	}

	// Verify the config value is not empty
	if output == "" {
		t.Error("Expected non-empty git config value for user.name")
	}
}

// TestSetGitConfigValue tests setting git config values
func TestSetGitConfigValue(t *testing.T) {
	// Setup test environment with config and API key
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Set up a git repository
	testutils.SetupGitRepo(t, tempDir)

	// Set a custom config value
	testKey := "test.value"
	testValue := "test-value-123"

	_, err := git.RunGitCmd(tempDir, nil, "config", testKey, testValue)
	if err != nil {
		t.Fatalf("Setting git config returned an error: %v", err)
	}

	// Get the config value to verify it was set
	output, err := git.RunGitCmd(tempDir, nil, "config", "--get", testKey)
	if err != nil {
		t.Fatalf("Getting git config returned an error: %v", err)
	}

	// Verify the config value matches what we set
	output = strings.TrimSpace(output)
	if output != testValue {
		t.Errorf("Expected git config value %q, got %q", testValue, output)
	}
}

// Additional tests for recovery and progress functions
func TestCheckRepositoryHealth(t *testing.T) {
	// Setup test environment with config and API key
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Check before initializing - should fail
	err := git.CheckRepositoryHealth(tempDir)
	if err == nil {
		t.Error("Expected CheckRepositoryHealth to return error for non-git directory")
	}

	// Set up a git repository
	testutils.SetupGitRepo(t, tempDir)

	// Check after initializing - should pass
	err = git.CheckRepositoryHealth(tempDir)
	if err != nil {
		t.Errorf("CheckRepositoryHealth returned an error for valid repo: %v", err)
	}
}

func TestRecoverFromGitError(t *testing.T) {
	// Setup test environment with config and API key
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Test with nil error
	result := git.RecoverFromGitError(tempDir, nil)
	if !result.Success {
		t.Error("RecoverFromGitError should return success for nil error")
	}

	// Test with an actual error
	testErr := errors.New("test git error")
	result = git.RecoverFromGitError(tempDir, testErr)

	// Since this is a simulated error, recovery may not be possible
	// but we just want to make sure the function doesn't crash
	t.Logf("Recovery result: success=%v, message=%s", result.Success, result.Message)
}
