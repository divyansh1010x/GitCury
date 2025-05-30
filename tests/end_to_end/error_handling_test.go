package end_to_end

import (
	"github.com/lakshyajain-0291/GitCury/tests/testutils"
	"github.com/lakshyajain-0291/GitCury/utils"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestErrorHandlingAndRecovery(t *testing.T) {
	// Set up test environment
	env, err := testutils.SetupTestEnv()
	if err != nil {
		t.Fatalf("Failed to set up test environment: %v", err)
	}
	defer env.Cleanup()

	// Test Git error recovery using mock operations
	t.Run("SafeGitOperation", func(t *testing.T) {
		// Configure the git mock to simulate recovery scenario
		env.GitMock.SetupMockCommandError("add test.txt", errors.New("fatal: Unable to create '.git/index.lock': File exists"))

		// Test that the mock can simulate recovery
		_, err := env.GitMock.RunGitCmd(env.TempDir, nil, "add", "test.txt")

		// The mock should initially fail with the lock error
		if err == nil {
			t.Error("Expected mock to initially fail with lock error")
		}

		// Clear the error to simulate recovery
		env.GitMock.ClearMockErrors()

		// Now the operation should succeed
		_, err = env.GitMock.RunGitCmd(env.TempDir, nil, "add", "test.txt")
		if err != nil {
			t.Errorf("Expected operation to succeed after clearing errors, but got: %v", err)
		}
	})

	// Test structured error creation and retrieval
	t.Run("StructuredErrors", func(t *testing.T) {
		// Create a structured error
		originalErr := errors.New("original error")
		structErr := utils.NewGitError(
			"Git operation failed",
			originalErr,
			map[string]interface{}{
				"directory": env.TempDir,
				"command":   "commit",
				"files":     []string{"file1.txt", "file2.txt"},
			},
			"file1.txt",
		)

		// Verify error properties - check that it contains expected components
		errorMsg := structErr.Error()
		if !strings.Contains(errorMsg, "[GIT]") {
			t.Errorf("Expected error to contain '[GIT]' prefix, got: %v", errorMsg)
		}
		if !strings.Contains(errorMsg, "Git operation failed: original error") {
			t.Errorf("Expected error to contain 'Git operation failed: original error', got: %v", errorMsg)
		}
		if !strings.Contains(errorMsg, "[File: file1.txt]") {
			t.Errorf("Expected error to contain '[File: file1.txt]', got: %v", errorMsg)
		}

		// Cast to structured error to access additional fields
		gitErr := structErr
		if gitErr == nil {
			t.Fatalf("Expected *utils.StructuredError, got nil")
		}

		// Check properties
		if gitErr.Message != "Git operation failed" {
			t.Errorf("Expected message 'Git operation failed', got %q", gitErr.Message)
		}

		if gitErr.Cause.Error() != "original error" {
			t.Errorf("Expected cause 'original error', got %q", gitErr.Cause.Error())
		}

		if gitErr.ProcessedFile != "file1.txt" {
			t.Errorf("Expected processed file 'file1.txt', got %q", gitErr.ProcessedFile)
		}

		// Check context values
		directory, ok := gitErr.Context["directory"].(string)
		if !ok || directory != env.TempDir {
			t.Errorf("Expected directory %q, got %v", env.TempDir, gitErr.Context["directory"])
		}

		command, ok := gitErr.Context["command"].(string)
		if !ok || command != "commit" {
			t.Errorf("Expected command 'commit', got %v", gitErr.Context["command"])
		}
	})

	// Test retry mechanism
	t.Run("RetryMechanism", func(t *testing.T) {
		// Setup a counter to track retry attempts
		attemptCount := 0

		// Configure retry settings
		retryConfig := utils.RetryConfig{
			MaxRetries:   3,
			InitialDelay: 10,  // milliseconds
			MaxDelay:     50,  // milliseconds
			Factor:       1.5, // exponential backoff factor
		}

		// Create a function that fails twice then succeeds
		testFunc := func() error {
			attemptCount++
			if attemptCount < 3 {
				return errors.New("temporary error")
			}
			return nil
		}

		// Execute with retry
		ctx := context.Background()
		err := utils.WithRetry(ctx, "test_operation", retryConfig, testFunc)

		// Should succeed after retries
		if err != nil {
			t.Errorf("Expected retry to succeed, but got error: %v", err)
		}

		// Should have attempted exactly 3 times
		if attemptCount != 3 {
			t.Errorf("Expected 3 attempts, got %d", attemptCount)
		}

		// Reset counter
		attemptCount = 0

		// Test with a permanent failure
		permanentFunc := func() error {
			attemptCount++
			return errors.New("permanent error")
		}

		// Execute with retry
		err = utils.WithRetry(ctx, "permanent_test_operation", retryConfig, permanentFunc)

		// Should fail after all retries
		if err == nil {
			t.Errorf("Expected retry to fail, but it succeeded")
		}

		// Should have attempted exactly MaxRetries+1 times
		if attemptCount != retryConfig.MaxRetries+1 {
			t.Errorf("Expected %d attempts, got %d", retryConfig.MaxRetries+1, attemptCount)
		}
	})
}
