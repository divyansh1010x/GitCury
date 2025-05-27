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
	if len(numFiles) == 0 || numFiles[0] <= 0 {
		numFiles[0] = defaultNumFiles
	}

	utils.Debug("[" + config.Aliases.GetMsgs + "]: Preparing commit messages for " + strconv.Itoa(numFiles[0]) + " files per folder.")

	rootFolders, ok := config.Get("root_folders").([]interface{})
	if !ok {
		utils.Error("[" + config.Aliases.GetMsgs + "]: Invalid or missing root_folders configuration.", "config")
		return utils.NewValidationError(
			"Invalid or missing root_folders configuration",
			nil,
			map[string]interface{}{
				"configKey": "root_folders",
			},
			"config",
		)
	}

	var rootFolderWg sync.WaitGroup
	var mu sync.Mutex
	var errors []string

	for _, rootFolder := range rootFolders {
		rootFolderStr, ok := rootFolder.(string)
		if !ok {
			utils.Error("[" + config.Aliases.GetMsgs + "]: Invalid root folder type.")
			continue
		}

		rootFolderWg.Add(1)
		go func(folder string) {
			defer rootFolderWg.Done()

			utils.Debug("[" + config.Aliases.GetMsgs + "]: Processing root folder: " + folder)

			changedFiles, err := git.GetAllChangedFiles(folder)
			if err != nil {
				// Extract file information if available in the error
				fileInfo := folder
				if structErr, ok := err.(*utils.StructuredError); ok && structErr.ProcessedFile != "" {
					fileInfo = structErr.ProcessedFile
				}
				
				utils.Error("[" + config.Aliases.GetMsgs + "]: Failed to retrieve changed files for folder '" + folder + "' - " + err.Error(), fileInfo)
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, File: %s, Error: %s", folder, fileInfo, err.Error()))
				mu.Unlock()
				return
			}

			if len(changedFiles) == 0 {
				utils.Debug("[" + config.Aliases.GetMsgs + "]: No changed files found in folder: " + folder)
				return
			}

			if len(changedFiles) > numFiles[0] {
				changedFiles = changedFiles[:numFiles[0]]
			}

			utils.Debug("[" + config.Aliases.GetMsgs + "]: Total files to process in folder '" + folder + "': " + strconv.Itoa(len(changedFiles)))

			err = git.BatchProcessGetMessages(changedFiles, folder)
			if err != nil {
				// Extract file information if available in the error
				fileInfo := folder
				if structErr, ok := err.(*utils.StructuredError); ok && structErr.ProcessedFile != "" {
					fileInfo = structErr.ProcessedFile
				}
				
				utils.Error("[" + config.Aliases.GetMsgs + "]: Batch processing failed for folder '" + folder + "' - " + err.Error(), fileInfo)
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, File: %s, Error: %s", folder, fileInfo, err.Error()))
				mu.Unlock()
			}
		}(rootFolderStr)
	}

	rootFolderWg.Wait()

	if len(errors) > 0 {
		utils.Error("[" + config.Aliases.GetMsgs + "]: Batch processing completed with errors.")
		utils.Debug("[" + config.Aliases.GetMsgs + "]: Errors encountered: " + fmt.Sprint(errors))
		return fmt.Errorf("one or more errors occurred while preparing commit messages")
	}

	utils.Success("[" + config.Aliases.GetMsgs + "]: Commit message generation completed successfully for all folders.")
	output.SaveToFile()
	return nil
}

