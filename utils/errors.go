package utils

import (
	"fmt"
	"strings"
)

// ErrorType represents the category of an error
type ErrorType string

const (
	// Error categories
	ConfigError     ErrorType = "CONFIG"
	GitError        ErrorType = "GIT"
	APIError        ErrorType = "API"
	ValidationError ErrorType = "VALIDATION"
	SystemError     ErrorType = "SYSTEM"
	UserError       ErrorType = "USER"
)

// StructuredError represents an error with additional context
type StructuredError struct {
	Type          ErrorType
	Message       string
	Cause         error
	Context       map[string]interface{}
	ProcessedFile string
}

// Error implements the error interface
func (e *StructuredError) Error() string {
	msg := fmt.Sprintf("[%s] %s", e.Type, e.Message)
	if e.Cause != nil {
		msg += fmt.Sprintf(": %s", e.Cause.Error())
	}

	if len(e.Context) > 0 {
		contextStrs := make([]string, 0, len(e.Context))
		for k, v := range e.Context {
			contextStrs = append(contextStrs, fmt.Sprintf("%s=%v", k, v))
		}
		msg += fmt.Sprintf(" [%s]", strings.Join(contextStrs, ", "))
	}

	if e.ProcessedFile != "" {
		msg += fmt.Sprintf(" [File: %s]", e.ProcessedFile)
	}

	return msg
}

// Unwrap implements the errors.Unwrap interface
func (e *StructuredError) Unwrap() error {
	return e.Cause
}

// Helper functions to create different types of errors

// NewConfigError creates a new configuration error
func NewConfigError(message string, cause error, context map[string]interface{}, processedFile ...string) *StructuredError {
	var file string
	if len(processedFile) > 0 {
		file = processedFile[0]
	}
	return &StructuredError{
		Type:          ConfigError,
		Message:       message,
		Cause:         cause,
		Context:       context,
		ProcessedFile: file,
	}
}

// NewGitError creates a new Git-related error
func NewGitError(message string, cause error, context map[string]interface{}, processedFile ...string) *StructuredError {
	var file string
	if len(processedFile) > 0 {
		file = processedFile[0]
	}
	return &StructuredError{
		Type:          GitError,
		Message:       message,
		Cause:         cause,
		Context:       context,
		ProcessedFile: file,
	}
}

// NewAPIError creates a new API-related error
func NewAPIError(message string, cause error, context map[string]interface{}, processedFile ...string) *StructuredError {
	var file string
	if len(processedFile) > 0 {
		file = processedFile[0]
	}
	return &StructuredError{
		Type:          APIError,
		Message:       message,
		Cause:         cause,
		Context:       context,
		ProcessedFile: file,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string, cause error, context map[string]interface{}, processedFile ...string) *StructuredError {
	var file string
	if len(processedFile) > 0 {
		file = processedFile[0]
	}
	return &StructuredError{
		Type:          ValidationError,
		Message:       message,
		Cause:         cause,
		Context:       context,
		ProcessedFile: file,
	}
}

// NewSystemError creates a new system-related error
func NewSystemError(message string, cause error, context map[string]interface{}, processedFile ...string) *StructuredError {
	var file string
	if len(processedFile) > 0 {
		file = processedFile[0]
	}
	return &StructuredError{
		Type:          SystemError,
		Message:       message,
		Cause:         cause,
		Context:       context,
		ProcessedFile: file,
	}
}

// NewUserError creates a new user-related error
func NewUserError(message string, cause error, context map[string]interface{}) *StructuredError {
	return &StructuredError{
		Type:    UserError,
		Message: message,
		Cause:   cause,
		Context: context,
	}
}

// ToUserFriendlyMessage converts an error to a user-friendly message with possible solution
func ToUserFriendlyMessage(err error) string {
	if err == nil {
		return ""
	}

	// Try to cast to StructuredError
	if structured, ok := err.(*StructuredError); ok {
		switch structured.Type {
		case ConfigError:
			return fmt.Sprintf("Configuration issue: %s\nSuggestion: Check your configuration file or run 'gitcury setup' to reconfigure.", structured.Message)
		case GitError:
			return fmt.Sprintf("Git operation failed: %s\nSuggestion: Verify that Git is installed and that you have the necessary permissions.", structured.Message)
		case APIError:
			return fmt.Sprintf("API connection issue: %s\nSuggestion: Check your internet connection and API key configuration.", structured.Message)
		case ValidationError:
			return fmt.Sprintf("Invalid input: %s\nSuggestion: Review the command syntax and parameters.", structured.Message)
		case SystemError:
			return fmt.Sprintf("System error: %s\nSuggestion: Verify that you have the necessary permissions and system resources.", structured.Message)
		case UserError:
			return fmt.Sprintf("User error: %s", structured.Message)
		default:
			return fmt.Sprintf("Error: %s", err.Error())
		}
	}

	// Default error handling
	return fmt.Sprintf("Error: %s", err.Error())
}
