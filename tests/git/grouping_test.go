package git_test

import (
	"GitCury/git"
	"GitCury/tests/testutils"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestIsBinaryFile tests the binary file detection functionality
func TestIsBinaryFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	tests := []struct {
		name         string
		content      []byte
		filename     string
		expectBinary bool
	}{
		{
			name:         "Text file with normal content",
			content:      []byte("This is a normal text file\nwith multiple lines\n"),
			filename:     "test.txt",
			expectBinary: false,
		},
		{
			name:         "Binary file with null bytes",
			content:      []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}, // PNG header with null
			filename:     "test.png",
			expectBinary: true,
		},
		{
			name:         "Executable file",
			content:      []byte{0x7F, 0x45, 0x4C, 0x46, 0x00, 0x01, 0x02, 0x00}, // ELF header with null bytes
			filename:     "executable",
			expectBinary: true,
		},
		{
			name:         "Go source file",
			content:      []byte("package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}"),
			filename:     "main.go",
			expectBinary: false,
		},
		{
			name:         "JSON file",
			content:      []byte(`{"name": "test", "value": 123}`),
			filename:     "config.json",
			expectBinary: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, tt.filename)
			if err := os.WriteFile(testFile, tt.content, 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test binary detection
			isBinary := git.IsBinaryFile(testFile)
			if isBinary != tt.expectBinary {
				t.Errorf("IsBinaryFile(%s) = %v, expected %v", tt.filename, isBinary, tt.expectBinary)
			}

			// Clean up
			os.Remove(testFile)
		})
	}
}

// TestIsIgnoredFile tests the file filtering functionality
func TestIsIgnoredFile(t *testing.T) {
	tests := []struct {
		filename      string
		expectIgnored bool
	}{
		{"main.go", false},
		{"README.md", false},
		{"config.json", false},
		{"test.txt", false},
		{"gitcury", true}, // Known executable name that should be ignored
		{"program.exe", true},
		{"library.so", true},
		{"archive.zip", true},
		{"image.png", true},
		{"document.pdf", true},
		{".gitignore", false},
		{"Dockerfile", false},
		{"package.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			isIgnored := git.IsIgnoredFile(tt.filename)
			if isIgnored != tt.expectIgnored {
				t.Errorf("IsIgnoredFile(%s) = %v, expected %v", tt.filename, isIgnored, tt.expectIgnored)
			}
		})
	}
}

// TestBatchProcessWithEmbeddingsFiltering tests that binary files are properly filtered
func TestBatchProcessWithEmbeddingsFiltering(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Set up a git repository
	testutils.SetupGitRepo(t, tempDir)

	// Create test files - mix of text and binary
	files := map[string][]byte{
		"main.go":     []byte("package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}"),
		"README.md":   []byte("# Test Project\n\nThis is a test project."),
		"config.json": []byte(`{"name": "test", "debug": true}`),
		"binary_data": {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}, // Binary with null bytes
		"executable":  {0x7F, 0x45, 0x4C, 0x46, 0x00, 0x01, 0x02, 0x00},             // ELF header with null bytes
		"image.png":   {0x89, 0x50, 0x4E, 0x47, 0x00, 0x00},                         // PNG header with null bytes
	}

	// Create all test files
	for filename, content := range files {
		testFile := filepath.Join(tempDir, filename)
		if err := os.WriteFile(testFile, content, 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}

		// Add files to git
		_, err := git.RunGitCmd(tempDir, nil, "add", filename)
		if err != nil {
			t.Fatalf("Failed to add file %s to git: %v", filename, err)
		}
	}

	// Test that GetAllChangedFiles properly filters binary files
	// This function is designed to return only non-binary, non-ignored files
	changedFiles, err := git.GetAllChangedFiles(tempDir)
	if err != nil {
		t.Fatalf("Failed to get changed files: %v", err)
	}

	t.Logf("Changed files returned by GetAllChangedFiles: %v", changedFiles)

	// GetAllChangedFiles should return only the 3 text files (binary files are filtered out)
	expectedTextFiles := 3
	if len(changedFiles) != expectedTextFiles {
		t.Errorf("Expected GetAllChangedFiles to return %d text files, got %d", expectedTextFiles, len(changedFiles))
	}

	// Verify that all returned files are indeed text files
	for _, file := range changedFiles {
		if git.IsBinaryFile(file) {
			t.Errorf("GetAllChangedFiles returned binary file: %s", file)
		}
		if git.IsIgnoredFile(file) {
			t.Errorf("GetAllChangedFiles returned ignored file: %s", file)
		}
	}

	// Test the binary file detection separately by checking all files in the directory
	allFiles := []string{"main.go", "README.md", "config.json", "binary_data", "executable", "image.png"}
	textCount := 0
	binaryCount := 0

	for _, filename := range allFiles {
		fullPath := filepath.Join(tempDir, filename)
		if git.IsBinaryFile(fullPath) {
			binaryCount++
			t.Logf("File %s correctly detected as binary", filename)
		} else {
			textCount++
			t.Logf("File %s correctly detected as text", filename)
		}
	}

	// Verify our binary detection works correctly
	if textCount != 3 {
		t.Errorf("Expected 3 text files, detected %d", textCount)
	}
	if binaryCount != 3 {
		t.Errorf("Expected 3 binary files, detected %d", binaryCount)
	}

	t.Logf("Binary file filtering test completed: %d text files processed, %d binary files filtered out", textCount, binaryCount)
}

