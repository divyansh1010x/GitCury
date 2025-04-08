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
		utils.Error("[DEPLOY]: ❌ Invalid or missing root_folders configuration")
		return fmt.Errorf("invalid or missing root_folders configuration")
	}

	var rootFolderWg sync.WaitGroup
	var mu sync.Mutex
	var errors []string

	for _, rootFolder := range rootFolders {
		rootFolderStr, ok := rootFolder.(string)
		if !ok {
			utils.Error("[DEPLOY]: ⚠️ Invalid root folder type")
			continue
		}

		rootFolderWg.Add(1)

		go func(folder string) {
			defer rootFolderWg.Done()
			utils.Debug("[DEPLOY]: 📂 Root folder to push: " + folder)

			err := git.PushBranch(folder, branchName)
			if err != nil {
				utils.Error("[DEPLOY]: ❌ Failed to push branch for folder '" + folder + "' - " + err.Error())
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, Error: %s", folder, err.Error()))
				mu.Unlock()
				return
			}
			utils.Success("[DEPLOY]: ✅ Successfully pushed branch for folder: " + folder)
		}(rootFolderStr)
	}

	rootFolderWg.Wait()
	if len(errors) > 0 {
		utils.Error("[DEPLOY]: ❌ Errors occurred during push operation")
		return fmt.Errorf("one or more errors occurred while pushing branches: %v", errors)
	}

	utils.Success("[DEPLOY]: 🌐 Push operation for all roots completed successfully")
	return nil
}

func PushOneRoot(rootFolderName, branchName string) error {
	utils.Debug("[DEPLOY]: 📂 Targeting root folder for push: " + rootFolderName)

	err := git.PushBranch(rootFolderName, branchName)
	if err != nil {
		utils.Error("[DEPLOY]: ❌ Failed to push branch for folder '" + rootFolderName + "' - " + err.Error())
		return fmt.Errorf("failed to push branch: %s", err.Error())
	}

	utils.Success("[DEPLOY]: ✅ Push operation for root folder '" + rootFolderName + "' completed successfully")
	return nil
}
