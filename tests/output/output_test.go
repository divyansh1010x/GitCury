package output_test

import (
	"GitCury/output"
	"os"
	"testing"
)

// Helper to set up test environment
func setupOutputTest(t *testing.T) func() {
	// Save original state
	originalData := output.GetAll()

	// Clear output data
	output.Clear()

	// Return cleanup function
	return func() {
		// Restore output data
		output.Clear()
		for _, folder := range originalData.Folders {
			for _, file := range folder.Files {
				output.Set(file.Name, folder.Name, file.Message)
			}
		}
	}
}

// TestSet tests setting commit messages
func TestSet(t *testing.T) {
	cleanup := setupOutputTest(t)
	defer cleanup()

	// Set a commit message
	output.Set("test-file.txt", "test-folder", "Test commit message")

	// Verify the message was set
	message := output.Get("test-file.txt", "test-folder")
	if message != "Test commit message" {
		t.Errorf("Expected message %q, got %q", "Test commit message", message)
	}

	// Set a different message for the same file
	output.Set("test-file.txt", "test-folder", "Updated commit message")

	// Verify the message was updated
	message = output.Get("test-file.txt", "test-folder")
	if message != "Updated commit message" {
		t.Errorf("Expected message %q, got %q", "Updated commit message", message)
	}
}

// TestGet tests retrieving commit messages
func TestGet(t *testing.T) {
	cleanup := setupOutputTest(t)
	defer cleanup()

	// Test with non-existent file
	message := output.Get("non-existent.txt", "test-folder")
	if message != "" {
		t.Errorf("Expected empty message for non-existent file, got %q", message)
	}

	// Set a commit message
	output.Set("test-file.txt", "test-folder", "Test commit message")

	// Verify the message can be retrieved
	message = output.Get("test-file.txt", "test-folder")
	if message != "Test commit message" {
		t.Errorf("Expected message %q, got %q", "Test commit message", message)
	}
}

// TestGetFolder tests retrieving a folder's commit messages
func TestGetFolder(t *testing.T) {
	cleanup := setupOutputTest(t)
	defer cleanup()

	// Test with non-existent folder
	folder := output.GetFolder("non-existent-folder")
	if len(folder.Files) != 0 {
		t.Errorf("Expected empty folder for non-existent folder, got %d files", len(folder.Files))
	}

	// Set commit messages for multiple files in a folder
	output.Set("file1.txt", "test-folder", "Message 1")
	output.Set("file2.txt", "test-folder", "Message 2")

	// Verify the folder can be retrieved with all files
	folder = output.GetFolder("test-folder")
	if len(folder.Files) != 2 {
		t.Errorf("Expected 2 files in folder, got %d", len(folder.Files))
	}

	// Verify folder name
	if folder.Name != "test-folder" {
		t.Errorf("Expected folder name %q, got %q", "test-folder", folder.Name)
	}
}

// TestGetAll tests retrieving all commit messages
func TestGetAll(t *testing.T) {
	cleanup := setupOutputTest(t)
	defer cleanup()

	// Test with empty output
	outputData := output.GetAll()
	if len(outputData.Folders) != 0 {
		t.Errorf("Expected empty output data, got %d folders", len(outputData.Folders))
	}

	// Set commit messages for multiple files in multiple folders
	output.Set("file1.txt", "folder1", "Message 1")
	output.Set("file2.txt", "folder1", "Message 2")
	output.Set("file3.txt", "folder2", "Message 3")

	// Verify all folders and files can be retrieved
	outputData = output.GetAll()
	if len(outputData.Folders) != 2 {
		t.Errorf("Expected 2 folders, got %d", len(outputData.Folders))
	}

	// Count total files
	totalFiles := 0
	for _, folder := range outputData.Folders {
		totalFiles += len(folder.Files)
	}
	if totalFiles != 3 {
		t.Errorf("Expected 3 files total, got %d", totalFiles)
	}
}

// TestRemoveFolder tests removing a folder's commit messages
func TestRemoveFolder(t *testing.T) {
	cleanup := setupOutputTest(t)
	defer cleanup()

	// Set commit messages for multiple files in multiple folders
	output.Set("file1.txt", "folder1", "Message 1")
	output.Set("file2.txt", "folder1", "Message 2")
	output.Set("file3.txt", "folder2", "Message 3")

	// Remove a folder
	output.RemoveFolder("folder1")

	// Verify the folder was removed
	outputData := output.GetAll()
	if len(outputData.Folders) != 1 {
		t.Errorf("Expected 1 folder after removal, got %d", len(outputData.Folders))
	}

	// Verify the remaining folder is the correct one
	if len(outputData.Folders) > 0 && outputData.Folders[0].Name != "folder2" {
		t.Errorf("Expected remaining folder to be %q, got %q", "folder2", outputData.Folders[0].Name)
	}
}

// TestClear tests clearing all commit messages
func TestClear(t *testing.T) {
	cleanup := setupOutputTest(t)
	defer cleanup()

	// Set commit messages for multiple files
	output.Set("file1.txt", "folder1", "Message 1")
	output.Set("file2.txt", "folder1", "Message 2")

	// Clear all output
	output.Clear()

	// Verify all output was cleared
	outputData := output.GetAll()
	if len(outputData.Folders) != 0 {
		t.Errorf("Expected empty output after clear, got %d folders", len(outputData.Folders))
	}
}

// TestSaveToFile tests saving output to a file
func TestSaveToFile(t *testing.T) {
	cleanup := setupOutputTest(t)
	defer cleanup()

	// Set commit messages
	output.Set("file1.txt", "folder1", "Message 1")

	// Get output file path from environment if available, otherwise use a temp file
	outputFile := os.Getenv("GITCURY_TEST_OUTPUT_FILE")
	if outputFile == "" {
		tempFile, err := os.CreateTemp("", "gitcury-output-test-*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		outputFile = tempFile.Name()
		tempFile.Close()
		defer os.Remove(outputFile)
	}

	// Set output file path in config (might require more complex mocking)
	// For now, just verify SaveToFile doesn't crash
	output.SaveToFile()
}
