package core

import (
	"GitCury/config"
	"GitCury/git"
	"GitCury/utils"
	"fmt"
	"sync"
)

func PushAllRoots(branchName string) error {
	rootFolders, ok := config.Get("root_folders").([]interface{})
	if !ok {
		utils.Error("["+config.Aliases.Push+"]: ❌ Invalid or missing root_folders configuration", "config")
		return utils.NewValidationError(
			"Invalid or missing root_folders configuration",
			nil,
			map[string]interface{}{
				"configKey": "root_folders",
			},
			"config",
		)
	}
	
	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("PushAllRoots", 20.0)
	}

	var rootFolderWg sync.WaitGroup
	var mu sync.Mutex
	var errors []string
	
	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("PushAllRoots", 40.0)
	}

	for _, rootFolder := range rootFolders {
		rootFolderStr, ok := rootFolder.(string)
		if !ok {
			utils.Error("["+config.Aliases.Push+"]: ⚠️ Invalid root folder type", "config")
			continue
		}

		rootFolderWg.Add(1)

		go func(folder string) {
			defer rootFolderWg.Done()
			utils.Debug("[" + config.Aliases.Push + "]: 📂 Root folder to push: " + folder)

			err := git.ProgressPushBranch(folder, branchName)
			if err != nil {
				// Extract file information if available in the error
				fileInfo := folder
				if structErr, ok := err.(*utils.StructuredError); ok && structErr.ProcessedFile != "" {
					fileInfo = structErr.ProcessedFile
				}

				utils.Error("["+config.Aliases.Push+"]: ❌ Failed to push branch for folder '"+folder+"' - "+err.Error(), fileInfo)
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, File: %s, Error: %s", folder, fileInfo, err.Error()))
				mu.Unlock()
				return
			}
			utils.Success("[" + config.Aliases.Push + "]: ✅ Successfully pushed branch for folder: " + folder)
		}(rootFolderStr)
	}

	rootFolderWg.Wait()
	
	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("PushAllRoots", 70.0)
	}
	
	if len(errors) > 0 {
		filesInfo := "multiple_folders"
		utils.Error("["+config.Aliases.Push+"]: ❌ Errors occurred during push operation", filesInfo)
		return utils.NewGitError(
			"One or more errors occurred while pushing branches",
			fmt.Errorf("multiple push errors"),
			map[string]interface{}{
				"errors": errors,
			},
			filesInfo,
		)
	}

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("PushAllRoots", 90.0)
	}

	utils.Success("[" + config.Aliases.Push + "]: 🌐 Push operation for all roots completed successfully")
	return nil
}

func PushOneRoot(rootFolderName, branchName string) error {
	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("PushOneRoot", 30.0)
	}
	
	utils.Debug("[" + config.Aliases.Push + "]: 📂 Targeting root folder for push: " + rootFolderName)

	err := git.ProgressPushBranch(rootFolderName, branchName)
	if err != nil {
		// Extract file information if available in the error
		fileInfo := rootFolderName
		if structErr, ok := err.(*utils.StructuredError); ok && structErr.ProcessedFile != "" {
			fileInfo = structErr.ProcessedFile
		}

		utils.Error("["+config.Aliases.Push+"]: ❌ Failed to push branch for folder '"+rootFolderName+"' - "+err.Error(), fileInfo)
		return utils.NewGitError(
			"Failed to push branch for folder",
			err,
			map[string]interface{}{
				"folder": rootFolderName,
				"branch": branchName,
			},
			fileInfo,
		)
	}

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("PushOneRoot", 80.0)
	}

	utils.Success("[" + config.Aliases.Push + "]: ✅ Push operation for root folder '" + rootFolderName + "' completed successfully")
	return nil
}
