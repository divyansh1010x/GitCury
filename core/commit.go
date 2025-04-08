package core

import (
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
			utils.Debug("[SEAL]: Targeting root folder for commit: " + rootFolder.Name)

			err := git.CommitBatch(rootFolder, env...)
			if err != nil {
				utils.Error("[SEAL.FAIL]: Failed to commit batch for folder '" + rootFolder.Name + "' - " + err.Error())
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, Error: %s", rootFolder.Name, err.Error()))
				mu.Unlock()
				return
			}
		}(rootFolder)
	}

	rootFolderWg.Wait()

	if len(errors) > 0 {
		utils.Error("[SEAL.FAIL]: Batch commit completed with errors")
		utils.Debug("[SEAL.FAIL]: Errors encountered: " + fmt.Sprint(errors))
		return fmt.Errorf("one or more errors occurred while committing files: %v", errors)
	}

	output.Clear()
	utils.Success("[SEAL.SUCCESS]: Batch commit completed successfully. Output cleared.")
	return nil
}

func CommitOneRoot(rootFolderName string, env ...[]string) error {
	rootFolder := output.GetFolder(rootFolderName)
	if len(rootFolder.Files) == 0 {
		utils.Error("[SEAL.FAIL]: Root folder '" + rootFolderName + "' not found or contains no files.")
		return fmt.Errorf("root folder not found or has no files: %s", rootFolderName)
	}

	utils.Debug("[SEAL]: Targeting root folder for commit: " + rootFolderName)

	err := git.CommitBatch(rootFolder, env...)
	if err != nil {
		utils.Error("[SEAL.FAIL]: Failed to commit batch for folder '" + rootFolderName + "' - " + err.Error())
		return fmt.Errorf("failed to commit batch: %s", err.Error())
	}

	utils.Success("[SEAL.SUCCESS]: Batch commit completed successfully for root folder: " + rootFolderName)
	return nil
}
