package core

import (
	"github.com/lakshyajain-0291/gitcury/config"
	"github.com/lakshyajain-0291/gitcury/output"
	"github.com/lakshyajain-0291/gitcury/utils"
	"fmt"
	"strconv"
	"sync"
)

func GetAllMsgs(numFiles ...int) error {
	defaultNumFiles := 10 // Default value
	if len(numFiles) == 0 || numFiles[0] <= 0 {
		numFiles[0] = defaultNumFiles
	}

	// Start creative loader for message generation
	utils.StartCreativeLoader("Analyzing repository changes", utils.GitAnimation)
	utils.UpdateCreativeLoaderPhase("analyzing")

	utils.Debug("Preparing commit messages for " + strconv.Itoa(numFiles[0]) + " files per folder.")

	// Update loader phase
	utils.UpdateCreativeLoaderPhase("processing")
	utils.UpdateCreativeLoaderMessage("Processing root folders")

	rootFolders, ok := config.Get("root_folders").([]interface{})
	if !ok {
		utils.StopCreativeLoader()
		utils.Error("Invalid or missing root_folders configuration.")
		return fmt.Errorf("invalid or missing root_folders configuration")
	}

	var rootFolderWg sync.WaitGroup
	var mu sync.Mutex
	var errors []string
	totalFolders := len(rootFolders)
	processedFolders := 0

	for _, rootFolder := range rootFolders {
		rootFolderStr, ok := rootFolder.(string)
		if !ok {
			utils.Error("Invalid root folder type.")
			continue
		}

		rootFolderWg.Add(1)
		go func(folder string) {
			defer rootFolderWg.Done()

			utils.Debug("Processing root folder: " + folder)
			utils.UpdateCreativeLoaderMessage(fmt.Sprintf("Processing folder: %s", folder))

			changedFiles, err := GitRunnerInstance.GetChangedFiles([]string{folder}, 5)
			if err != nil {
				utils.Error("Failed to retrieve changed files for folder '" + folder + "' - " + err.Error())
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, Error: %s", folder, err.Error()))
				mu.Unlock()
				return
			}

			// Extract file paths from folders
			var allChangedFiles []string
			for _, folderData := range changedFiles {
				for _, fileEntry := range folderData.Files {
					allChangedFiles = append(allChangedFiles, fileEntry.Name)
				}
			}

			if len(allChangedFiles) == 0 {
				utils.Debug("No changed files found in folder: " + folder)
				mu.Lock()
				processedFolders++
				utils.UpdateCreativeLoaderMessage(fmt.Sprintf("Processed %d/%d folders", processedFolders, totalFolders))
				mu.Unlock()
				return
			}

			if len(allChangedFiles) > numFiles[0] {
				allChangedFiles = allChangedFiles[:numFiles[0]]
			}

			utils.Debug("Total files to process in folder '" + folder + "': " + strconv.Itoa(len(allChangedFiles)))
			utils.UpdateCreativeLoaderPhase("generating")
			utils.UpdateCreativeLoaderMessage(fmt.Sprintf("Generating messages for %d files in %s", len(allChangedFiles), folder))

			err = GitRunnerInstance.BatchProcessGetMessages(allChangedFiles, folder)
			if err != nil {
				utils.Error("Batch processing failed for folder '" + folder + "' - " + err.Error())
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, Error: %s", folder, err.Error()))
				mu.Unlock()
			}

			mu.Lock()
			processedFolders++
			utils.UpdateCreativeLoaderMessage(fmt.Sprintf("Processed %d/%d folders", processedFolders, totalFolders))
			mu.Unlock()
		}(rootFolderStr)
	}

	// Update loader to show waiting for completion
	utils.UpdateCreativeLoaderPhase("finalizing")
	utils.UpdateCreativeLoaderMessage("Waiting for all folders to complete")

	rootFolderWg.Wait()

	if len(errors) > 0 {
		utils.StopCreativeLoader()
		utils.ShowCompletionMessage("Batch processing completed with errors", false)
		utils.Error("Batch processing completed with errors.")
		utils.Debug("Errors encountered: " + fmt.Sprint(errors))
		return fmt.Errorf("one or more errors occurred while preparing commit messages")
	}

	// Stop loader and show success
	utils.StopCreativeLoader()
	utils.ShowCompletionMessage("Commit message generation completed successfully for all folders", true)
	utils.Success("Commit message generation completed successfully for all folders.")

	output.SaveToFile()
	return nil
}

