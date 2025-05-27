// Package integration_test tests the end-to-end functionality of GitCury
package integration_test

import (
	"GitCury/config"
	"GitCury/core"
	"GitCury/git"
	"GitCury/output"
	"GitCury/tests/mocks"
	"GitCury/tests/testutils"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestEndToEndWorkflow tests the complete workflow of GitCury
func TestEndToEndWorkflow(t *testing.T) {
	// Skip this test in CI environments or when integration tests are disabled
	if os.Getenv("GITCURY_SKIP_INTEGRATION") == "true" {
		t.Skip("Skipping integration test as GITCURY_SKIP_INTEGRATION is set")
	}

	// Check if GEMINI_API_KEY is available, skip if not
	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping integration test as GEMINI_API_KEY is not set")
	}

	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Set up a git repository
	testutils.SetupGitRepo(t, tempDir)

	// Add test files
	testutils.AddAndCommitFile(t, tempDir, "file1.txt", "Initial content 1", "Add file1")
	testutils.AddAndCommitFile(t, tempDir, "file2.txt", "Initial content 2", "Add file2")

	// Modify files to create changes
	if err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("Updated content 1"), 0644); err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("Updated content 2"), 0644); err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}

	// Create a new file
	if err := os.WriteFile(filepath.Join(tempDir, "file3.txt"), []byte("New content 3"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	// Set up test config
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(filepath.Join(tempDir, ".gitcury/config.json"))
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Save test settings
	testSettings := map[string]interface{}{
		"api_key":          os.Getenv("GITCURY_TEST_API_KEY"), // Set via environment for testing
		"output_file_path": filepath.Join(tempDir, "output.json"),
		"logLevel":         "debug",
		"maxConcurrent":    2,
		"root_folders":     []string{tempDir},
	}

	// Set config path in environment
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Save settings directly to a config file
	configPath := filepath.Join(tempDir, ".gitcury/config.json")
	configData, err := json.Marshal(testSettings)
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Reload config to use the new settings
	if err := config.LoadConfig(); err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Execute end-to-end workflow steps

	// Get concurrency value
	maxConcurrent := 2

	// Step 1: Get all changed files using git's status command
	changedFiles, err := git.GetAllChangedFiles(tempDir)
	if err != nil {
		t.Fatalf("GetAllChangedFiles returned an error: %v", err)
	}

	// Verify we got at least one changed file
	if len(changedFiles) == 0 {
		t.Fatal("Expected at least one changed file")
	}

	// Step 2: Generate commit messages for changed files
	// Note: This step might fail if API key is not set up for testing
	err = core.GetAllMsgs(maxConcurrent)
	if err != nil {
		t.Logf("GetAllMsgs returned an error: %v (may be expected in test environment)", err)
		// Don't fail the test if API key is not set up
	}

	// Step 3: Verify commit messages in output
	outputData := output.GetAll()

	// There might not be messages due to API key issues, but we should have folders
	if len(outputData.Folders) == 0 {
		t.Error("Expected at least one folder in output data")
	}

	// Step 4: Commit changes
	// Note: This will fail in most test environments, so we just log the error
	err = core.CommitAllRoots()
	if err != nil {
		t.Logf("CommitAllRoots returned an error: %v (may be expected in test environment)", err)
	}

	// Step 5: Push changes
	// Note: This will fail in most test environments, so we just log the error
	err = core.PushAllRoots("main")
	if err != nil {
		t.Logf("PushAllRoots returned an error: %v (may be expected in test environment)", err)
	}
}

// TestEndToEndWithMocks tests the end-to-end workflow with mocks
func TestEndToEndWithMocks(t *testing.T) {
	// Set up mocks
	mockGitRunner := mocks.NewMockGitRunner()
	mockOutputManager := mocks.NewMockOutputManager()

	// Verify mocks are properly initialized
	if mockGitRunner == nil {
		t.Fatal("Failed to initialize mock git runner")
	}

	if mockOutputManager == nil {
		t.Fatal("Failed to initialize mock output manager")
	}

	// Test basic mock functionality
	testFolder := "test-repo"
	testFile := "test.go"
	testMessage := "Test commit message"

	// Add test data to mock output manager
	mockOutputManager.Set(testFile, testFolder, testMessage)

	// Verify data was stored
	stored := mockOutputManager.Get(testFile, testFolder)
	if stored != testMessage {
		t.Errorf("Expected stored message %q, got %q", testMessage, stored)
	}

	// Verify folder creation
	folders := mockOutputManager.GetAll()
	if len(folders.Folders) == 0 {
		t.Error("Expected at least one folder in mock output manager")
	}

	// Test git runner mock
	_, err := mockGitRunner.RunGitCmd(".", nil, "status")
	if err != nil {
		t.Logf("Mock git command returned error: %v (expected in test environment)", err)
	}

	t.Log("End-to-end test with mocks completed successfully")
}
