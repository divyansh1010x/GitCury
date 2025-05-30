package end_to_end

import (
	"github.com/lakshyajain-0291/GitCury/output"
	"github.com/lakshyajain-0291/GitCury/tests/testutils"
	"testing"
)

func TestMessageGeneration(t *testing.T) {
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
		"docs/README.md",
		"tests/test_helper.go",
	}

	// Create test files
	filePaths := env.CreateTestFiles(testFiles)

	// Setup mock responses
	env.GitMock.SetupMockChangedFiles(env.TempDir, filePaths)

	// Configure expected commit messages directly in the output system
	expectedMessages := map[string]string{
		"src/main.go":          "feat: update main application entry point",
		"src/utils/helper.go":  "refactor: improve helper function efficiency",
		"src/models/user.go":   "feat: add user authentication model",
		"docs/README.md":       "docs: update installation instructions",
		"tests/test_helper.go": "test: add new test helpers",
	}

	// Mock the messages
	env.MockMessages(testFiles, expectedMessages)

	// Verify the results
	outputData := output.GetAll()

	// Check that we have the expected folder
	if len(outputData.Folders) != 1 {
		t.Errorf("Expected 1 folder, got %d", len(outputData.Folders))
	}

	// Find our test folder
	var testFolder output.Folder
	var foundFolder bool
	for _, folder := range outputData.Folders {
		if folder.Name == env.TempDir {
			testFolder = folder
			foundFolder = true
			break
		}
	}

	if !foundFolder {
		t.Fatalf("Test folder not found in output data")
	}

	// Check the number of files
	if len(testFolder.Files) != len(testFiles) {
		t.Errorf("Expected %d files, got %d", len(testFiles), len(testFolder.Files))
	}

	// Check each file message
	filePathMap := make(map[string]string)
	for i, file := range testFiles {
		filePathMap[filePaths[i]] = file
	}

	for _, file := range testFolder.Files {
		relFile, ok := filePathMap[file.Name]
		if !ok {
			t.Errorf("Unexpected file in output: %s", file.Name)
			continue
		}

		expectedMsg := expectedMessages[relFile]
		if file.Message != expectedMsg {
			t.Errorf("For file %s, expected message %q, got %q", relFile, expectedMsg, file.Message)
		}
	}
}
