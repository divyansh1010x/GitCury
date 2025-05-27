// Package mocks provides mock implementations for testing
package mocks

import (
	"GitCury/interfaces"
	"GitCury/output"
	"errors"
	"sync"
	"time"
)

// MockGitRunner is a mock implementation of git command runner for testing
type MockGitRunner struct {
	Commands          []string
	Directories       []string
	EnvVars           []map[string]string
	ReturnValueMap    map[string]string
	ReturnErrorMap    map[string]error
	DirReturnValueMap map[string]map[string]string // Map of directory to command to response
	DirReturnErrorMap map[string]map[string]error  // Map of directory to command to error
	CallCount         map[string]int
	mu                sync.Mutex
	DefaultResponse   string
	DefaultError      error
}

// Ensure MockGitRunner implements GitRunner interface
var _ interfaces.GitRunner = &MockGitRunner{}

// NewMockGitRunner creates a new mock git runner
func NewMockGitRunner() *MockGitRunner {
	return &MockGitRunner{
		Commands:          make([]string, 0),
		Directories:       make([]string, 0),
		EnvVars:           make([]map[string]string, 0),
		ReturnValueMap:    make(map[string]string),
		ReturnErrorMap:    make(map[string]error),
		DirReturnValueMap: make(map[string]map[string]string),
		DirReturnErrorMap: make(map[string]map[string]error),
		CallCount:         make(map[string]int),
		DefaultResponse:   "",
		DefaultError:      nil,
	}
}

// RunGitCommand records the command and returns predefined response
func (m *MockGitRunner) RunGitCommand(dir string, envVars map[string]string, args ...string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Build command string for lookup
	cmd := args[0]
	if len(args) > 1 {
		cmd = args[0] + " " + args[1]
	}

	// Record this call
	m.Commands = append(m.Commands, cmd)
	m.Directories = append(m.Directories, dir)
	m.EnvVars = append(m.EnvVars, envVars)

	// Increment call count
	m.CallCount[cmd] = m.CallCount[cmd] + 1

	// Check if we have a directory-specific response
	if dirMap, ok := m.DirReturnValueMap[dir]; ok {
		if response, ok := dirMap[cmd]; ok {
			if dirErrMap, ok := m.DirReturnErrorMap[dir]; ok {
				if err, ok := dirErrMap[cmd]; ok {
					return response, err
				}
			}
			return response, nil
		}
	}

	// Check if we have a predefined response for this command
	if response, ok := m.ReturnValueMap[cmd]; ok {
		if err, ok := m.ReturnErrorMap[cmd]; ok {
			return response, err
		}
		return response, nil
	}

	return m.DefaultResponse, m.DefaultError
}

// RunGitCmd implements the GitRunner interface
func (m *MockGitRunner) RunGitCmd(dir string, envVars map[string]string, args ...string) (string, error) {
	return m.RunGitCommand(dir, envVars, args...)
}

// RunGitCmdWithTimeout implements the GitRunner interface
func (m *MockGitRunner) RunGitCmdWithTimeout(dir string, envVars map[string]string, timeout time.Duration, args ...string) (string, error) {
	return m.RunGitCommand(dir, envVars, args...)
}

// CommitBatch implements the GitRunner interface
func (m *MockGitRunner) CommitBatch(folder output.Folder, env ...[]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record the commit batch operation
	cmd := "commit-batch-" + folder.Name
	m.Commands = append(m.Commands, cmd)
	m.CallCount[cmd] = m.CallCount[cmd] + 1

	// Check if we have a predefined error for this operation
	if err, ok := m.ReturnErrorMap[cmd]; ok {
		return err
	}

	return m.DefaultError
}

// GetChangedFiles implements the GitRunner interface
func (m *MockGitRunner) GetChangedFiles(rootFolders []string, maxConcurrency int, env ...[]string) ([]output.Folder, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := "get-changed-files"
	m.Commands = append(m.Commands, cmd)
	m.CallCount[cmd] = m.CallCount[cmd] + 1

	// Return empty folders by default
	return []output.Folder{}, m.DefaultError
}

// Status implements the GitRunner interface
func (m *MockGitRunner) Status(rootPaths []string) ([]output.Folder, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := "status"
	m.Commands = append(m.Commands, cmd)
	m.CallCount[cmd] = m.CallCount[cmd] + 1

	// Return empty folders by default
	return []output.Folder{}, m.DefaultError
}

