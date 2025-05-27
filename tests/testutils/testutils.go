package testutils

import (
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

// Helper to run commands
func runCommand(dir string, args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	return cmd.Run()
}
