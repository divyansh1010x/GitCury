package interfaces

import (
	"time"
)

// FileEntry represents a file with its commit message
type FileEntry struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

// Folder represents a folder containing files
type Folder struct {
	Name  string      `json:"name"`
	Files []FileEntry `json:"files"`
}

// OutputData represents the complete output structure
type OutputData struct {
	Folders []Folder `json:"folders"`
}

// GitRunner defines the interface for git operations
type GitRunner interface {
	RunGitCmd(dir string, envVars map[string]string, args ...string) (string, error)
	RunGitCmdWithTimeout(dir string, envVars map[string]string, timeout time.Duration, args ...string) (string, error)
	CommitBatch(folder Folder, env ...[]string) error
	GetChangedFiles(rootFolders []string, maxConcurrency int, env ...[]string) ([]Folder, error)
	Status(rootPaths []string) ([]Folder, error)
	ProcessOneFile(filePath, commitMessage string, env ...[]string) error
	GetDiff(filePath string, env ...[]string) (string, error)
	IsGitRepository(path string) bool
	GetGitConfigValue(key string, env ...[]string) (string, error)
	SetGitConfigValue(key, value string, env ...[]string) error
	// Message processing methods
	GetAllChangedFiles(dir string) ([]string, error)
	BatchProcessGetMessages(allChangedFiles []string, rootFolder string) error
	BatchProcessWithEmbeddings(allChangedFiles []string, rootFolder string, numClusters int) error
	// Progress tracking methods
	ProgressCommitBatch(folder Folder, env ...[]string) error
	ProgressPushBranch(rootFolderName string, branch string) error
}

// OutputManager defines the interface for output operations
type OutputManager interface {
	Set(file, rootFolder, commitMessage string)
	Get(file, rootFolder string) string
	GetAll() OutputData
	GetFolder(rootFolder string) Folder
	RemoveFolder(name string)
	Clear()
	SaveToFile()
}