func GetMsgsForRootFolder(folder string, numFiles ...int) error {
	if folder == "" {
		utils.Error("[" + config.Aliases.GetMsgs + "]: Root folder is empty.", "")
		return utils.NewValidationError(
			"Root folder is empty",
			nil,
			map[string]interface{}{
				"folder": folder,
			},
			"",
		)
	}

	numFilesToCommit := 10 // Default value
	if len(numFiles) > 0 && numFiles[0] > 0 {
		utils.Debug("[" + config.Aliases.GetMsgs + "]: Using provided number of files to commit: " + strconv.Itoa(numFiles[0]))
		numFilesToCommit = numFiles[0]
	} else {
		if configValue := config.Get("numFilesToCommit"); configValue != "" {
			if configValueFloat, ok := configValue.(float64); ok {
				utils.Debug("[" + config.Aliases.GetMsgs + "]: Using config value for numFilesToCommit: " + strconv.FormatFloat(configValueFloat, 'f', -1, 64))
				numFilesToCommit = int(configValueFloat)
			} else if configValueStr, ok := configValue.(string); ok {
				if parsedValue, err := strconv.Atoi(configValueStr); err == nil {
					utils.Debug("[" + config.Aliases.GetMsgs + "]: Using config value for numFilesToCommit from string: " + configValueStr)
					numFilesToCommit = parsedValue
				} else {
					utils.Error("[" + config.Aliases.GetMsgs + "]: Invalid string value for numFilesToCommit: " + configValueStr, "config")
				}
			}
		}
	}

	utils.Debug("[" + config.Aliases.GetMsgs + "]: Preparing commit messages for " + strconv.Itoa(numFilesToCommit) + " files in folder: " + folder)

	changedFiles, err := git.GetAllChangedFiles(folder)
	if err != nil {
		// Extract file information if available in the error
		fileInfo := folder
		if structErr, ok := err.(*utils.StructuredError); ok && structErr.ProcessedFile != "" {
			fileInfo = structErr.ProcessedFile
		}
		
		utils.Error("[" + config.Aliases.GetMsgs + "]: Failed to retrieve changed files for folder '" + folder + "' - " + err.Error(), fileInfo)
		return utils.NewGitError(
			"Failed to get changed files",
			err,
			map[string]interface{}{
				"folder": folder,
			},
			fileInfo,
		)
	}

	if len(changedFiles) == 0 {
		utils.Debug("[" + config.Aliases.GetMsgs + "]: No changed files found in folder: " + folder)
		return nil
	}

	if len(changedFiles) > numFilesToCommit {
		changedFiles = changedFiles[:numFilesToCommit]
	}

	utils.Debug("[" + config.Aliases.GetMsgs + "]: Total files to process in folder '" + folder + "': " + strconv.Itoa(len(changedFiles)))

	err = git.BatchProcessGetMessages(changedFiles, folder)
	if err != nil {
		// Extract file information if available in the error
		fileInfo := folder
		if structErr, ok := err.(*utils.StructuredError); ok && structErr.ProcessedFile != "" {
			fileInfo = structErr.ProcessedFile
		}
		
		utils.Error("[" + config.Aliases.GetMsgs + "]: Batch processing failed for folder '" + folder + "' - " + err.Error(), fileInfo)
		return utils.NewGitError(
			"Batch processing failed",
			err,
			map[string]interface{}{
				"folder": folder,
			},
			fileInfo,
		)
	}

	utils.Success("[" + config.Aliases.GetMsgs + "]: Commit message generation completed successfully for folder: " + folder)
	utils.Debug("[" + config.Aliases.GetMsgs + "]: All output: " + fmt.Sprint(output.GetAll()))
	output.SaveToFile()
	return nil
}

func GroupAndGetAllMsgs(numFiles ...int) error {
	clusters := 0 // Default to threshold-based clustering when no explicit cluster count is provided
	isExplicitClusterCount := false
	
	if len(numFiles) > 0 && numFiles[0] > 0 {
		utils.Debug("[" + config.Aliases.GetMsgs + "]: Using provided number of files to commit: " + strconv.Itoa(numFiles[0]))
		clusters = numFiles[0]
		isExplicitClusterCount = true
	} else {
		if configValue := config.Get("numFilesToCommit"); configValue != "" {
			if configValueFloat, ok := configValue.(float64); ok {
				utils.Debug("[" + config.Aliases.GetMsgs + "]: Using config value for numFilesToCommit: " + strconv.FormatFloat(configValueFloat, 'f', -1, 64))
				clusters = int(configValueFloat)
				isExplicitClusterCount = true
			} else if configValueStr, ok := configValue.(string); ok {
				if parsedValue, err := strconv.Atoi(configValueStr); err == nil {
					utils.Debug("[" + config.Aliases.GetMsgs + "]: Using config value for numFilesToCommit from string: " + configValueStr)
					clusters = parsedValue
					isExplicitClusterCount = true
				} else {
					utils.Error("[" + config.Aliases.GetMsgs + "]: Invalid string value for numFilesToCommit: " + configValueStr)
				}
			}
		}
	}

	if isExplicitClusterCount {
		utils.Debug("[" + config.Aliases.GetMsgs + "]: Preparing grouped commit messages with embeddings for " + strconv.Itoa(clusters) + " clusters per folder.")
	} else {
		utils.Debug("[" + config.Aliases.GetMsgs + "]: Preparing grouped commit messages with threshold-based clustering per folder.")
	}

	rootFolders, ok := config.Get("root_folders").([]interface{})
	if !ok {
		utils.Error("[" + config.Aliases.GetMsgs + "]: Invalid or missing root_folders configuration.")
		return fmt.Errorf("invalid or missing root_folders configuration")
	}

	var rootFolderWg sync.WaitGroup
	var mu sync.Mutex
	var errors []string

	for _, rootFolder := range rootFolders {
		rootFolderStr, ok := rootFolder.(string)
		if !ok {
			utils.Error("[" + config.Aliases.GetMsgs + "]: Invalid root folder type.")
			continue
		}

		rootFolderWg.Add(1)
		go func(folder string) {
			defer rootFolderWg.Done()

			utils.Debug("[" + config.Aliases.GetMsgs + "]: Grouped (embedding-based) processing for folder: " + folder)

			changedFiles, err := git.GetAllChangedFiles(folder)
			if err != nil {
				utils.Error("[" + config.Aliases.GetMsgs + "]: Failed to retrieve changed files for folder '" + folder + "' - " + err.Error())
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, Error: %s", folder, err.Error()))
				mu.Unlock()
				return
			}

			if len(changedFiles) == 0 {
				utils.Debug("[" + config.Aliases.GetMsgs + "]: No changed files found in folder: " + folder)
				return
			}

			utils.Debug("[" + config.Aliases.GetMsgs + "]: Total files to process with embeddings in folder '" + folder + "': " + strconv.Itoa(len(changedFiles)))

			err = git.BatchProcessWithEmbeddings(changedFiles, folder, clusters)
			if err != nil {
				utils.Error("[" + config.Aliases.GetMsgs + "]: Embedding-based batch processing failed for folder '" + folder + "' - " + err.Error())
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, Error: %s", folder, err.Error()))
				mu.Unlock()
			}
		}(rootFolderStr)
	}

	rootFolderWg.Wait()

	if len(errors) > 0 {
		utils.Error("[" + config.Aliases.GetMsgs + "]: Grouped embedding-based batch processing completed with errors.")
		utils.Debug("[" + config.Aliases.GetMsgs + "]: Errors encountered: " + fmt.Sprint(errors))
		return fmt.Errorf("one or more errors occurred during grouped commit message generation with embeddings")
	}

	utils.Success("[" + config.Aliases.GetMsgs + "]: Grouped commit message generation with embeddings completed successfully for all folders.")
	output.SaveToFile()
	return nil
}

