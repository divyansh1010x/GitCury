package mock

import (
	"GitCury/interfaces"
	"fmt"
	"path/filepath"
	"strings"
)

// MockGeminiAPI simulates the Google Gemini API for testing
type MockGeminiAPI struct {
	// State for testing
	CommitMessages  map[string]string            // map[filePathKey]commitMessage
	DefaultMessage  string                       // Default message if not specifically configured
	ResponseDelay   int                          // Simulated response delay in milliseconds
	ShouldFail      bool                         // Whether API calls should fail
	FailureMessage  string                       // Error message when failing
	LastPrompt      string                       // Last prompt sent to API
	LastContextData map[string]map[string]string // Last context data sent to API
	CallCount       int                          // Number of times the API was called
}

// NewMockGeminiAPI creates a new instance with default testing values
func NewMockGeminiAPI() *MockGeminiAPI {
	return &MockGeminiAPI{
		CommitMessages: make(map[string]string),
		DefaultMessage: "feat: implement new feature",
		ResponseDelay:  0,
		ShouldFail:     false,
		FailureMessage: "mock API error",
		CallCount:      0,
	}
}

// Ensure MockGeminiAPI implements GeminiRunner interface
var _ interfaces.GeminiRunner = (*MockGeminiAPI)(nil)

// SetupMockCommitMessage configures a specific commit message for a file or files
func (m *MockGeminiAPI) SetupMockCommitMessage(filePathKey string, message string) {
	m.CommitMessages[filePathKey] = message
}

// SetupDefaultMessage configures the default commit message
func (m *MockGeminiAPI) SetupDefaultMessage(message string) {
	m.DefaultMessage = message
}

// SetupResponseDelay configures a simulated API response delay
func (m *MockGeminiAPI) SetupResponseDelay(milliseconds int) {
	m.ResponseDelay = milliseconds
}

// SetupShouldFail configures whether API calls should fail
func (m *MockGeminiAPI) SetupShouldFail(shouldFail bool, message string) {
	m.ShouldFail = shouldFail
	if message != "" {
		m.FailureMessage = message
	}
}

// SendToGemini mocks the SendToGemini function for testing
func (m *MockGeminiAPI) SendToGemini(contextData map[string]map[string]string, apiKey string, customInstructions ...string) (string, error) {
	m.CallCount++
	m.LastContextData = contextData

	// Record the last prompt
	if len(customInstructions) > 0 {
		m.LastPrompt = customInstructions[0]
	}

	// Simulate failure if configured
	if m.ShouldFail {
		return "", fmt.Errorf("%s", m.FailureMessage)
	}

	// Check for specific file paths to determine the commit message
	var filesList []string
	for filePath := range contextData {
		filesList = append(filesList, filePath)
	}

	// Sort files to create a consistent key
	filesKey := strings.Join(filesList, "|")

	// Return specific message if configured
	if message, ok := m.CommitMessages[filesKey]; ok {
		return message, nil
	}

	// For single file cases, check individual files
	if len(filesList) == 1 && len(m.CommitMessages) > 0 {
		for filePath, message := range m.CommitMessages {
			if strings.Contains(filesKey, filePath) {
				return message, nil
			}
		}
	}

	// Default message with file info
	fileCount := len(filesList)
	if fileCount == 1 {
		baseName := filepath.Base(filesList[0])
		return fmt.Sprintf("feat: update %s", baseName), nil
	} else {
		return fmt.Sprintf("feat: update %d files", fileCount), nil
	}
}
