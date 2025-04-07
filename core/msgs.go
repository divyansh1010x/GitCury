package core

import (
	"GitCury/config"
	"GitCury/git"
	"GitCury/output"
	"GitCury/utils"
	"fmt"
	"strconv"
	"sync"
)

func GetAllMsgs(numFiles ...int) error {
	defaultNumFiles := 10 // Default value
	if len(numFiles) == 0 {
		numFiles = append(numFiles, defaultNumFiles)
	}

	utils.Debug("Number of files to prepare commit messages for: " + strconv.Itoa(numFiles[0]))

	rootFolders, ok := config.Get("root_folders").([]interface{})
	if !ok {
		utils.Error("Invalid or missing root_folders configuration")
		return fmt.Errorf("invalid or missing root_folders configuration")
	}

	var rootFolderWg sync.WaitGroup
	var mu sync.Mutex
	var errors []string

	for _, rootFolder := range rootFolders {
		rootFolderStr, ok := rootFolder.(string)
		if !ok {
			utils.Error("Invalid root folder type")
			continue
		}

		rootFolderWg.Add(1)
		go func(folder string) {
			defer rootFolderWg.Done()

			utils.Debug("Root folder to get messages : " + folder)

			changedFiles, err := git.GetAllChangedFiles(folder)
			if err != nil {
				utils.Error("Failed to get changed files: " + err.Error())
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, Error: %s", folder, err.Error()))
				mu.Unlock()
				return
			}

			if len(changedFiles) == 0 {
				utils.Info("No changed files found")
				return
			}

			if len(changedFiles) > numFiles[0] {
				changedFiles = changedFiles[:numFiles[0]]
			}

			utils.Debug("Total files to process: " + strconv.Itoa(len(changedFiles)))

			err = git.BatchProcessGetMessages(changedFiles, folder)
			if err != nil {
				utils.Error("Batch processing failed for folder: " + folder + ", Error: " + err.Error())
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, Error: %s", folder, err.Error()))
				mu.Unlock()
			}
		}(rootFolderStr)
	}

	rootFolderWg.Wait()

	if len(errors) > 0 {
		utils.Error("Batch processing completed with errors")
		utils.Debug("Errors: " + fmt.Sprint(errors))
		return fmt.Errorf("one or more errors occurred while preparing commit messages")
	}

	utils.Info("Batch generation of all messages completed successfully")
	output.SaveToFile()
	return nil
}

func GetMsgsForRootFolder(folder string, numFiles ...int) error {
	if folder == "" {
		utils.Error("Root folder is empty")
		return fmt.Errorf("root folder is empty")
	}

	numFilesToCommit := 10 // Default value
	if len(numFiles) > 0 && numFiles[0] > 0 {
		utils.Debug("Using provided number of files to commit: " + strconv.Itoa(numFiles[0]))
		numFilesToCommit = numFiles[0]
	} else {
		if configValue := config.Get("numFilesToCommit"); configValue != "" {
			if configValueFloat, ok := configValue.(float64); ok {
				utils.Debug("Using config value for numFilesToCommit: " + strconv.FormatFloat(configValueFloat, 'f', -1, 64))
				numFilesToCommit = int(configValueFloat)
			} else if configValueStr, ok := configValue.(string); ok {
				if parsedValue, err := strconv.Atoi(configValueStr); err == nil {
					utils.Debug("Using config value for numFilesToCommit from string: " + configValueStr)
					numFilesToCommit = parsedValue
				} else {
					utils.Error("Invalid string value for numFilesToCommit: " + configValueStr)
				}
			}
		}
	}

	utils.Debug("Number of files to prepare commit messages for: " + strconv.Itoa(numFilesToCommit))

	utils.Debug("Root folder to get messages : " + folder)

	changedFiles, err := git.GetAllChangedFiles(folder)
	if err != nil {
		utils.Error("Failed to get changed files: " + err.Error())
		return fmt.Errorf("failed to get changed files: %s", err.Error())
	}

	if len(changedFiles) == 0 {
		utils.Info("No changed files found")
		return nil
	}

	if len(changedFiles) > numFiles[0] {
		changedFiles = changedFiles[:numFilesToCommit]
	}

	utils.Debug("Total files to process: " + strconv.Itoa(len(changedFiles)))

	err = git.BatchProcessGetMessages(changedFiles, folder)
	if err != nil {
		utils.Error("Batch processing failed: " + err.Error())
		return fmt.Errorf("batch processing failed: %s", err.Error())
	}

	utils.Info(fmt.Sprintf("Batch generation of all messages for %s completed successfully", folder))
	utils.Debug(fmt.Sprintf("All output : %v", output.GetAll()))
	output.SaveToFile()
	return nil
}
