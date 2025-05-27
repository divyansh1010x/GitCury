package testutils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// CreateTempDir creates a temporary directory for testing
func CreateTempDir(t *testing.T) string {
	tempDir, err := ioutil.TempDir("", "gitcury-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return tempDir
}

// CreateTempFile creates a temporary file with the given content
func CreateTempFile(t *testing.T, dir, prefix, content string) string {
	file, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if content != "" {
		if _, err := file.WriteString(content); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
	}

	if err := file.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return file.Name()
}

// SetupGitRepo initializes a git repository in the given directory
func SetupGitRepo(t *testing.T, dir string) {
	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, cmd := range commands {
		if err := runCommand(dir, cmd...); err != nil {
			t.Fatalf("Failed to run command %v: %v", cmd, err)
		}
	}
}

// AddAndCommitFile adds and commits a file to the repository
func AddAndCommitFile(t *testing.T, repoDir, filename, content, message string) {
	filePath := filepath.Join(repoDir, filename)

	// Create the file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", filename, err)
	}

	commands := [][]string{
		{"git", "add", filename},
		{"git", "commit", "-m", message},
	}

	for _, cmd := range commands {
		if err := runCommand(repoDir, cmd...); err != nil {
			t.Fatalf("Failed to run command %v: %v", cmd, err)
		}
	}
}

// CleanupTestData removes the test directory and any temporary files
func CleanupTestData(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Logf("Warning: Failed to cleanup test directory %s: %v", dir, err)
	}
}

// SetupTestEnvironment sets up environment variables for testing
func SetupTestEnvironment(t *testing.T) func() {
	// Set required environment variables for tests
	originalAPIKey := os.Getenv("GEMINI_API_KEY")
	originalHome := os.Getenv("HOME")
	originalTestMode := os.Getenv("GITCURY_TEST_MODE")

	// Set test values
	os.Setenv("GEMINI_API_KEY", "test-api-key-for-testing")
	os.Setenv("GITCURY_TEST_MODE", "true")

	// Create a temporary home directory for config
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)

	// Create .gitcury directory and config file
	gitcuryDir := filepath.Join(tempDir, ".gitcury")
	err := os.MkdirAll(gitcuryDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .gitcury directory: %v", err)
	}

	// Create a valid config.json
	config := map[string]interface{}{
		"app_name":            "GitCury",
		"version":             "1.0.0",
		"GEMINI_API_KEY":      "test-api-key-for-testing",
		"model":               "gemini-1.5-flash",
		"output_format":       "standard",
		"commit_types":        []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"},
		"custom_instructions": "",
		"max_tokens":          1000,
		"temperature":         0.7,
		"auto_commit":         false,
		"verbose":             false,
		"root_folders":        []string{"."},
		"numFilesToCommit":    5,
		"config_dir":          filepath.Join(tempDir, ".gitcury"),
		"output_file_path":    filepath.Join(tempDir, ".gitcury", "output.json"),
		"editor":              "nano",
		"aliases": map[string]string{
			"commit":  "seal",
			"push":    "deploy",
			"getmsgs": "genesis",
			"output":  "trace",
			"config":  "nexus",
			"setup":   "bootstrap",
			"boom":    "cascade",
		},
	}

	configBytes, _ := json.MarshalIndent(config, "", "  ")
	configPath := filepath.Join(gitcuryDir, "config.json")
	err = os.WriteFile(configPath, configBytes, 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Return cleanup function
	return func() {
		if originalAPIKey == "" {
			os.Unsetenv("GEMINI_API_KEY")
		} else {
			os.Setenv("GEMINI_API_KEY", originalAPIKey)
		}
		if originalHome == "" {
			os.Unsetenv("HOME")
		} else {
			os.Setenv("HOME", originalHome)
		}
		if originalTestMode == "" {
			os.Unsetenv("GITCURY_TEST_MODE")
		} else {
			os.Setenv("GITCURY_TEST_MODE", originalTestMode)
		}
	}
}

// Helper to run commands
func runCommand(dir string, args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	return cmd.Run()
}