func GetMsgsForRootFolder(folder string, numFiles ...int) error {
	if folder == "" {
		utils.Error("Root folder is empty.")
		return fmt.Errorf("root folder is empty")
	}

	// Start creative loader for single folder processing
	utils.StartCreativeLoader(fmt.Sprintf("Analyzing folder: %s", folder), utils.ProcessingAnimation)
	utils.UpdateCreativeLoaderPhase("analyzing")

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

	utils.Debug("Preparing commit messages for " + strconv.Itoa(numFilesToCommit) + " files in folder: " + folder)

	utils.UpdateCreativeLoaderPhase("processing")
	utils.UpdateCreativeLoaderMessage("Scanning for changed files")

	changedFiles, err := GitRunnerInstance.GetAllChangedFiles(folder)
	if err != nil {
		utils.StopCreativeLoader()
		utils.Error("Failed to retrieve changed files for folder '" + folder + "' - " + err.Error())
		return fmt.Errorf("failed to get changed files: %s", err.Error())
	}

	if len(changedFiles) == 0 {
		utils.StopCreativeLoader()
		utils.ShowCompletionMessage(fmt.Sprintf("No changed files found in folder: %s", folder), true)
		utils.Debug("No changed files found in folder: " + folder)
		return nil
	}

	if len(changedFiles) > numFilesToCommit {
		changedFiles = changedFiles[:numFilesToCommit]
	}

	utils.Debug("Total files to process in folder '" + folder + "': " + strconv.Itoa(len(changedFiles)))

	utils.UpdateCreativeLoaderPhase("generating")
	utils.UpdateCreativeLoaderMessage(fmt.Sprintf("Generating messages for %d files", len(changedFiles)))

	err = GitRunnerInstance.BatchProcessGetMessages(changedFiles, folder)
	if err != nil {
		utils.StopCreativeLoader()
		utils.ShowCompletionMessage("Batch processing failed", false)
		utils.Error("Batch processing failed for folder '" + folder + "' - " + err.Error())
		return fmt.Errorf("batch processing failed: %s", err.Error())
	}

	// Stop loader and show success
	utils.StopCreativeLoader()
	utils.ShowCompletionMessage(fmt.Sprintf("Commit message generation completed for folder: %s", folder), true)
	utils.Success("Commit message generation completed successfully for folder: " + folder)
	utils.Debug("All output: " + fmt.Sprint(output.GetAll()))

	output.SaveToFile()
	return nil
}

func GroupAndGetAllMsgs(numFiles ...int) error {
	// Start creative loader for grouped processing
	utils.StartCreativeLoader("Analyzing repository for grouped processing", utils.BrailleAnimation)
	utils.UpdateCreativeLoaderPhase("clustering")

	utils.Debug("Preparing grouped commit messages with embeddings for " + strconv.Itoa(numFiles[0]) + " files per folder.")

	rootFolders, ok := config.Get("root_folders").([]interface{})
	if !ok {
		utils.StopCreativeLoader()
		utils.Error("Invalid or missing root_folders configuration.")
		return fmt.Errorf("invalid or missing root_folders configuration")
	}

	clusters := 10 // Default value
	if len(numFiles) > 0 && numFiles[0] > 0 {
		utils.Debug("Using provided number of files to commit: " + strconv.Itoa(numFiles[0]))
		clusters = numFiles[0]
	} else {
		if configValue := config.Get("numFilesToCommit"); configValue != "" {
			if configValueFloat, ok := configValue.(float64); ok {
				utils.Debug("Using config value for numFilesToCommit: " + strconv.FormatFloat(configValueFloat, 'f', -1, 64))
				clusters = int(configValueFloat)
			} else if configValueStr, ok := configValue.(string); ok {
				if parsedValue, err := strconv.Atoi(configValueStr); err == nil {
					utils.Debug("Using config value for numFilesToCommit from string: " + configValueStr)
					clusters = parsedValue
				} else {
					utils.Error("Invalid string value for numFilesToCommit: " + configValueStr)
				}
			}
		}
	}

	var rootFolderWg sync.WaitGroup
	var mu sync.Mutex
	var errors []string

	for _, rootFolder := range rootFolders {
		rootFolderStr, ok := rootFolder.(string)
		if !ok {
			utils.Error("Invalid root folder type.")
			continue
		}

		rootFolderWg.Add(1)
		go func(folder string) {
			defer rootFolderWg.Done()

			utils.Debug("Grouped (embedding-based) processing for folder: " + folder)
			utils.UpdateCreativeLoaderMessage(fmt.Sprintf("Clustering files in folder: %s", folder))

			changedFiles, err := GitRunnerInstance.GetAllChangedFiles(folder)
			if err != nil {
				utils.Error("Failed to retrieve changed files for folder '" + folder + "' - " + err.Error())
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, Error: %s", folder, err.Error()))
				mu.Unlock()
				return
			}

			if len(changedFiles) == 0 {
				utils.Debug("No changed files found in folder: " + folder)
				return
			}

			utils.Debug("Total files to process with embeddings in folder '" + folder + "': " + strconv.Itoa(len(changedFiles)))
			utils.UpdateCreativeLoaderPhase("generating")
			utils.UpdateCreativeLoaderMessage(fmt.Sprintf("Generating grouped messages for %d files", len(changedFiles)))

			err = GitRunnerInstance.BatchProcessWithEmbeddings(changedFiles, folder, clusters)
			if err != nil {
				utils.Error("Embedding-based batch processing failed for folder '" + folder + "' - " + err.Error())
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, Error: %s", folder, err.Error()))
				mu.Unlock()
			}
		}(rootFolderStr)
	}

	utils.UpdateCreativeLoaderPhase("finalizing")
	utils.UpdateCreativeLoaderMessage("Completing grouped processing")

	rootFolderWg.Wait()

	if len(errors) > 0 {
		utils.StopCreativeLoader()
		utils.ShowCompletionMessage("Grouped processing completed with errors", false)
		utils.Error("Grouped embedding-based batch processing completed with errors.")
		utils.Debug("Errors encountered: " + fmt.Sprint(errors))
		return fmt.Errorf("one or more errors occurred during grouped commit message generation with embeddings")
	}

	// Stop loader and show success
	utils.StopCreativeLoader()
	utils.ShowCompletionMessage("Grouped commit message generation completed successfully for all folders", true)
	utils.Success("Grouped commit message generation with embeddings completed successfully for all folders.")
	output.SaveToFile()
	return nil
}