// ProcessOneFile implements the GitRunner interface
func (m *MockGitRunner) ProcessOneFile(filePath, commitMessage string, env ...[]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := "process-one-file-" + filePath
	m.Commands = append(m.Commands, cmd)
	m.CallCount[cmd] = m.CallCount[cmd] + 1

	// Check if we have a predefined error for this operation
	if err, ok := m.ReturnErrorMap[cmd]; ok {
		return err
	}

	return m.DefaultError
}

// GetDiff implements the GitRunner interface
func (m *MockGitRunner) GetDiff(filePath string, env ...[]string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := "diff-" + filePath
	m.Commands = append(m.Commands, cmd)
	m.CallCount[cmd] = m.CallCount[cmd] + 1

	// Check if we have a predefined response for this file
	if response, ok := m.ReturnValueMap[cmd]; ok {
		if err, ok := m.ReturnErrorMap[cmd]; ok {
			return response, err
		}
		return response, nil
	}

	return m.DefaultResponse, m.DefaultError
}

// IsGitRepository implements the GitRunner interface
func (m *MockGitRunner) IsGitRepository(path string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := "is-git-repo-" + path
	m.Commands = append(m.Commands, cmd)
	m.CallCount[cmd] = m.CallCount[cmd] + 1

	// Check if we have a predefined response
	if response, ok := m.ReturnValueMap[cmd]; ok {
		return response == "true"
	}

	// Default to true for testing
	return true
}

// GetGitConfigValue implements the GitRunner interface
func (m *MockGitRunner) GetGitConfigValue(key string, env ...[]string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := "config-get-" + key
	m.Commands = append(m.Commands, cmd)
	m.CallCount[cmd] = m.CallCount[cmd] + 1

	// Check if we have a predefined response
	if response, ok := m.ReturnValueMap[cmd]; ok {
		if err, ok := m.ReturnErrorMap[cmd]; ok {
			return response, err
		}
		return response, nil
	}

	return m.DefaultResponse, m.DefaultError
}

// SetGitConfigValue implements the GitRunner interface
func (m *MockGitRunner) SetGitConfigValue(key, value string, env ...[]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := "config-set-" + key + "-" + value
	m.Commands = append(m.Commands, cmd)
	m.CallCount[cmd] = m.CallCount[cmd] + 1

	// Check if we have a predefined error
	if err, ok := m.ReturnErrorMap[cmd]; ok {
		return err
	}

	return m.DefaultError
}

// MockOutputManager mocks the output.go functionality
type MockOutputManager struct {
	Folders      map[string]output.Folder
	SavedToFile  bool
	ClearedCalls int
	mu           sync.Mutex
}

// Ensure MockOutputManager implements OutputManager interface
var _ interfaces.OutputManager = &MockOutputManager{}

// NewMockOutputManager creates a new mock output manager
func NewMockOutputManager() *MockOutputManager {
	return &MockOutputManager{
		Folders:      make(map[string]output.Folder),
		SavedToFile:  false,
		ClearedCalls: 0,
	}
}

// Set mocks output.Set
func (m *MockOutputManager) Set(file, rootFolder, commitMessage string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	folder, ok := m.Folders[rootFolder]
	if !ok {
		folder = output.Folder{
			Name:  rootFolder,
			Files: []output.FileEntry{},
		}
	}

	updated := false
	for i, entry := range folder.Files {
		if entry.Name == file {
			folder.Files[i].Message = commitMessage
			updated = true
			break
		}
	}

	if !updated {
		folder.Files = append(folder.Files, output.FileEntry{
			Name:    file,
			Message: commitMessage,
		})
	}

	m.Folders[rootFolder] = folder
}

// Get mocks output.Get
func (m *MockOutputManager) Get(file, rootFolder string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	folder, ok := m.Folders[rootFolder]
	if !ok {
		return ""
	}

	for _, entry := range folder.Files {
		if entry.Name == file {
			return entry.Message
		}
	}
	return ""
}

// GetFolder mocks output.GetFolder
func (m *MockOutputManager) GetFolder(rootFolder string) output.Folder {
	m.mu.Lock()
	defer m.mu.Unlock()

	if folder, ok := m.Folders[rootFolder]; ok {
		return folder
	}

	return output.Folder{
		Name:  rootFolder,
		Files: []output.FileEntry{},
	}
}

// GetAll mocks output.GetAll
func (m *MockOutputManager) GetAll() output.OutputData {
	m.mu.Lock()
	defer m.mu.Unlock()

	folders := make([]output.Folder, 0, len(m.Folders))
	for _, folder := range m.Folders {
		folders = append(folders, folder)
	}

	return output.OutputData{
		Folders: folders,
	}
}

