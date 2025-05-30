package end_to_end

import (
	"github.com/lakshyajain-0291/gitcury/core"
	"github.com/lakshyajain-0291/gitcury/output"
	"github.com/lakshyajain-0291/gitcury/tests/testutils"
	"testing"
)

func TestCommitOperation(t *testing.T) {
	// Set up test environment
	env, err := testutils.SetupTestEnv()
	if err != nil {
		t.Fatalf("Failed to set up test environment: %v", err)
	}
	defer env.Cleanup()

	// Test files
	testFiles := []string{
		"src/main.go",
		"src/utils/helper.go",
		"src/models/user.go",
	}

	// Create test files
	filePaths := env.CreateTestFiles(testFiles)

	// Setup mock responses
	env.GitMock.SetupMockChangedFiles(env.TempDir, filePaths)

	// Configure mock commit success
	env.GitMock.SetupMockCommitResult(env.TempDir, true)

	// Setup commit messages in output
	for i, file := range filePaths {
		relPath := testFiles[i]
		commitMsg := "feat: update " + relPath
		output.Set(file, env.TempDir, commitMsg)
	}

	// Run the commit operation
	err = core.CommitAllRoots()
	if err != nil {
		t.Fatalf("Commit operation failed: %v", err)
	}

	// Verify that the commit operation was called with the correct folder
	if env.GitMock.LastCommitFolder.Name != env.TempDir {
		t.Errorf("Expected commit on folder %s, but was %s", env.TempDir, env.GitMock.LastCommitFolder.Name)
	}

	// Verify the number of files included in the commit
	if len(env.GitMock.LastCommitFolder.Files) != len(testFiles) {
		t.Errorf("Expected %d files in commit, got %d", len(testFiles), len(env.GitMock.LastCommitFolder.Files))
	}

	// Check that output is cleared after commit
	outputData := output.GetAll()
	for _, folder := range outputData.Folders {
		if folder.Name == env.TempDir {
			t.Errorf("Expected folder %s to be removed from output after commit", env.TempDir)
		}
	}
}

func TestCommitOperationWithError(t *testing.T) {
	// Set up test environment
	env, err := testutils.SetupTestEnv()
	if err != nil {
		t.Fatalf("Failed to set up test environment: %v", err)
	}
	defer env.Cleanup()

	// Test files
	testFiles := []string{
		"src/main.go",
		"src/utils/helper.go",
	}

	// Create test files
	filePaths := env.CreateTestFiles(testFiles)

	// Setup mock responses
	env.GitMock.SetupMockChangedFiles(env.TempDir, filePaths)

	// Configure mock commit to fail
	env.GitMock.SetupMockCommitResult(env.TempDir, false)

	// Setup commit messages in output
	for i, file := range filePaths {
		relPath := testFiles[i]
		commitMsg := "feat: update " + relPath
		output.Set(file, env.TempDir, commitMsg)
	}

	// Run the commit operation
	err = core.CommitAllRoots()

	// Verify that the function returns an error
	if err == nil {
		t.Errorf("Expected error when commit fails, but got no error")
	}

	// Check that output is not cleared after failed commit
	outputData := output.GetAll()
	foundFolder := false
	for _, folder := range outputData.Folders {
		if folder.Name == env.TempDir {
			foundFolder = true
			break
		}
	}

	if !foundFolder {
		t.Errorf("Expected folder %s to remain in output after failed commit", env.TempDir)
	}
}

func TestCommitOperationWithNoMessages(t *testing.T) {
	// Set up test environment
	env, err := testutils.SetupTestEnv()
	if err != nil {
		t.Fatalf("Failed to set up test environment: %v", err)
	}
	defer env.Cleanup()

	// Clear output to simulate no commit messages
	output.Clear()

	// Run the commit operation
	err = core.CommitAllRoots()

	// Should handle empty commit messages gracefully
	if err != nil {
		t.Errorf("Expected no error with no commit messages, but got: %v", err)
	}
}
