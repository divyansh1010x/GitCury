package end_to_end

import (
	"github.com/lakshyajain-0291/gitcury/config"
	"github.com/lakshyajain-0291/gitcury/core"
	"github.com/lakshyajain-0291/gitcury/output"
	"github.com/lakshyajain-0291/gitcury/tests/testutils"
	"os"
	"path/filepath"
	"testing"
)

func TestEndToEndWorkflow(t *testing.T) {
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
	}

	// Create test files
	filePaths := env.CreateTestFiles(testFiles)

	// Configure the mock git runner
	env.GitMock.SetupMockChangedFiles(env.TempDir, filePaths)
	env.GitMock.SetupMockCommitResult(env.TempDir, true)
	env.GitMock.SetupMockPushResult(env.TempDir, true)
	env.GitMock.SetupMockIsRepo(env.TempDir, true)

	// Configure mock commit messages
	for i, file := range filePaths {
		baseName := filepath.Base(file)
		var message string
		switch i % 3 {
		case 0:
			message = "feat: update " + baseName
		case 1:
			message = "fix: resolve issue in " + baseName
		case 2:
			message = "docs: improve documentation in " + baseName
		}
		env.GeminiMock.SetupMockCommitMessage(file, message)
	}

	// Configure root folders
	config.Set("root_folders", []interface{}{env.TempDir})

	// Step 1: Generate commit messages
	err = core.GetAllMsgs(10)
	if err != nil {
		t.Fatalf("Message generation failed: %v", err)
	}

	// Verify messages were generated
	outputData := output.GetAll()
	if len(outputData.Folders) != 1 {
		t.Fatalf("Expected 1 folder in output, got %d", len(outputData.Folders))
	}

	var testFolder output.Folder
	for _, folder := range outputData.Folders {
		if folder.Name == env.TempDir {
			testFolder = folder
			break
		}
	}

	if len(testFolder.Files) != len(testFiles) {
		t.Errorf("Expected %d files in output, got %d", len(testFiles), len(testFolder.Files))
	}

	// Step 2: Commit changes
	err = core.CommitAllRoots()
	if err != nil {
		t.Fatalf("Commit operation failed: %v", err)
	}

	// Verify commit was called
	if env.GitMock.LastCommitFolder.Name != env.TempDir {
		t.Errorf("Expected commit on folder %s, but was %s", env.TempDir, env.GitMock.LastCommitFolder.Name)
	}

	// Step 3: Push changes
	err = core.PushAllRoots("main")
	if err != nil {
		t.Fatalf("Push operation failed: %v", err)
	}

	// Verify the branch was pushed
	if env.GitMock.LastPushBranch != "main" {
		t.Errorf("Expected push to branch 'main', but was %s", env.GitMock.LastPushBranch)
	}

	// Verify the output is cleared after commit
	outputData = output.GetAll()
	if len(outputData.Folders) > 0 {
		t.Errorf("Expected output to be cleared after commit, but still has %d folders", len(outputData.Folders))
	}
}

// Test the end-to-end command provided by the cmd package
func TestEndToEndCommand(t *testing.T) {
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

	// Configure the mock git runner
	env.GitMock.SetupMockChangedFiles(env.TempDir, filePaths)
	env.GitMock.SetupMockCommitResult(env.TempDir, true)
	env.GitMock.SetupMockPushResult(env.TempDir, true)
	env.GitMock.SetupMockIsRepo(env.TempDir, true)

	// Configure mock commit messages
	for _, file := range filePaths {
		baseName := filepath.Base(file)
		message := "feat: update " + baseName
		env.GeminiMock.SetupMockCommitMessage(file, message)
	}

	// Configure root folders
	config.Set("root_folders", []interface{}{env.TempDir})

	// Set non-interactive mode to avoid confirmation prompts
	originalEnv := os.Getenv("GITCURY_NONINTERACTIVE")
	os.Setenv("GITCURY_NONINTERACTIVE", "1")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("GITCURY_NONINTERACTIVE")
		} else {
			os.Setenv("GITCURY_NONINTERACTIVE", originalEnv)
		}
	}()

	// Execute the end-to-end command
	// Note: In a real scenario, this would be cmd.Execute() with arguments,
	// but for testing we directly call the core functions it would trigger
	err = core.GetAllMsgs(10)
	if err != nil {
		t.Fatalf("Message generation failed: %v", err)
	}

	err = core.CommitAllRoots()
	if err != nil {
		t.Fatalf("Commit operation failed: %v", err)
	}

	err = core.PushAllRoots("main")
	if err != nil {
		t.Fatalf("Push operation failed: %v", err)
	}

	// Verify all steps executed successfully
	if env.GeminiMock.CallCount == 0 {
		t.Errorf("Expected Gemini API to be called, but it wasn't")
	}

	if env.GitMock.LastCommitFolder.Name != env.TempDir {
		t.Errorf("Expected commit operation to be called, but it wasn't")
	}

	if env.GitMock.LastPushBranch != "main" {
		t.Errorf("Expected push operation to be called with branch 'main', but it wasn't")
	}
}
