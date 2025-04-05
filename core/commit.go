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
			utils.Debug("Root folder to commit in: " + rootFolder.Name)

			err := git.CommitBatch(rootFolder, env...)
			if err != nil {
				utils.Error("Failed to commit batch: " + err.Error())
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, Error: %s", rootFolder.Name, err.Error()))
				mu.Unlock()
				return
			}
		}(rootFolder)
	}

	rootFolderWg.Wait()

	if len(errors) > 0 {
		utils.Error("Batch commit completed with errors")
		utils.Debug("Errors: " + fmt.Sprint(errors))
		return fmt.Errorf("one or more errors occurred while committing files: %v", errors)
	}

	output.Clear()
	utils.Info("Batch commit completed successfully and output cleared")
	return nil
}

func CommitOneRoot(rootFolderName string, env ...[]string) error {
	rootFolder := output.GetFolder(rootFolderName)
	if len(rootFolder.Files) == 0 {
		utils.Error("Root folder not found or has no files: " + rootFolderName)
		return fmt.Errorf("root folder not found or has no files: %s", rootFolderName)
	}

	err := git.CommitBatch(rootFolder, env...)
	if err != nil {
		utils.Error("Failed to commit batch: " + err.Error())
		return fmt.Errorf("failed to commit batch: %s", err.Error())
	}

	utils.Info("Batch commit completed successfully for root folder: " + rootFolderName)
	return nil
}
