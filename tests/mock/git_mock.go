package mock

import (
	"GitCury/di"
	"GitCury/interfaces"
	"GitCury/output"
	"GitCury/utils"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// MockGitRunner implements the GitRunner interface for testing
type MockGitRunner struct {
	// State for testing
	ChangedFiles           map[string][]string // map[rootPath][]filePaths
	CommitResults          map[string]bool     // map[folderName]success
	PushResults            map[string]bool     // map[folderName]success
	DiffResults            map[string]string   // map[filePath]diffOutput
	ConfigValues           map[string]string   // map[key]value
	IsRepoResults          map[string]bool     // map[path]isRepo
	CommandResults         map[string]string   // map[command]output
	CommandErrors          map[string]error    // map[command]error
	CommandCalls           []CommandCall       // Record of all commands called
	LastCommitFolder       interfaces.Folder   // Last folder committed
	LastPushBranch         string              // Last branch pushed
	ShouldFailBatchProcess bool                // Control batch processing failures
}

// CommandCall records details of a command that was called
type CommandCall struct {
	Dir     string
	Env     map[string]string
	Args    []string
	Timeout time.Duration
}

// NewMockGitRunner creates a new instance with default testing values
func NewMockGitRunner() *MockGitRunner {
	return &MockGitRunner{
		ChangedFiles:           make(map[string][]string),
		CommitResults:          make(map[string]bool),
		PushResults:            make(map[string]bool),
		DiffResults:            make(map[string]string),
		ConfigValues:           make(map[string]string),
		IsRepoResults:          make(map[string]bool),
		CommandResults:         make(map[string]string),
		CommandErrors:          make(map[string]error),
		CommandCalls:           make([]CommandCall, 0),
		ShouldFailBatchProcess: false, // Default to success
	}
}

// SetupMockChangedFiles configures mock changed files for testing
func (m *MockGitRunner) SetupMockChangedFiles(rootFolder string, files []string) {
	m.ChangedFiles[rootFolder] = files
}

// SetupMockCommitResult configures mock commit results for testing
func (m *MockGitRunner) SetupMockCommitResult(folderName string, success bool) {
	m.CommitResults[folderName] = success
}

// SetupMockPushResult configures mock push results for testing
func (m *MockGitRunner) SetupMockPushResult(folderName string, success bool) {
	m.PushResults[folderName] = success
}

// SetupMockDiffResult configures mock diff results for testing
func (m *MockGitRunner) SetupMockDiffResult(filePath string, diffOutput string) {
	m.DiffResults[filePath] = diffOutput
}

// SetupMockConfigValue configures mock git config values for testing
func (m *MockGitRunner) SetupMockConfigValue(key string, value string) {
	m.ConfigValues[key] = value
}

// SetupMockIsRepo configures mock repository check results for testing
func (m *MockGitRunner) SetupMockIsRepo(path string, isRepo bool) {
	m.IsRepoResults[path] = isRepo
}

// SetupMockCommandResult configures mock command results for testing
func (m *MockGitRunner) SetupMockCommandResult(command string, output string) {
	m.CommandResults[command] = output
}

// SetupMockCommandError configures mock command errors for testing
func (m *MockGitRunner) SetupMockCommandError(command string, err error) {
	m.CommandErrors[command] = err
}

// ClearMockErrors clears all configured command errors
func (m *MockGitRunner) ClearMockErrors() {
	m.CommandErrors = make(map[string]error)
}

// RunGitCmd implements the GitRunner.RunGitCmd interface method
func (m *MockGitRunner) RunGitCmd(dir string, envVars map[string]string, args ...string) (string, error) {
	return m.RunGitCmdWithTimeout(dir, envVars, 30*time.Second, args...)
}

// RunGitCmdWithTimeout implements the GitRunner.RunGitCmdWithTimeout interface method
func (m *MockGitRunner) RunGitCmdWithTimeout(dir string, envVars map[string]string, timeout time.Duration, args ...string) (string, error) {
	// Record the command call
	m.CommandCalls = append(m.CommandCalls, CommandCall{
		Dir:     dir,
		Env:     envVars,
		Args:    args,
		Timeout: timeout,
	})

	// Create a command string for lookup
	cmdStr := strings.Join(args, " ")

	// Check for specific command errors
	if err, ok := m.CommandErrors[cmdStr]; ok && err != nil {
		return "", err
	}

	// Return command result if configured
	if output, ok := m.CommandResults[cmdStr]; ok {
		return output, nil
	}

	// Default behavior based on command type
	switch {
	case len(args) >= 2 && args[0] == "diff":
		// Handle diff command
		filePath := args[len(args)-1]
		if diff, ok := m.DiffResults[filePath]; ok {
			return diff, nil
		}
		return fmt.Sprintf("mock diff for %s", filePath), nil

	case len(args) >= 1 && args[0] == "status":
		// Generate mock status output
		var output strings.Builder
		output.WriteString("On branch main\n")
		output.WriteString("Changes not staged for commit:\n")

		if files, ok := m.ChangedFiles[dir]; ok {
			for _, file := range files {
				output.WriteString(fmt.Sprintf("  modified: %s\n", file))
			}
		}

		return output.String(), nil

	case len(args) >= 2 && args[0] == "config" && args[1] == "--get":
		// Handle config get command
		if len(args) >= 3 {
			key := args[2]
			if value, ok := m.ConfigValues[key]; ok {
				return value, nil
			}
			return "", errors.New("mock: config value not found")
		}
		return "", errors.New("mock: invalid config get command")

	case len(args) >= 3 && args[0] == "config" && args[1] == "--local":
		// Handle config set command
		key := args[2]
		value := args[3]
		m.ConfigValues[key] = value
		return "", nil

	case len(args) >= 2 && args[0] == "add":
		// Handle add command
		return "", nil

	case len(args) >= 2 && args[0] == "commit":
		// Handle commit command
		return "", nil

	case len(args) >= 2 && args[0] == "push":
		// Handle push command
		return "", nil

	default:
		// Default mock response
		return fmt.Sprintf("mock output for git %s", strings.Join(args, " ")), nil
	}
}

// CommitBatch implements the GitRunner.CommitBatch interface method
func (m *MockGitRunner) CommitBatch(folder interfaces.Folder, env ...[]string) error {
	m.LastCommitFolder = folder

	// Check if we have a specific result for this folder
	if result, ok := m.CommitResults[folder.Name]; ok {
		if !result {
			return errors.New("mock commit error for folder: " + folder.Name)
		}
		return nil
	}

	// Default to success
	return nil
}

// GetChangedFiles implements the GitRunner.GetChangedFiles interface method
func (m *MockGitRunner) GetChangedFiles(rootFolders []string, maxConcurrency int, env ...[]string) ([]interfaces.Folder, error) {
	var folders []interfaces.Folder

	for _, rootFolder := range rootFolders {
		if files, ok := m.ChangedFiles[rootFolder]; ok && len(files) > 0 {
			var fileEntries []interfaces.FileEntry
			for _, file := range files {
				fileEntries = append(fileEntries, interfaces.FileEntry{
					Name:    filepath.Join(rootFolder, file),
					Message: "", // Empty message initially
				})
			}

			folders = append(folders, interfaces.Folder{
				Name:  rootFolder,
				Files: fileEntries,
			})
		}
	}

	return folders, nil
}

// Status implements the GitRunner.Status interface method
func (m *MockGitRunner) Status(rootPaths []string) ([]interfaces.Folder, error) {
	return m.GetChangedFiles(rootPaths, 1)
}

// ProcessOneFile implements the GitRunner.ProcessOneFile interface method
func (m *MockGitRunner) ProcessOneFile(filePath, commitMessage string, env ...[]string) error {
	// Check if the file is in our mock changed files
	for _, files := range m.ChangedFiles {
		for _, file := range files {
			if strings.HasSuffix(filePath, file) {
				return nil
			}
		}
	}

	return errors.New("mock: file not found: " + filePath)
}

// GetDiff implements the GitRunner.GetDiff interface method
func (m *MockGitRunner) GetDiff(filePath string, env ...[]string) (string, error) {
	if diff, ok := m.DiffResults[filePath]; ok {
		return diff, nil
	}

	return fmt.Sprintf("mock diff for %s", filePath), nil
}

// IsGitRepository implements the GitRunner.IsGitRepository interface method
func (m *MockGitRunner) IsGitRepository(path string) bool {
	if result, ok := m.IsRepoResults[path]; ok {
		return result
	}

	// Default to true for testing
	return true
}

// GetGitConfigValue implements the GitRunner.GetGitConfigValue interface method
func (m *MockGitRunner) GetGitConfigValue(key string, env ...[]string) (string, error) {
	if value, ok := m.ConfigValues[key]; ok {
		return value, nil
	}

	return "", errors.New("mock: config value not found: " + key)
}

// SetGitConfigValue implements the GitRunner.SetGitConfigValue interface method
func (m *MockGitRunner) SetGitConfigValue(key, value string, env ...[]string) error {
	m.ConfigValues[key] = value
	return nil
}

// ProgressCommitBatch implements the GitRunner.ProgressCommitBatch interface method
func (m *MockGitRunner) ProgressCommitBatch(folder interfaces.Folder, env ...[]string) error {
	m.LastCommitFolder = folder

	// Check if we have a specific result for this folder
	if result, ok := m.CommitResults[folder.Name]; ok {
		if !result {
			return errors.New("mock commit error for folder: " + folder.Name)
		}
		return nil
	}

	// Default to success
	return nil
}

// ProgressPushBranch implements the GitRunner.ProgressPushBranch interface method
func (m *MockGitRunner) ProgressPushBranch(rootFolderName string, branch string) error {
	m.LastPushBranch = branch

	// Record the command call
	m.CommandCalls = append(m.CommandCalls, CommandCall{
		Dir:  rootFolderName,
		Args: []string{"push", "origin", branch},
	})

	// Check if we have a specific result for this folder
	if result, ok := m.PushResults[rootFolderName]; ok {
		if !result {
			return errors.New("mock push error for folder: " + rootFolderName)
		}
		return nil
	}

	// Default to success
	return nil
}

// GetAllChangedFiles implements the GitRunner.GetAllChangedFiles interface method
func (m *MockGitRunner) GetAllChangedFiles(dir string) ([]string, error) {
	if files, ok := m.ChangedFiles[dir]; ok {
		var absolutePaths []string
		for _, file := range files {
			absolutePaths = append(absolutePaths, filepath.Join(dir, file))
		}
		return absolutePaths, nil
	}

	// Return empty slice for testing
	return []string{}, nil
}

// BatchProcessGetMessages implements the GitRunner.BatchProcessGetMessages interface method
func (m *MockGitRunner) BatchProcessGetMessages(allChangedFiles []string, rootFolder string) error {
	// This mock should call the Gemini API to ensure dependency injection testing works correctly
	// Instead of calling the real git.GenCommitMessage which requires real git repos,
	// we'll directly call the dependency-injected Gemini runner

	// Separate binary and text files like the real implementation
	var binaryFiles []string
	var textFiles []string

	for _, file := range allChangedFiles {
		if utils.IsBinaryFile(file) {
			binaryFiles = append(binaryFiles, file)
		} else {
			textFiles = append(textFiles, file)
		}
	}

	// Handle binary files with automated messages
	for _, file := range binaryFiles {
		message := utils.GenerateBinaryCommitMessage(file, "M") // Mock as modified
		output.Set(file, rootFolder, message)
	}

	// Handle text files with Gemini API
	for _, file := range textFiles {
		// Create mock context data for the file
		contextData := map[string]map[string]string{
			file: {
				"type": "modified",
				"diff": fmt.Sprintf("mock diff for %s", file),
			},
		}

		// Call the dependency-injected Gemini runner to test the DI system
		message, err := di.GetGeminiRunner().SendToGemini(contextData, "test-api-key")
		if err != nil {
			// If there's an error, use a fallback mock message
			message = fmt.Sprintf("Mock commit message for %s", filepath.Base(file))
		}

		// Use the output system to store the message
		output.Set(file, rootFolder, message)
	}

	// Check if there's a predefined error for this operation
	if m.ShouldFailBatchProcess {
		return errors.New("mock: batch process failed")
	}

	return nil
}

// BatchProcessWithEmbeddings implements the GitRunner.BatchProcessWithEmbeddings interface method
func (m *MockGitRunner) BatchProcessWithEmbeddings(allChangedFiles []string, rootFolder string, numClusters int) error {
	// Mock implementation that simulates successful clustering and processing with binary file handling

	// Separate binary and text files like the real implementation
	var binaryFiles []string
	var textFiles []string

	for _, file := range allChangedFiles {
		if utils.IsBinaryFile(file) {
			binaryFiles = append(binaryFiles, file)
		} else {
			textFiles = append(textFiles, file)
		}
	}

	// Handle binary files with automated messages
	for _, file := range binaryFiles {
		message := utils.GenerateBinaryCommitMessage(file, "M") // Mock as modified
		output.Set(file, rootFolder, message)
	}

	// Handle text files with clustering simulation
	for i, file := range textFiles {
		// Generate a mock grouped commit message
		clusterID := i % numClusters
		message := fmt.Sprintf("Mock grouped commit message for cluster %d: %s", clusterID, filepath.Base(file))

		// Use the output system to store the message
		output.Set(file, rootFolder, message)
	}

	// Check if there's a predefined error for this operation
	if m.ShouldFailBatchProcess {
		return errors.New("mock: batch process with embeddings failed")
	}

	return nil
}

// Ensure MockGitRunner implements GitRunner interface
var _ interfaces.GitRunner = (*MockGitRunner)(nil)