// TestProcessFilesForGrouping tests the end-to-end file processing for grouping
func TestProcessFilesForGrouping(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Set up a git repository
	testutils.SetupGitRepo(t, tempDir)

	// Create related files that should be grouped together
	files := map[string]string{
		"user.go": `package models

type User struct {
	ID   int    ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}

func (u *User) GetName() string {
	return u.Name
}`,
		"user_test.go": `package models

import "testing"

func TestUser_GetName(t *testing.T) {
	user := &User{Name: "John"}
	if user.GetName() != "John" {
		t.Error("Expected name to be John")
	}
}`,
		"README.md": `# User Model

This package contains the User model and related functionality.

## Usage

` + "```go" + `
user := &User{ID: 1, Name: "John"}
name := user.GetName()
` + "```" + ``,
	}

	// Create all test files
	for filename, content := range files {
		testFile := filepath.Join(tempDir, filename)
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}

		// Add files to git
		_, err := git.RunGitCmd(tempDir, nil, "add", filename)
		if err != nil {
			t.Fatalf("Failed to add file %s to git: %v", filename, err)
		}
	}

	// Get the changed files
	changedFiles, err := git.GetAllChangedFiles(tempDir)
	if err != nil {
		t.Fatalf("Failed to get changed files: %v", err)
	}

	if len(changedFiles) < 3 {
		t.Fatalf("Expected at least 3 changed files, got %d", len(changedFiles))
	}

	// Verify that all our test files are present
	expectedFiles := []string{"user.go", "user_test.go", "README.md"}
	for _, expectedFile := range expectedFiles {
		found := false
		for _, file := range changedFiles {
			if strings.Contains(file, expectedFile) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected file %s not found in changed files", expectedFile)
		}
	}

	t.Logf("File processing test completed with %d files ready for grouping", len(changedFiles))

	// Log the files that would be processed
	for _, file := range changedFiles {
		fullPath := filepath.Join(tempDir, file)
		if !git.IsBinaryFile(fullPath) && !git.IsIgnoredFile(file) {
			t.Logf("File ready for semantic grouping: %s", file)
		}
	}
}

// TestGetFileDiff tests the diff generation for files
func TestGetFileDiff(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := testutils.CreateTempDir(t)

	// Set up a git repository
	testutils.SetupGitRepo(t, tempDir)

	// Create and commit an initial file
	initialFile := filepath.Join(tempDir, "test.go")
	initialContent := "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}"
	if err := os.WriteFile(initialFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	_, err := git.RunGitCmd(tempDir, nil, "add", "test.go")
	if err != nil {
		t.Fatalf("Failed to add initial file: %v", err)
	}

	_, err = git.RunGitCmd(tempDir, nil, "commit", "-m", "Initial commit")
	if err != nil {
		t.Fatalf("Failed to commit initial file: %v", err)
	}

	// Modify the file
	modifiedContent := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n\tfmt.Println(\"Modified version\")\n}"
	if err := os.WriteFile(initialFile, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("Failed to modify file: %v", err)
	}

	// Get the diff
	diff, err := git.RunGitCmd(tempDir, nil, "diff", "test.go")
	if err != nil {
		t.Fatalf("Failed to get diff: %v", err)
	}

	// Verify diff contains expected changes
	if !strings.Contains(diff, "+import \"fmt\"") {
		t.Error("Expected diff to contain import addition")
	}

	if !strings.Contains(diff, "+\tfmt.Println(\"Hello, World!\")") {
		t.Error("Expected diff to contain modified println")
	}

	if !strings.Contains(diff, "+\tfmt.Println(\"Modified version\")") {
		t.Error("Expected diff to contain new println")
	}

	t.Logf("Diff generation test completed successfully")
}
