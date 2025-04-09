package core

import (
	"GitCury/config"
	"GitCury/git"
	"GitCury/output"
	"GitCury/utils"
	"fmt"
	"sync"
)

func CommitAllRoots(env ...[]string) error {
	rootFolders := output.GetAll().Folders

	var rootFolderWg sync.WaitGroup
	var mu sync.Mutex
	var errors []string

	for _, rootFolder := range rootFolders {
		rootFolderWg.Add(1)

		go func(rootFolder output.Folder) {
			defer rootFolderWg.Done()
			utils.Debug("[" + config.Aliases.Commit + "]: Targeting root folder for commit: " + rootFolder.Name)

			err := git.CommitBatch(rootFolder, env...)
			if err != nil {
				utils.Error("[" + config.Aliases.Commit + ".FAIL]: Failed to commit batch for folder '" + rootFolder.Name + "' - " + err.Error())
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, Error: %s", rootFolder.Name, err.Error()))
				mu.Unlock()
				return
			}
		}(rootFolder)
	}

	rootFolderWg.Wait()

	if len(errors) > 0 {
		utils.Error("[" + config.Aliases.Commit + ".FAIL]: Batch commit completed with errors")
		utils.Debug("[" + config.Aliases.Commit + ".FAIL]: Errors encountered: " + fmt.Sprint(errors))
		return fmt.Errorf("one or more errors occurred while committing files: %v", errors)
	}

	output.Clear()
	utils.Success("[" + config.Aliases.Commit + ".SUCCESS]: Batch commit completed successfully. Output cleared.")
	return nil
}

func CommitOneRoot(rootFolderName string, env ...[]string) error {
	rootFolder := output.GetFolder(rootFolderName)
	if len(rootFolder.Files) == 0 {
		utils.Error("[" + config.Aliases.Commit + ".FAIL]: Root folder '" + rootFolderName + "' not found or contains no files.")
		return fmt.Errorf("root folder not found or has no files: %s", rootFolderName)
	}

	utils.Debug("[" + config.Aliases.Commit + "]: Targeting root folder for commit: " + rootFolderName)

	err := git.CommitBatch(rootFolder, env...)
	if err != nil {
		utils.Error("[" + config.Aliases.Commit + ".FAIL]: Failed to commit batch for folder '" + rootFolderName + "' - " + err.Error())
		return fmt.Errorf("failed to commit batch: %s", err.Error())
	}

	utils.Success("[" + config.Aliases.Commit + ".SUCCESS]: Batch commit completed successfully for root folder: " + rootFolderName)
	return nil
}
