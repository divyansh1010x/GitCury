package utils_test

import (
	"GitCury/utils"
	"errors"
	"testing"
	"time"
)

// TestSafeExecute tests the panic recovery mechanism
func TestSafeExecute(t *testing.T) {
	// Test normal execution without panic
	err := utils.SafeExecute("TestOperation", func() error {
		return nil
	})
	if err != nil {
		t.Errorf("SafeExecute returned an error for successful operation: %v", err)
	}

	// Test with returned error
	testErr := errors.New("test error")
	err = utils.SafeExecute("TestOperation", func() error {
		return testErr
	})
	if err != testErr {
		t.Errorf("SafeExecute didn't return the expected error: got %v, want %v", err, testErr)
	}

	// Test with panic
	err = utils.SafeExecute("TestOperation", func() error {
		panic("test panic")
	})
	if err == nil {
		t.Error("SafeExecute didn't return an error for panic")
	}
}

// TestNewWorkerPool tests the worker pool functionality
func TestNewWorkerPool(t *testing.T) {
	// Test with valid number of workers
	pool := utils.NewWorkerPool(3)
	if pool == nil {
		t.Fatal("NewWorkerPool returned nil")
	}

	// Test with invalid number of workers (should default to 1)
	pool = utils.NewWorkerPool(0)
	if pool == nil {
		t.Fatal("NewWorkerPool returned nil for invalid worker count")
	}
}

// TestWorkerPoolSubmit tests submitting tasks to the worker pool
func TestWorkerPoolSubmit(t *testing.T) {
	pool := utils.NewWorkerPool(2)

	// Submit a successful task
	pool.Submit("SuccessTask", 1*time.Second, func() error {
		return nil
	})

	// Submit a task that returns an error
	testErr := errors.New("test error")
	pool.Submit("ErrorTask", 1*time.Second, func() error {
		return testErr
	})

	// Wait for tasks to complete and get errors
	errors := pool.Wait()

	// Verify we got one error
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}
}

// TestErrorHelpers tests the structured error helpers
func TestErrorHelpers(t *testing.T) {
	// Test creating a Git error
	err := utils.NewGitError("git error", nil, nil)
	if err == nil {
		t.Fatal("NewGitError returned nil")
	}

	// Verify error message contains expected text
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}

	// Test with cause
	cause := errors.New("cause error")
	err = utils.NewGitError("git error with cause", cause, nil)

	// Verify unwrapping returns the cause
	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("Expected unwrapped error to be %v, got %v", cause, unwrapped)
	}

	// Test with context
	context := map[string]interface{}{
		"key": "value",
	}
	err = utils.NewGitError("git error with context", nil, context)

	// Verify error is properly formatted
	if err.Error() == "" {
		t.Error("Expected non-empty error message with context")
	}
}

// TestToUserFriendlyMessage tests converting errors to user-friendly messages
func TestToUserFriendlyMessage(t *testing.T) {
	// Test with regular error
	regularErr := errors.New("regular error")
	msg := utils.ToUserFriendlyMessage(regularErr)
	if msg == "" {
		t.Error("Expected non-empty user-friendly message")
	}

	// Test with structured error
	structuredErr := utils.NewGitError("git error", nil, nil)
	msg = utils.ToUserFriendlyMessage(structuredErr)
	if msg == "" {
		t.Error("Expected non-empty user-friendly message for structured error")
	}
}

// TestLogger tests logging functionality
func TestLogger(t *testing.T) {
	// Set debug log level
	originalLevel := utils.LogLevel
	utils.SetLogLevel("debug")

	// Test various log functions
	utils.Debug("Debug message")
	utils.Info("Info message")
	utils.Success("Success message")
	utils.Warning("Warning message")
	utils.Error("Error message")

	// Restore original log level
	utils.SetLogLevel(originalLevel)
}

// TestFileUtils tests file utility functions
func TestFileUtils(t *testing.T) {
	// Test JSON serialization
	data := map[string]string{
		"key": "value",
	}
	json := utils.ToJSON(data)
	if json == "{}" {
		t.Error("Expected non-empty JSON")
	}

	// Test numeric checks
	if !utils.IsNumeric("123") {
		t.Error("Expected '123' to be numeric")
	}

	if utils.IsNumeric("12a3") {
		t.Error("Expected '12a3' to not be numeric")
	}

	// Test parsing
	value, err := utils.ParseInt("123")
	if err != nil {
		t.Errorf("ParseInt returned an error: %v", err)
	}
	if value != 123 {
		t.Errorf("Expected ParseInt to return 123, got %d", value)
	}

	_, err = utils.ParseInt("12a3")
	if err == nil {
		t.Error("Expected ParseInt to return an error for invalid input")
	}
}
