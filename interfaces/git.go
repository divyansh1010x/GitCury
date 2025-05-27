package interfaces

import (
	"GitCury/output"
	"time"
)

// GitRunner defines the interface for git operations
type GitRunner interface {
	RunGitCmd(dir string, envVars map[string]string, args ...string) (string, error)
	RunGitCmdWithTimeout(dir string, envVars map[string]string, timeout time.Duration, args ...string) (string, error)
	CommitBatch(folder output.Folder, env ...[]string) error
	GetChangedFiles(rootFolders []string, maxConcurrency int, env ...[]string) ([]output.Folder, error)
	Status(rootPaths []string) ([]output.Folder, error)
	ProcessOneFile(filePath, commitMessage string, env ...[]string) error
	GetDiff(filePath string, env ...[]string) (string, error)
	IsGitRepository(path string) bool
	GetGitConfigValue(key string, env ...[]string) (string, error)
	SetGitConfigValue(key, value string, env ...[]string) error
}

// OutputManager defines the interface for output operations
type OutputManager interface {
	GetAll() output.OutputData
	GetFolder(rootFolder string) output.Folder
	RemoveFolder(name string)
	Clear()
	SaveToFile()
}
