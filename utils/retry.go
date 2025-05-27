package utils

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"
)

// SafeExecute runs a function with panic recovery
func SafeExecute(operation string, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			Debug("[RECOVERY]: Recovered from panic in " + operation)
			Debug("[RECOVERY]: Stack trace:\n" + string(debug.Stack()))

			switch x := r.(type) {
			case string:
				err = NewSystemError("Panic occurred", fmt.Errorf("%s", x), map[string]interface{}{
					"operation": operation,
				})
			case error:
				err = NewSystemError("Panic occurred", x, map[string]interface{}{
					"operation": operation,
				})
			default:
				err = NewSystemError("Panic occurred", fmt.Errorf("%v", x), map[string]interface{}{
					"operation": operation,
				})
			}
		}
	}()

	return fn()
}

// RetryConfig holds the configuration for retry operations
type RetryConfig struct {
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Factor       float64 // Exponential backoff factor
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Factor:       2.0,
	}
}

// WithRetry executes a function with automatic retries using exponential backoff
func WithRetry(ctx context.Context, operation string, config RetryConfig, fn func() error) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			Debug(fmt.Sprintf("[RETRY]: Attempt %d/%d for operation '%s' after delay of %v",
				attempt, config.MaxRetries, operation, delay))
		}

		err := SafeExecute(operation, fn)
		if err == nil {
			if attempt > 0 {
				Debug(fmt.Sprintf("[RETRY]: Operation '%s' succeeded after %d attempts",
					operation, attempt+1))
			}
			return nil
		}

		lastErr = err
		Debug(fmt.Sprintf("[RETRY]: Operation '%s' failed (attempt %d/%d): %v",
			operation, attempt+1, config.MaxRetries, err))

		// Don't sleep if this was the last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Check if context is cancelled before sleeping
		select {
		case <-ctx.Done():
			return NewSystemError(
				"Operation cancelled",
				ctx.Err(),
				map[string]interface{}{
					"operation": operation,
					"attempts":  attempt + 1,
				},
			)
		case <-time.After(delay):
			// Calculate next delay with exponential backoff
			delay = time.Duration(float64(delay) * config.Factor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return NewSystemError(
		fmt.Sprintf("Operation failed after %d attempts", config.MaxRetries+1),
		lastErr,
		map[string]interface{}{
			"operation":  operation,
			"maxRetries": config.MaxRetries,
		},
	)
}
