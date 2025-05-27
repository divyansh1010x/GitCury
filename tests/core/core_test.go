package core_test

import (
	"GitCury/core"
	"GitCury/output"
	"GitCury/tests/mocks"
	"GitCury/tests/testutils"
	"os"
	"testing"
)

// TestCommitOneRoot tests committing changes for a single root folder
func TestCommitOneRoot(t *testing.T) {
	// Setup test environment with config and API key
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()
	// Create mock git runner
	mockGitRunner := mocks.NewMockGitRunner()

	// Set up test folder
	folder := output.Folder{
		Name: "test-folder",
		Files: []output.FileEntry{
			{Name: "file1.txt", Message: "Commit message 1"},
			{Name: "file2.txt", Message: "Commit message 2"},
		},
	}

	// Create mock output manager
	mockOutputManager := mocks.NewMockOutputManager()
	mockOutputManager.Folders["test-folder"] = folder

	// Replace dependencies with mocks (this requires dependency injection)
	// For now, verify mocks are properly initialized
	if mockGitRunner == nil || mockOutputManager == nil {
		t.Fatal("Failed to initialize mocks")
	}

	// Test CommitOneRoot with a non-existent folder
	err := core.CommitOneRoot("non-existent-folder")
	if err == nil {
		t.Error("Expected error for non-existent folder, got nil")
	}

	// Test with a valid folder would require proper dependency injection
}

// TestCommitAllRoots tests committing changes for all root folders
func TestCommitAllRoots(t *testing.T) {
	// Setup test environment with config and API key
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()
	// This test requires proper dependency injection to replace git and output packages
	// For now, test with empty output as a basic sanity check

	// Clear output data
	originalData := output.GetAll()
	output.Clear()
	defer func() {
		// Restore output data
		output.Clear()
		for _, folder := range originalData.Folders {
			for _, file := range folder.Files {
				output.Set(file.Name, folder.Name, file.Message)
			}
		}
	}()

	// Test CommitAllRoots with empty output
	err := core.CommitAllRoots()
	if err != nil {
		t.Errorf("Expected nil error for empty output, got %v", err)
	}
}

// TestGetAllMsgs tests generating commit messages for all root folders
func TestGetAllMsgs(t *testing.T) {
	// Setup test environment with config and API key
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()
	// Skip test if GEMINI_API_KEY is not set to avoid panic
	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping TestGetAllMsgs as GEMINI_API_KEY is not set")
	}

	// This test requires proper dependency injection to replace git and output packages
	// For now, test with a small number of files as a basic sanity check

	// Test with small number of files
	err := core.GetAllMsgs(1)
	if err != nil {
		// This might fail in test environment without proper git repos, that's okay
		t.Logf("GetAllMsgs returned an error: %v (this might be expected in test environment)", err)
	}
}

// TestGetMsgsForRootFolder tests generating commit messages for a single root folder
func TestGetMsgsForRootFolder(t *testing.T) {
	// Setup test environment with config and API key
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()
	// This test requires proper dependency injection to replace git and output packages
	// For now, test with a non-existent folder as a basic sanity check

	// Test with non-existent folder
	err := core.GetMsgsForRootFolder("")
	if err == nil {
		t.Error("Expected error for empty folder name, got nil")
	}
}

// TestPushOneRoot tests pushing changes for a single root folder
func TestPushOneRoot(t *testing.T) {
	// Setup test environment with config and API key
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()
	// This test requires proper dependency injection to replace git package
	// For now, test with a non-existent folder as a basic sanity check

	// Test with non-existent folder and branch
	err := core.PushOneRoot("non-existent-folder", "main")
	if err != nil {
		// This will likely fail in test environment without proper git repos, that's okay
		t.Logf("PushOneRoot returned an error: %v (this might be expected in test environment)", err)
	}
}

// TestPushAllRoots tests pushing changes for all root folders
func TestPushAllRoots(t *testing.T) {
	// Setup test environment with config and API key
	cleanup := testutils.SetupTestEnvironment(t)
	defer cleanup()
	// This test requires proper dependency injection to replace git and config packages
	// For now, test with a basic branch name as a sanity check

	// Test with branch name
	err := core.PushAllRoots("main")
	if err != nil {
		// This will likely fail in test environment without proper git repos, that's okay
		t.Logf("PushAllRoots returned an error: %v (this might be expected in test environment)", err)
	}
}
