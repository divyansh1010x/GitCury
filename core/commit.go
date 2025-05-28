package core

import (
	"GitCury/config"
	"GitCury/git"
	"GitCury/output"
	"GitCury/utils"
	"fmt"
	"strings"
	"time"
)

func CommitAllRoots(env ...[]string) error {
	rootFolders := output.GetAll().Folders
	if len(rootFolders) == 0 {
		utils.Warning("[" + config.Aliases.Commit + "]: No root folders with changes to commit")
		return nil
	}

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("CommitAllRoots", 20.0)
	}

	// Determine optimal worker count based on available folders
	workerCount := 3
	if len(rootFolders) < workerCount {
		workerCount = len(rootFolders)
	}

	// Create worker pool for parallel execution with limited concurrency
	pool := utils.NewWorkerPool(workerCount)
	utils.Debug("[" + config.Aliases.Commit + "]: Created worker pool with " + fmt.Sprint(workerCount) + " workers for " + fmt.Sprint(len(rootFolders)) + " folders")

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("CommitAllRoots", 30.0)
	}

	// Submit commit tasks for each root folder
	for _, rootFolder := range rootFolders {
		folder := rootFolder // Create local copy to avoid closure issues
		taskName := "CommitRoot:" + folder.Name

		pool.Submit(taskName, 2*time.Minute, func() error {
			utils.Debug("[" + config.Aliases.Commit + "]: Processing root folder: " + folder.Name)

			if len(folder.Files) == 0 {
				utils.Debug("[" + config.Aliases.Commit + "]: No files to commit in folder: " + folder.Name)
				return nil
			}

			err := git.ProgressCommitBatch(folder, env...)
			if err != nil {
				// Extract file information if available in the error
				fileInfo := folder.Name
				if structErr, ok := err.(*utils.StructuredError); ok && structErr.ProcessedFile != "" {
					fileInfo = structErr.ProcessedFile
				}

				utils.Error("["+config.Aliases.Commit+".FAIL]: Failed to commit batch for folder '"+folder.Name+"' - "+err.Error(), fileInfo)
				return utils.NewGitError(
					"Failed to commit changes in folder",
					err,
					map[string]interface{}{
						"folder":    folder.Name,
						"fileCount": len(folder.Files),
					},
					fileInfo,
				)
			}

			utils.Debug("[" + config.Aliases.Commit + "]: Successfully committed changes in folder: " + folder.Name)
			return nil
		})
	}

	// Wait for all commit tasks to complete
	errors := pool.Wait()

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("CommitAllRoots", 80.0)
	}

	if len(errors) > 0 {
		errorDetails := make([]string, 0, len(errors))
		filesList := make([]string, 0, len(errors))

		for _, err := range errors {
			errorDetails = append(errorDetails, err.Error())

			// Extract file information if available
			if structErr, ok := err.(*utils.StructuredError); ok && structErr.ProcessedFile != "" {
				filesList = append(filesList, structErr.ProcessedFile)
			}
		}

		filesInfo := "multiple_folders"
		if len(filesList) > 0 {
			filesInfo = strings.Join(filesList, ", ")
		}

		utils.Error("["+config.Aliases.Commit+".FAIL]: Batch commit completed with "+fmt.Sprint(len(errors))+" errors", filesInfo)
		utils.Debug("[" + config.Aliases.Commit + ".FAIL]: Errors encountered: " + strings.Join(errorDetails, "; "))

		return utils.NewGitError(
			fmt.Sprintf("%d errors occurred during batch commit", len(errors)),
			fmt.Errorf("multiple commit errors"),
			map[string]interface{}{
				"errorCount": len(errors),
				"errors":     errorDetails,
			},
			filesInfo,
		)
	}

	output.Clear()

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("CommitAllRoots", 95.0)
	}

	utils.Success("[" + config.Aliases.Commit + ".SUCCESS]: Batch commit completed successfully. Output cleared.")
	return nil
}

func CommitOneRoot(rootFolderName string, env ...[]string) error {
	rootFolder := output.GetFolder(rootFolderName)
	if len(rootFolder.Files) == 0 {
		utils.Error("["+config.Aliases.Commit+".FAIL]: Root folder '"+rootFolderName+"' not found or contains no files.", rootFolderName)
		return utils.NewValidationError(
			"Root folder not found or has no files",
			nil,
			map[string]interface{}{
				"folderName": rootFolderName,
			},
			rootFolderName,
		)
	}

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("CommitOneRoot", 30.0)
	}

	utils.Debug("[" + config.Aliases.Commit + "]: Targeting root folder for commit: " + rootFolderName)

	err := git.ProgressCommitBatch(rootFolder, env...)
	if err != nil {
		// Extract file information if available in the error
		fileInfo := rootFolderName
		if structErr, ok := err.(*utils.StructuredError); ok && structErr.ProcessedFile != "" {
			fileInfo = structErr.ProcessedFile
		}

		utils.Error("["+config.Aliases.Commit+".FAIL]: Failed to commit batch for folder '"+rootFolderName+"' - "+err.Error(), fileInfo)
		return utils.NewGitError(
			"Failed to commit batch for folder",
			err,
			map[string]interface{}{
				"folder": rootFolderName,
			},
			fileInfo,
		)
	}

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("CommitOneRoot", 90.0)
	}

	utils.Success("[" + config.Aliases.Commit + ".SUCCESS]: Batch commit completed successfully for root folder: " + rootFolderName)
	return nil
}