// RemoveFolder implements the OutputManager interface
func (m *MockOutputManager) RemoveFolder(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.Folders, name)
}

// Clear mocks output.Clear
func (m *MockOutputManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Folders = make(map[string]output.Folder)
	m.ClearedCalls++
}

// SaveToFile implements the OutputManager interface
func (m *MockOutputManager) SaveToFile() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SavedToFile = true
}

// MockConfig mocks the config functionality
type MockConfig struct {
	Settings map[string]interface{}
	mu       sync.Mutex
}

// NewMockConfig creates a new mock config
func NewMockConfig() *MockConfig {
	return &MockConfig{
		Settings: make(map[string]interface{}),
	}
}

// Get mocks config.Get
func (m *MockConfig) Get(key string) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	if value, ok := m.Settings[key]; ok {
		return value
	}
	return nil
}

// Set mocks config.Set
func (m *MockConfig) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Settings[key] = value
}

// GetAll mocks config.GetAll
func (m *MockConfig) GetAll() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.Settings
}

// MockAPIClient mocks API interactions for testing
type MockAPIClient struct {
	Responses       map[string]string
	Errors          map[string]error
	CallCount       map[string]int
	DefaultResponse string
	DefaultError    error
	mu              sync.Mutex
}

// NewMockAPIClient creates a new mock API client
func NewMockAPIClient() *MockAPIClient {
	return &MockAPIClient{
		Responses:       make(map[string]string),
		Errors:          make(map[string]error),
		CallCount:       make(map[string]int),
		DefaultResponse: "",
		DefaultError:    nil,
	}
}

// SendToGemini mocks utils.SendToGemini
func (m *MockAPIClient) SendToGemini(contextData map[string]map[string]string, apiKey string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a key from the context data
	var key string
	if len(contextData) > 0 {
		for file := range contextData {
			key = file
			break
		}
	}

	// Increment call count
	m.CallCount[key] = m.CallCount[key] + 1

	// Check if we have a predefined response for this key
	if response, ok := m.Responses[key]; ok {
		if err, ok := m.Errors[key]; ok {
			return response, err
		}
		return response, nil
	}

	return m.DefaultResponse, m.DefaultError
}

// MockFileSystem mocks filesystem operations for testing
type MockFileSystem struct {
	FileContent map[string]string
	FileInfo    map[string]bool // true = directory, false = file
	mu          sync.Mutex
}

// NewMockFileSystem creates a new mock filesystem
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		FileContent: make(map[string]string),
		FileInfo:    make(map[string]bool),
	}
}

// ReadFile mocks reading a file
func (m *MockFileSystem) ReadFile(path string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if content, ok := m.FileContent[path]; ok {
		return content, nil
	}

	return "", errors.New("file not found: " + path)
}

// WriteFile mocks writing to a file
func (m *MockFileSystem) WriteFile(path, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.FileContent[path] = content
	m.FileInfo[path] = false
	return nil
}

// Exists mocks checking if a file exists
func (m *MockFileSystem) Exists(path string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.FileInfo[path]
	return ok, nil
}

// IsDir mocks checking if a path is a directory
func (m *MockFileSystem) IsDir(path string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if isDir, ok := m.FileInfo[path]; ok {
		return isDir, nil
	}

	return false, errors.New("path not found: " + path)
}

// CreateDir mocks creating a directory
func (m *MockFileSystem) CreateDir(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.FileInfo[path] = true
	return nil
}

// MockProgressReporter provides a mock implementation of progress reporting
type MockProgressReporter struct {
	Reports      []string
	StartCalled  bool
	FinishCalled bool
	ErrorCount   int
	mu           sync.Mutex
}

// NewMockProgressReporter creates a new mock progress reporter
func NewMockProgressReporter() *MockProgressReporter {
	return &MockProgressReporter{
		Reports:      make([]string, 0),
		StartCalled:  false,
		FinishCalled: false,
		ErrorCount:   0,
	}
}

// Start mocks starting progress reporting
func (m *MockProgressReporter) Start(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StartCalled = true
	m.Reports = append(m.Reports, "START: "+message)
}

// Update mocks updating progress
func (m *MockProgressReporter) Update(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Reports = append(m.Reports, "UPDATE: "+message)
}

// Finish mocks finishing progress reporting
func (m *MockProgressReporter) Finish(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.FinishCalled = true
	m.Reports = append(m.Reports, "FINISH: "+message)
}

// Error mocks reporting an error
func (m *MockProgressReporter) Error(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ErrorCount++
	m.Reports = append(m.Reports, "ERROR: "+message)
}
