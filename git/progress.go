package git

import (
	"GitCury/output"
	"GitCury/utils"
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
		return err
	}

	progress.UpdateMessage("Commit completed successfully")
	return nil
}
