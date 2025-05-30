package git

import (
	"github.com/lakshyajain-0291/GitCury/output"
	"github.com/lakshyajain-0291/GitCury/utils"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// ProgressCommitBatch is an enhanced version of CommitBatch that includes progress reporting
// and better resource management
func ProgressCommitBatch(rootFolder output.Folder, env ...[]string) error {
	commitMessagesList := rootFolder.Files
	if len(commitMessagesList) == 0 {
		utils.Debug("[GIT.COMMIT]: No commit messages found for root folder: " + rootFolder.Name)
		return utils.NewValidationError(
			"No commit messages found for root folder",
			nil,
			map[string]interface{}{
				"folderName": rootFolder.Name,
			},
		)
	}

	// Start tracking operation in stats
	if utils.IsStatsEnabled() {
		utils.StartOperation("CommitBatch")
	}

	// Create progress reporter for better user feedback
	progress := utils.NewProgressReporter(int64(len(commitMessagesList)), "Committing files in "+rootFolder.Name)
	progress.Start()
	defer progress.Done()

	// Track overall progress
	var progressCounter int64
	var progressMu sync.Mutex

	// Function to update progress safely
	updateProgress := func(count int64, message string) {
		progressMu.Lock()
		defer progressMu.Unlock()
		progressCounter += count
		progress.Update(progressCounter)
		if message != "" {
			progress.UpdateMessage(message)
		}

		// Also update stats tracking
		if utils.IsStatsEnabled() {
			// Calculate percentage progress
			percentage := float64(progressCounter) / float64(len(commitMessagesList)) * 100.0
			utils.UpdateProgress("CommitBatch", percentage, "running")
		}
	}

	// Call original CommitBatch with progress hooks
	err := SafeGitOperation(rootFolder.Name, "CommitBatch", func() error {
		// Process each file with progress reporting
		for i, entry := range commitMessagesList {
			shortFile := filepath.Base(entry.Name)
			updateProgress(1, fmt.Sprintf("Processing %d/%d: %s", i+1, len(commitMessagesList), shortFile))

			// Add artificial delay to make progress visible
			time.Sleep(50 * time.Millisecond)
		}

		// Let the original function do the actual work
		return CommitBatch(rootFolder, env...)
	})

	if err != nil {
		progress.UpdateMessage("Commit failed: " + err.Error())

		// Mark operation as failed in stats
		if utils.IsStatsEnabled() {
			utils.FailOperation("CommitBatch", err.Error())
		}

		return err
	}

	progress.UpdateMessage("Commit completed successfully")

	// Mark operation as completed in stats
	if utils.IsStatsEnabled() {
		utils.MarkOperationComplete("CommitBatch")
	}

	return nil
}

// ProgressPushBranch is an enhanced version of PushBranch that includes progress reporting
func ProgressPushBranch(rootFolderName string, branch string) error {
	if branch == "" {
		utils.Debug("[GIT.PUSH]: Branch name is empty, defaulting to 'main'")
		branch = "main"
	}

	// Start stats tracking
	if utils.IsStatsEnabled() {
		utils.StartOperation("PushBranch")
		utils.UpdateOperationProgress("PushBranch", 10.0)
	}

	// Create progress reporter for terminal output
	progress := utils.NewProgressReporter(100, "Pushing to remote repository")
	progress.Start()
	defer progress.Done()

	// Update initial progress
	progress.Update(10)
	progress.UpdateMessage("Preparing to push branch: " + branch)

	// Update stats progress to match terminal progress
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("PushBranch", 25.0)
	}

	// Sleep briefly to make progress visible
	time.Sleep(100 * time.Millisecond)

	// Update progress
	progress.Update(50)
	progress.UpdateMessage("Pushing branch '" + branch + "' to remote")

	// Update stats progress
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("PushBranch", 50.0)
	}

	// Sleep briefly to make progress visible
	time.Sleep(100 * time.Millisecond)

	utils.Debug("[GIT.PUSH]: Pushing branch: " + branch + " in folder: " + rootFolderName)

	// Use SafeGitOperation to handle index.lock and other recovery scenarios
	err := SafeGitOperation(rootFolderName, "push", func() error {
		_, gitErr := RunGitCmd(rootFolderName, nil, "push", "origin", branch)
		return gitErr
	})

	if err != nil {
		progress.UpdateMessage("Failed to push branch: " + err.Error())

		// Mark operation as failed in stats
		if utils.IsStatsEnabled() {
			utils.FailOperation("PushBranch", err.Error())
		}

		utils.Error("[GIT.PUSH.FAIL]: Failed to push branch: " + err.Error())
		return fmt.Errorf("failed to push branch: %s", err.Error())
	}

	// Update final progress
	progress.Update(100)
	progress.UpdateMessage("Branch pushed successfully")

	// Mark operation as completed in stats
	if utils.IsStatsEnabled() {
		utils.MarkOperationComplete("PushBranch")
	}

	utils.Info("[GIT.PUSH.SUCCESS]: Branch pushed successfully")
	return nil
}