func GroupAndGetMsgsForRootFolder(folder string, numFiles ...int) error {
	if folder == "" {
		utils.Error("[" + config.Aliases.GetMsgs + "]: Root folder is empty.")
		return fmt.Errorf("root folder is empty")
	}

	clusters := 0 // Default to threshold-based clustering when no explicit cluster count is provided
	isExplicitClusterCount := false
	
	if len(numFiles) > 0 && numFiles[0] > 0 {
		utils.Debug("[" + config.Aliases.GetMsgs + "]: Using provided number of files to commit: " + strconv.Itoa(numFiles[0]))
		clusters = numFiles[0]
		isExplicitClusterCount = true
	} else {
		// Only use config value if no explicit numFiles was provided
		if configValue := config.Get("numFilesToCommit"); configValue != "" {
			if configValueFloat, ok := configValue.(float64); ok {
				utils.Debug("[" + config.Aliases.GetMsgs + "]: Using config value for numFilesToCommit: " + strconv.FormatFloat(configValueFloat, 'f', -1, 64))
				clusters = int(configValueFloat)
				isExplicitClusterCount = true
			} else if configValueStr, ok := configValue.(string); ok {
				if parsedValue, err := strconv.Atoi(configValueStr); err == nil {
					utils.Debug("[" + config.Aliases.GetMsgs + "]: Using config value for numFilesToCommit from string: " + configValueStr)
					clusters = parsedValue
					isExplicitClusterCount = true
				} else {
					utils.Error("[" + config.Aliases.GetMsgs + "]: Invalid string value for numFilesToCommit: " + configValueStr)
				}
			}
		}
	}

	if isExplicitClusterCount {
		utils.Debug("[" + config.Aliases.GetMsgs + "]: Preparing commit messages for " + strconv.Itoa(clusters) + " clusters in folder: " + folder)
	} else {
		utils.Debug("[" + config.Aliases.GetMsgs + "]: Preparing commit messages with threshold-based clustering in folder: " + folder)
	}

	changedFiles, err := git.GetAllChangedFiles(folder)
	if err != nil {
		utils.Error("[" + config.Aliases.GetMsgs + "]: Failed to retrieve changed files for folder '" + folder + "' - " + err.Error())
		return fmt.Errorf("failed to get changed files: %s", err.Error())
	}

	if len(changedFiles) == 0 {
		utils.Debug("[" + config.Aliases.GetMsgs + "]: No changed files found in folder: " + folder)
		return nil
	}

	utils.Debug("[" + config.Aliases.GetMsgs + "]: Total files to process in folder '" + folder + "': " + strconv.Itoa(len(changedFiles)))

	err = git.BatchProcessWithEmbeddings(changedFiles, folder, clusters)
	if err != nil {
		utils.Error("[" + config.Aliases.GetMsgs + "]: Batch processing failed for folder '" + folder + "' - " + err.Error())
		return fmt.Errorf("batch processing failed: %s", err.Error())
	}

	utils.Success("[" + config.Aliases.GetMsgs + "]: Commit message generation completed successfully for folder: " + folder)
	utils.Debug("[" + config.Aliases.GetMsgs + "]: All output: " + fmt.Sprint(output.GetAll()))
	output.SaveToFile()
	return nil
}
