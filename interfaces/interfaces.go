package interfaces

import (
	"time"
)

// ConfigManager defines the interface for configuration operations
type ConfigManager interface {
	LoadConfig() (interface{}, error)
	SaveConfig(config interface{}) error
	GetDefaultConfigPath() string
}

// APIClient defines the interface for API operations
type APIClient interface {
	GenerateEmbedding(text string) ([]float32, error)
	GenerateCommitMessage(diff string, path string) (string, error)
}

// FileSystem defines the interface for file system operations
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm int) error
	Exists(path string) bool
	IsDir(path string) bool
	MkdirAll(path string, perm int) error
	Remove(path string) error
	RemoveAll(path string) error
	TempDir(dir, pattern string) (string, error)
	TempFile(dir, pattern string) (string, error)
}

// ProgressReporter defines the interface for progress reporting
type ProgressReporter interface {
	Start(total int, message string)
	Increment()
	IncrementWithMessage(message string)
	Stop()
	SetTotal(total int)
}

// Logger defines the interface for logging
type Logger interface {
	Debug(message string, args ...interface{})
	Info(message string, args ...interface{})
	Success(message string, args ...interface{})
	Warning(message string, args ...interface{})
	Error(message string, args ...interface{})
	SetLogLevel(level string)
}

// WorkerPool defines the interface for worker pools
type WorkerPool interface {
	Submit(name string, timeout time.Duration, task func() error)
	Wait() []error
}

// ErrorHandler defines the interface for error handling
type ErrorHandler interface {
	NewError(message string, cause error, context map[string]interface{}) error
	ToUserFriendlyMessage(err error) string
	SafeExecute(operation string, fn func() error) error
}