func GroupAndGetMsgsForRootFolder(folder string, numFiles ...int) error {
	if folder == "" {
		utils.Error("Root folder is empty.")
		return fmt.Errorf("root folder is empty")
	}

	// Start creative loader for grouped single folder processing
	utils.StartCreativeLoader(fmt.Sprintf("Clustering files in folder: %s", folder), utils.BrailleAnimation)
	utils.UpdateCreativeLoaderPhase("clustering")

	clusters := 10 // Default value
	if len(numFiles) > 0 && numFiles[0] > 0 {
		utils.Debug("Using provided number of files to commit: " + strconv.Itoa(numFiles[0]))
		clusters = numFiles[0]
	} else {
		if configValue := config.Get("numFilesToCommit"); configValue != "" {
			if configValueFloat, ok := configValue.(float64); ok {
				utils.Debug("Using config value for numFilesToCommit: " + strconv.FormatFloat(configValueFloat, 'f', -1, 64))
				clusters = int(configValueFloat)
			} else if configValueStr, ok := configValue.(string); ok {
				if parsedValue, err := strconv.Atoi(configValueStr); err == nil {
					utils.Debug("Using config value for numFilesToCommit from string: " + configValueStr)
					clusters = parsedValue
				} else {
					utils.Error("Invalid string value for numFilesToCommit: " + configValueStr)
				}
			}
		}
	}

	utils.Debug("Preparing commit messages for " + strconv.Itoa(clusters) + " files in folder: " + folder)

	utils.UpdateCreativeLoaderPhase("processing")
	utils.UpdateCreativeLoaderMessage("Scanning for changed files")

	changedFiles, err := GitRunnerInstance.GetAllChangedFiles(folder)
	if err != nil {
		utils.StopCreativeLoader()
		utils.Error("Failed to retrieve changed files for folder '" + folder + "' - " + err.Error())
		return fmt.Errorf("failed to get changed files: %s", err.Error())
	}

	if len(changedFiles) == 0 {
		utils.StopCreativeLoader()
		utils.ShowCompletionMessage(fmt.Sprintf("No changed files found in folder: %s", folder), true)
		utils.Debug("No changed files found in folder: " + folder)
		return nil
	}

	utils.Debug("Total files to process in folder '" + folder + "': " + strconv.Itoa(len(changedFiles)))
	utils.UpdateCreativeLoaderPhase("generating")
	utils.UpdateCreativeLoaderMessage(fmt.Sprintf("Generating grouped messages for %d files", len(changedFiles)))

	err = GitRunnerInstance.BatchProcessWithEmbeddings(changedFiles, folder, clusters)
	if err != nil {
		utils.StopCreativeLoader()
		utils.ShowCompletionMessage("Grouped batch processing failed", false)
		utils.Error("Batch processing failed for folder '" + folder + "' - " + err.Error())
		return fmt.Errorf("batch processing failed: %s", err.Error())
	}

	// Stop loader and show success
	utils.StopCreativeLoader()
	utils.ShowCompletionMessage(fmt.Sprintf("Grouped commit message generation completed for folder: %s", folder), true)
	utils.Success("Commit message generation completed successfully for folder: " + folder)
	utils.Debug("All output: " + fmt.Sprint(output.GetAll()))

	output.SaveToFile()
	return nil
}
