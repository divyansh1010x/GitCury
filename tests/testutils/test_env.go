// package testutils provides utilities for setting up test environments
package testutils

import (
	"GitCury/config"
	"GitCury/core"
	"GitCury/di"
	"GitCury/output"
	"GitCury/tests/mock"
	"os"
	"path/filepath"
)

// TestEnv holds references to mocks and test setup
type TestEnv struct {
	TempDir    string              // Temporary directory for test files
	GitMock    *mock.MockGitRunner // Mock git implementation
	GeminiMock *mock.MockGeminiAPI // Mock Gemini API
}

// SetupTestEnv creates a test environment with mocks
func SetupTestEnv() (*TestEnv, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gitcury-test-")
	if err != nil {
		return nil, err
	}

	// Create mock instances
	gitMock := mock.NewMockGitRunner()
	geminiMock := mock.NewMockGeminiAPI()

	// Set test mode environment variable to bypass config validation
	os.Setenv("GITCURY_TEST_MODE", "true")

	// Inject the mock git runner for testing
	core.SetGitRunner(gitMock)

	// Inject the mock Gemini runner for testing
	di.SetGeminiRunner(geminiMock)

	// Reset config for testing
	config.ResetConfig()

	// Configure test settings
	config.Set("app_name", "GitCury-Test")
	config.Set("numFilesToCommit", 5)
	config.Set("root_folders", []string{tempDir})
	config.Set("GEMINI_API_KEY", "test-api-key")

	// Return the test environment
	return &TestEnv{
		TempDir:    tempDir,
		GitMock:    gitMock,
		GeminiMock: geminiMock,
	}, nil
}

// CreateTestFiles creates test files in the temp directory
func (env *TestEnv) CreateTestFiles(files []string) []string {
	var paths []string
	for _, file := range files {
		path := filepath.Join(env.TempDir, file)
		dir := filepath.Dir(path)

		// Create directory if it doesn't exist
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, 0755)
		}

		// Create file with simple content
		content := []byte("Test content for " + file)
		os.WriteFile(path, content, 0644)

		paths = append(paths, path)
	}
	return paths
}

// MockMessages adds mock messages directly to the output system
func (env *TestEnv) MockMessages(files []string, messages map[string]string) {
	for _, file := range files {
		filePath := filepath.Join(env.TempDir, file)
		message := "Default mock message"

		if msg, ok := messages[file]; ok {
			message = msg
		}

		output.Set(filePath, env.TempDir, message)
	}
}

// Cleanup releases resources used by the test environment
func (env *TestEnv) Cleanup() {
	// Remove temporary directory
	os.RemoveAll(env.TempDir)

	// Reset config
	config.ResetConfig()

	// Clear output data
	output.Clear()

	// Reset test mode environment variable
	os.Unsetenv("GITCURY_TEST_MODE")
}
