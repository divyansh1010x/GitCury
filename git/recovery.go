package git

import (
	"github.com/lakshyajain-0291/gitcury/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GitOperationResult represents the result of a git operation with recovery information
type GitOperationResult struct {
	Success      bool
	Error        error
	RecoveryPath string
	Message      string
}

// CheckRepositoryHealth checks if a git repository is in a healthy state
// and returns any issues found
func CheckRepositoryHealth(dir string) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return utils.NewValidationError(
			"Directory does not exist",
			err,
			map[string]interface{}{
				"directory": dir,
			},
		)
	}

	// Check if this is a git repository
	_, err := RunGitCmdWithTimeout(dir, nil, 5*time.Second, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return utils.NewGitError(
			"Not a git repository",
			err,
			map[string]interface{}{
				"directory":  dir,
				"suggestion": "Initialize a git repository with 'git init'",
			},
		)
	}

	// Check for uncommitted changes that might cause conflicts
	output, err := RunGitCmdWithTimeout(dir, nil, 5*time.Second, "status", "--porcelain")
	if err != nil {
		return utils.NewGitError(
			"Failed to get repository status",
			err,
			map[string]interface{}{
				"directory": dir,
			},
		)
	}

	if strings.TrimSpace(output) != "" {
		// There are uncommitted changes, but this is not an error
		utils.Debug("[GIT.HEALTH]: Repository has uncommitted changes: " + dir)
	}

	// Check if the git index is locked
	indexLockPath := filepath.Join(dir, ".git", "index.lock")
	if _, err := os.Stat(indexLockPath); err == nil {
		// Index is locked, this might indicate a problem
		lockFileInfo, err := os.Stat(indexLockPath)
		if err == nil {
			lockAge := time.Since(lockFileInfo.ModTime())

			// If lock file is older than 10 minutes, it's likely stale
			if lockAge > 10*time.Minute {
				utils.Warning("[GIT.HEALTH]: Found stale index.lock file (age: " + lockAge.String() + ")")
				return utils.NewGitError(
					"Repository index is locked (stale lock)",
					nil,
					map[string]interface{}{
						"directory":  dir,
						"lockFile":   indexLockPath,
						"lockAge":    lockAge.String(),
						"suggestion": "Remove the lock file with 'rm " + indexLockPath + "'",
					},
				)
			}

			return utils.NewGitError(
				"Repository index is locked",
				nil,
				map[string]interface{}{
					"directory":  dir,
					"lockFile":   indexLockPath,
					"lockAge":    lockAge.String(),
					"suggestion": "Wait for the current git operation to complete or remove the lock file",
				},
			)
		}
	}

	// Check if we can access the git config
	_, err = RunGitCmdWithTimeout(dir, nil, 5*time.Second, "config", "--local", "--list")
	if err != nil {
		return utils.NewGitError(
			"Failed to access git config",
			err,
			map[string]interface{}{
				"directory":  dir,
				"suggestion": "The .git directory might be corrupted, try reinitializing the repository",
			},
		)
	}

	return nil
}

// RecoverFromGitError attempts to recover from common git errors
func RecoverFromGitError(dir string, err error) GitOperationResult {
	if err == nil {
		return GitOperationResult{
			Success: true,
			Message: "No error to recover from",
		}
	}

	errMsg := err.Error()

	// Check for index.lock issues
	if strings.Contains(errMsg, "index.lock") {
		indexLockPath := filepath.Join(dir, ".git", "index.lock")
		if _, statErr := os.Stat(indexLockPath); statErr == nil {
			// Index lock exists, try to remove it
			utils.Warning("[GIT.RECOVERY]: Found index.lock file, attempting to remove it")

			if rmErr := os.Remove(indexLockPath); rmErr == nil {
				return GitOperationResult{
					Success:      true,
					RecoveryPath: "index_lock_removed",
					Message:      "Removed stale index.lock file",
				}
			} else {
				return GitOperationResult{
					Success:      false,
					Error:        rmErr,
					RecoveryPath: "index_lock_removal_failed",
					Message:      "Failed to remove stale index.lock file",
				}
			}
		}
	}

	// Check for conflicts
	if strings.Contains(errMsg, "conflict") || strings.Contains(errMsg, "CONFLICT") {
		utils.Warning("[GIT.RECOVERY]: Detected conflict in git operation")
		return GitOperationResult{
			Success:      false,
			Error:        err,
			RecoveryPath: "conflict_detected",
			Message:      "Git operation failed due to conflicts. Manual resolution required.",
		}
	}

	// Check for permission issues
	if strings.Contains(errMsg, "Permission denied") {
		utils.Warning("[GIT.RECOVERY]: Detected permission issue in git operation")
		return GitOperationResult{
			Success:      false,
			Error:        err,
			RecoveryPath: "permission_denied",
			Message:      "Git operation failed due to permission issues. Check file permissions.",
		}
	}

	// No specific recovery path identified
	return GitOperationResult{
		Success:      false,
		Error:        err,
		RecoveryPath: "unknown_error",
		Message:      "No automatic recovery available for this git error",
	}
}

// SafeGitOperation executes a git operation with built-in recovery mechanisms
func SafeGitOperation(dir string, operation string, fn func() error) error {
	// First check repository health
	if err := CheckRepositoryHealth(dir); err != nil {
		utils.Warning("[GIT.SAFE]: Repository health check failed: " + err.Error())

		// If the directory doesn't exist, no recovery is possible
		if os.IsNotExist(err) {
			return err
		}

		// Try to recover from repository issues
		result := RecoverFromGitError(dir, err)
		if !result.Success {
			utils.Error("[GIT.SAFE]: Failed to recover from repository issue: " + result.Message)
			return utils.NewGitError(
				"Repository is in an unhealthy state",
				err,
				map[string]interface{}{
					"directory":    dir,
					"operation":    operation,
					"recoveryPath": result.RecoveryPath,
					"message":      result.Message,
				},
			)
		}

		utils.Info("[GIT.SAFE]: Successfully recovered from repository issue: " + result.Message)
	}

	// Execute the git operation
	err := fn()
	if err != nil {
		utils.Error("[GIT.SAFE]: Git operation failed: " + err.Error())

		// Try to recover from the error
		result := RecoverFromGitError(dir, err)
		if !result.Success {
			return utils.NewGitError(
				fmt.Sprintf("Git operation '%s' failed", operation),
				err,
				map[string]interface{}{
					"directory":    dir,
					"operation":    operation,
					"recoveryPath": result.RecoveryPath,
					"message":      result.Message,
				},
			)
		}

		utils.Info("[GIT.SAFE]: Successfully recovered from git error: " + result.Message)

		// Retry the operation after recovery
		retryErr := fn()
		if retryErr != nil {
			return utils.NewGitError(
				fmt.Sprintf("Git operation '%s' failed even after recovery", operation),
				retryErr,
				map[string]interface{}{
					"directory":    dir,
					"operation":    operation,
					"recoveryPath": result.RecoveryPath,
					"message":      "Recovery was successful but operation still failed on retry",
				},
			)
		}
	}

	return nil
}
