package git

import (
	"GitCury/interfaces"
	"GitCury/output"
	"time"
)

// DefaultGitRunner implements the GitRunner interface with real Git operations
type DefaultGitRunner struct{}

// Ensure DefaultGitRunner implements GitRunner interface
var _ interfaces.GitRunner = (*DefaultGitRunner)(nil)

// RunGitCmd executes a Git command in the specified directory
func (d *DefaultGitRunner) RunGitCmd(dir string, envVars map[string]string, args ...string) (string, error) {
	return RunGitCmd(dir, envVars, args...)
}

// RunGitCmdWithTimeout executes a Git command with a timeout
func (d *DefaultGitRunner) RunGitCmdWithTimeout(dir string, envVars map[string]string, timeout time.Duration, args ...string) (string, error) {
	return RunGitCmdWithTimeout(dir, envVars, timeout, args...)
}

// CommitBatch commits a batch of files for a folder
func (d *DefaultGitRunner) CommitBatch(folder interfaces.Folder, env ...[]string) error {
	outputFolder := interfaceToOutput(folder)
	return CommitBatch(outputFolder, env...)
}

// GetChangedFiles gets changed files from the specified root folders
func (d *DefaultGitRunner) GetChangedFiles(rootFolders []string, maxConcurrency int, env ...[]string) ([]interfaces.Folder, error) {
	folders, err := GetChangedFiles(rootFolders, maxConcurrency, env...)
	if err != nil {
		return nil, err
	}
	return outputFoldersToInterface(folders), nil
}

// Status gets the status of the specified root paths
func (d *DefaultGitRunner) Status(rootPaths []string) ([]interfaces.Folder, error) {
	folders, err := Status(rootPaths)
	if err != nil {
		return nil, err
	}
	return outputFoldersToInterface(folders), nil
}

// ProcessOneFile processes a single file with the given commit message
func (d *DefaultGitRunner) ProcessOneFile(filePath, commitMessage string, env ...[]string) error {
	return ProcessOneFile(filePath, commitMessage, env...)
}

// GetDiff gets the diff for a specific file
func (d *DefaultGitRunner) GetDiff(filePath string, env ...[]string) (string, error) {
	// For now, use empty rootFolder since GetFileDiff doesn't support env vars yet
	return GetFileDiff(filePath, "")
}

// IsGitRepository checks if the given path is a Git repository
func (d *DefaultGitRunner) IsGitRepository(path string) bool {
	return IsGitRepository(path)
}

// GetGitConfigValue gets a Git configuration value
func (d *DefaultGitRunner) GetGitConfigValue(key string, env ...[]string) (string, error) {
	return GetGitConfigValue(key, env...)
}

// SetGitConfigValue sets a Git configuration value
func (d *DefaultGitRunner) SetGitConfigValue(key, value string, env ...[]string) error {
	return SetGitConfigValue(key, value, env...)
}

// ProgressCommitBatch is an enhanced version with progress reporting
func (d *DefaultGitRunner) ProgressCommitBatch(folder interfaces.Folder, env ...[]string) error {
	outputFolder := interfaceToOutput(folder)
	return ProgressCommitBatch(outputFolder, env...)
}

// ProgressPushBranch is an enhanced version with progress reporting
func (d *DefaultGitRunner) ProgressPushBranch(rootFolderName string, branch string) error {
	return ProgressPushBranch(rootFolderName, branch)
}

// GetAllChangedFiles gets all changed files for a directory
func (d *DefaultGitRunner) GetAllChangedFiles(dir string) ([]string, error) {
	return GetAllChangedFiles(dir)
}

// BatchProcessGetMessages processes multiple files and generates commit messages
func (d *DefaultGitRunner) BatchProcessGetMessages(allChangedFiles []string, rootFolder string) error {
	return BatchProcessGetMessages(allChangedFiles, rootFolder)
}

// BatchProcessWithEmbeddings processes files using embeddings and clustering
func (d *DefaultGitRunner) BatchProcessWithEmbeddings(allChangedFiles []string, rootFolder string, numClusters int) error {
	return BatchProcessWithEmbeddings(allChangedFiles, rootFolder, numClusters)
}

// Conversion functions between output and interface types
func outputToInterface(folder output.Folder) interfaces.Folder {
	var files []interfaces.FileEntry
	for _, file := range folder.Files {
		files = append(files, interfaces.FileEntry{
			Name:    file.Name,
			Message: file.Message,
		})
	}
	return interfaces.Folder{
		Name:  folder.Name,
		Files: files,
	}
}

func interfaceToOutput(folder interfaces.Folder) output.Folder {
	var files []output.FileEntry
	for _, file := range folder.Files {
		files = append(files, output.FileEntry{
			Name:    file.Name,
			Message: file.Message,
		})
	}
	return output.Folder{
		Name:  folder.Name,
		Files: files,
	}
}

func outputFoldersToInterface(folders []output.Folder) []interfaces.Folder {
	var result []interfaces.Folder
	for _, folder := range folders {
		result = append(result, outputToInterface(folder))
	}
	return result
}
