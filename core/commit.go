package core

import (
	"GitCury/git"
	"GitCury/interfaces"
	"GitCury/output"
	"GitCury/utils"
	"fmt"
	"strings"
	"time"
)

// GitRunnerInstance allows dependency injection for testing
var GitRunnerInstance interfaces.GitRunner

// init initializes the default Git runner
func init() {
	if GitRunnerInstance == nil {
		GitRunnerInstance = &git.DefaultGitRunner{}
	}
}

// SetGitRunner allows injecting a custom GitRunner (used in tests)
func SetGitRunner(runner interfaces.GitRunner) {
	GitRunnerInstance = runner
}

func CommitAllRoots(env ...[]string) error {
	rootFolders := output.GetAll().Folders
	if len(rootFolders) == 0 {
		utils.Warning("No root folders with changes to commit")
		return nil
	}

	// Determine optimal worker count based on available folders
	workerCount := 3
	if len(rootFolders) < workerCount {
		workerCount = len(rootFolders)
	}

	// Create worker pool for parallel execution with limited concurrency
	pool := utils.NewWorkerPool(workerCount)

	// Submit commit tasks for each root folder
	for _, rootFolder := range rootFolders {
		folder := rootFolder // Create local copy to avoid closure issues
		taskName := "CommitRoot:" + folder.Name

		pool.Submit(taskName, 2*time.Minute, func() error {
			if len(folder.Files) == 0 {
				return nil
			}

			err := GitRunnerInstance.ProgressCommitBatch(outputToInterface(folder), env...)
			if err != nil {
				// Extract file information if available in the error
				fileInfo := folder.Name
				if structErr, ok := err.(*utils.StructuredError); ok && structErr.ProcessedFile != "" {
					fileInfo = structErr.ProcessedFile
				}

				utils.Error("Failed to commit batch for folder '"+folder.Name+"' - "+err.Error(), fileInfo)
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

			return nil
		})
	}

	// Wait for all commit tasks to complete
	errors := pool.Wait()

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

		utils.Error("Batch commit completed with "+fmt.Sprint(len(errors))+" errors", filesInfo)

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
	utils.Success("✅ Batch commit completed successfully. Output cleared.")
	return nil
}

func CommitOneRoot(rootFolderName string, env ...[]string) error {
	rootFolder := output.GetFolder(rootFolderName)
	if len(rootFolder.Files) == 0 {
		utils.Error("Root folder '"+rootFolderName+"' not found or contains no files.", rootFolderName)
		return utils.NewValidationError(
			"Root folder not found or has no files",
			nil,
			map[string]interface{}{
				"folderName": rootFolderName,
			},
			rootFolderName,
		)
	}

	err := GitRunnerInstance.ProgressCommitBatch(outputToInterface(rootFolder), env...)
	if err != nil {
		// Extract file information if available in the error
		fileInfo := rootFolderName
		if structErr, ok := err.(*utils.StructuredError); ok && structErr.ProcessedFile != "" {
			fileInfo = structErr.ProcessedFile
		}

		utils.Error("Failed to commit batch for folder '"+rootFolderName+"' - "+err.Error(), fileInfo)
		return utils.NewGitError(
			"Failed to commit batch for folder",
			err,
			map[string]interface{}{
				"folder": rootFolderName,
			},
			fileInfo,
		)
	}

	utils.Success("✅ Batch commit completed successfully for root folder: " + rootFolderName)
	return nil
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
