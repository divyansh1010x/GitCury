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

	// Start creative loader for message generation
	utils.StartCreativeLoader("Analyzing repository changes", utils.GitAnimation)
	utils.UpdateCreativeLoaderPhase("analyzing")

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		// Capture clustering configuration for stats
		clusteringConfig := config.GetClusteringConfig()

		// Determine enabled methods
		enabledMethods := []string{}
		if config.IsMethodEnabled(config.DirectoryMethod) {
			enabledMethods = append(enabledMethods, "directory")
		}
		if config.IsMethodEnabled(config.PatternMethod) {
			enabledMethods = append(enabledMethods, "pattern")
		}
		if config.IsMethodEnabled(config.CachedMethod) {
			enabledMethods = append(enabledMethods, "cached")
		}
		if config.IsMethodEnabled(config.SemanticMethod) {
			enabledMethods = append(enabledMethods, "semantic")
		}

		// Determine performance mode based on config
		performanceMode := "balanced" // default
		if clusteringConfig.Performance.PreferSpeed {
			performanceMode = "speed"
		}
		if clusteringConfig.Performance.MaxProcessingTime > 90 {
			performanceMode = "quality"
		}

		utils.CaptureClusteringConfigFromSettings(
			clusteringConfig.DefaultMethod,
			enabledMethods,
			clusteringConfig.ConfidenceThresholds,
			clusteringConfig.SimilarityThresholds,
			clusteringConfig.MaxFilesForSemanticClustering,
			clusteringConfig.EnableFallbackMethods,
			performanceMode,
			clusteringConfig.Performance.MaxProcessingTime,
			clusteringConfig.Performance.EnableBenchmarking,
			clusteringConfig.Performance.AdaptiveOptimization,
		)

		utils.UpdateOperationProgress("GenerateAllMessages", 20.0)
	}

	utils.Debug("[" + config.Aliases.GetMsgs + "]: Preparing commit messages for " + strconv.Itoa(numFiles[0]) + " files per folder.")

	// Update loader phase
	utils.UpdateCreativeLoaderPhase("processing")
	utils.UpdateCreativeLoaderMessage("Processing root folders")

	rootFolders, ok := config.Get("root_folders").([]interface{})
	if !ok {
		utils.StopCreativeLoader()
		utils.Error("[" + config.Aliases.GetMsgs + "]: Invalid or missing root_folders configuration.")
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
			utils.Error("[" + config.Aliases.GetMsgs + "]: Invalid root folder type.")
			continue
		}

		rootFolderWg.Add(1)
		go func(folder string) {
			defer rootFolderWg.Done()

			utils.Debug("[" + config.Aliases.GetMsgs + "]: Processing root folder: " + folder)
			utils.UpdateCreativeLoaderMessage(fmt.Sprintf("Processing folder: %s", folder))

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
				mu.Lock()
				processedFolders++
				utils.UpdateCreativeLoaderMessage(fmt.Sprintf("Processed %d/%d folders", processedFolders, totalFolders))
				mu.Unlock()
				return
			}

			if len(changedFiles) > numFiles[0] {
				changedFiles = changedFiles[:numFiles[0]]
			}

			utils.Debug("[" + config.Aliases.GetMsgs + "]: Total files to process in folder '" + folder + "': " + strconv.Itoa(len(changedFiles)))
			utils.UpdateCreativeLoaderPhase("generating")
			utils.UpdateCreativeLoaderMessage(fmt.Sprintf("Generating messages for %d files in %s", len(changedFiles), folder))

			err = git.BatchProcessGetMessages(changedFiles, folder)
			if err != nil {
				utils.Error("[" + config.Aliases.GetMsgs + "]: Batch processing failed for folder '" + folder + "' - " + err.Error())
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
		utils.Error("[" + config.Aliases.GetMsgs + "]: Batch processing completed with errors.")
		utils.Debug("[" + config.Aliases.GetMsgs + "]: Errors encountered: " + fmt.Sprint(errors))
		return fmt.Errorf("one or more errors occurred while preparing commit messages")
	}

	// Stop loader and show success
	utils.StopCreativeLoader()
	utils.ShowCompletionMessage("Commit message generation completed successfully for all folders", true)
	utils.Success("[" + config.Aliases.GetMsgs + "]: Commit message generation completed successfully for all folders.")

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("GenerateAllMessages", 90.0)
	}

	output.SaveToFile()
	return nil
}

func GetMsgsForRootFolder(folder string, numFiles ...int) error {
	if folder == "" {
		utils.Error("[" + config.Aliases.GetMsgs + "]: Root folder is empty.")
		return fmt.Errorf("root folder is empty")
	}

	// Start creative loader for single folder processing
	utils.StartCreativeLoader(fmt.Sprintf("Analyzing folder: %s", folder), utils.ProcessingAnimation)
	utils.UpdateCreativeLoaderPhase("analyzing")

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("GenerateRootMessages", 20.0)
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
					utils.Error("[" + config.Aliases.GetMsgs + "]: Invalid string value for numFilesToCommit: " + configValueStr)
				}
			}
		}
	}

	utils.Debug("[" + config.Aliases.GetMsgs + "]: Preparing commit messages for " + strconv.Itoa(numFilesToCommit) + " files in folder: " + folder)

	utils.UpdateCreativeLoaderPhase("processing")
	utils.UpdateCreativeLoaderMessage("Scanning for changed files")

	changedFiles, err := git.GetAllChangedFiles(folder)
	if err != nil {
		utils.StopCreativeLoader()
		utils.Error("[" + config.Aliases.GetMsgs + "]: Failed to retrieve changed files for folder '" + folder + "' - " + err.Error())
		return fmt.Errorf("failed to get changed files: %s", err.Error())
	}

	if len(changedFiles) == 0 {
		utils.StopCreativeLoader()
		utils.ShowCompletionMessage(fmt.Sprintf("No changed files found in folder: %s", folder), true)
		utils.Debug("[" + config.Aliases.GetMsgs + "]: No changed files found in folder: " + folder)
		return nil
	}

	if len(changedFiles) > numFilesToCommit {
		changedFiles = changedFiles[:numFilesToCommit]
	}

	utils.Debug("[" + config.Aliases.GetMsgs + "]: Total files to process in folder '" + folder + "': " + strconv.Itoa(len(changedFiles)))

	utils.UpdateCreativeLoaderPhase("generating")
	utils.UpdateCreativeLoaderMessage(fmt.Sprintf("Generating messages for %d files", len(changedFiles)))

	err = git.BatchProcessGetMessages(changedFiles, folder)
	if err != nil {
		utils.StopCreativeLoader()
		utils.ShowCompletionMessage("Batch processing failed", false)
		utils.Error("[" + config.Aliases.GetMsgs + "]: Batch processing failed for folder '" + folder + "' - " + err.Error())
		return fmt.Errorf("batch processing failed: %s", err.Error())
	}

	// Stop loader and show success
	utils.StopCreativeLoader()
	utils.ShowCompletionMessage(fmt.Sprintf("Commit message generation completed for folder: %s", folder), true)
	utils.Success("[" + config.Aliases.GetMsgs + "]: Commit message generation completed successfully for folder: " + folder)
	utils.Debug("[" + config.Aliases.GetMsgs + "]: All output: " + fmt.Sprint(output.GetAll()))

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("GenerateRootMessages", 90.0)
	}

	output.SaveToFile()
	return nil
}

func GroupAndGetAllMsgs(numFiles ...int) error {
	// Start creative loader for grouped processing
	utils.StartCreativeLoader("Analyzing repository for grouped processing", utils.BrailleAnimation)
	utils.UpdateCreativeLoaderPhase("clustering")

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("GenerateAllMessages", 20.0)
	}

	utils.Debug("[" + config.Aliases.GetMsgs + "]: Preparing grouped commit messages with embeddings for " + strconv.Itoa(numFiles[0]) + " files per folder.")

	rootFolders, ok := config.Get("root_folders").([]interface{})
	if !ok {
		utils.StopCreativeLoader()
		utils.Error("[" + config.Aliases.GetMsgs + "]: Invalid or missing root_folders configuration.")
		return fmt.Errorf("invalid or missing root_folders configuration")
	}

	clusters := 10 // Default value
	if len(numFiles) > 0 && numFiles[0] > 0 {
		utils.Debug("[" + config.Aliases.GetMsgs + "]: Using provided number of files to commit: " + strconv.Itoa(numFiles[0]))
		clusters = numFiles[0]
	} else {
		if configValue := config.Get("numFilesToCommit"); configValue != "" {
			if configValueFloat, ok := configValue.(float64); ok {
				utils.Debug("[" + config.Aliases.GetMsgs + "]: Using config value for numFilesToCommit: " + strconv.FormatFloat(configValueFloat, 'f', -1, 64))
				clusters = int(configValueFloat)
			} else if configValueStr, ok := configValue.(string); ok {
				if parsedValue, err := strconv.Atoi(configValueStr); err == nil {
					utils.Debug("[" + config.Aliases.GetMsgs + "]: Using config value for numFilesToCommit from string: " + configValueStr)
					clusters = parsedValue
				} else {
					utils.Error("[" + config.Aliases.GetMsgs + "]: Invalid string value for numFilesToCommit: " + configValueStr)
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
			utils.Error("[" + config.Aliases.GetMsgs + "]: Invalid root folder type.")
			continue
		}

		rootFolderWg.Add(1)
		go func(folder string) {
			defer rootFolderWg.Done()

			utils.Debug("[" + config.Aliases.GetMsgs + "]: Grouped (embedding-based) processing for folder: " + folder)
			utils.UpdateCreativeLoaderMessage(fmt.Sprintf("Clustering files in folder: %s", folder))

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
			utils.UpdateCreativeLoaderPhase("generating")
			utils.UpdateCreativeLoaderMessage(fmt.Sprintf("Generating grouped messages for %d files", len(changedFiles)))

			err = git.BatchProcessWithEmbeddings(changedFiles, folder, clusters)
			if err != nil {
				utils.Error("[" + config.Aliases.GetMsgs + "]: Embedding-based batch processing failed for folder '" + folder + "' - " + err.Error())
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
		utils.Error("[" + config.Aliases.GetMsgs + "]: Grouped embedding-based batch processing completed with errors.")
		utils.Debug("[" + config.Aliases.GetMsgs + "]: Errors encountered: " + fmt.Sprint(errors))
		return fmt.Errorf("one or more errors occurred during grouped commit message generation with embeddings")
	}

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("GenerateAllMessages", 90.0)
	}

	// Stop loader and show success
	utils.StopCreativeLoader()
	utils.ShowCompletionMessage("Grouped commit message generation completed successfully for all folders", true)
	utils.Success("[" + config.Aliases.GetMsgs + "]: Grouped commit message generation with embeddings completed successfully for all folders.")
	output.SaveToFile()
	return nil
}

func GroupAndGetMsgsForRootFolder(folder string, numFiles ...int) error {
	if folder == "" {
		utils.Error("[" + config.Aliases.GetMsgs + "]: Root folder is empty.")
		return fmt.Errorf("root folder is empty")
	}

	// Start creative loader for grouped single folder processing
	utils.StartCreativeLoader(fmt.Sprintf("Clustering files in folder: %s", folder), utils.BrailleAnimation)
	utils.UpdateCreativeLoaderPhase("clustering")

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("GenerateRootMessages", 20.0)
	}

	clusters := 10 // Default value
	if len(numFiles) > 0 && numFiles[0] > 0 {
		utils.Debug("[" + config.Aliases.GetMsgs + "]: Using provided number of files to commit: " + strconv.Itoa(numFiles[0]))
		clusters = numFiles[0]
	} else {
		if configValue := config.Get("numFilesToCommit"); configValue != "" {
			if configValueFloat, ok := configValue.(float64); ok {
				utils.Debug("[" + config.Aliases.GetMsgs + "]: Using config value for numFilesToCommit: " + strconv.FormatFloat(configValueFloat, 'f', -1, 64))
				clusters = int(configValueFloat)
			} else if configValueStr, ok := configValue.(string); ok {
				if parsedValue, err := strconv.Atoi(configValueStr); err == nil {
					utils.Debug("[" + config.Aliases.GetMsgs + "]: Using config value for numFilesToCommit from string: " + configValueStr)
					clusters = parsedValue
				} else {
					utils.Error("[" + config.Aliases.GetMsgs + "]: Invalid string value for numFilesToCommit: " + configValueStr)
				}
			}
		}
	}

	utils.Debug("[" + config.Aliases.GetMsgs + "]: Preparing commit messages for " + strconv.Itoa(clusters) + " files in folder: " + folder)

	utils.UpdateCreativeLoaderPhase("processing")
	utils.UpdateCreativeLoaderMessage("Scanning for changed files")

	changedFiles, err := git.GetAllChangedFiles(folder)
	if err != nil {
		utils.StopCreativeLoader()
		utils.Error("[" + config.Aliases.GetMsgs + "]: Failed to retrieve changed files for folder '" + folder + "' - " + err.Error())
		return fmt.Errorf("failed to get changed files: %s", err.Error())
	}

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("GenerateRootMessages", 40.0)
	}

	if len(changedFiles) == 0 {
		utils.StopCreativeLoader()
		utils.ShowCompletionMessage(fmt.Sprintf("No changed files found in folder: %s", folder), true)
		utils.Debug("[" + config.Aliases.GetMsgs + "]: No changed files found in folder: " + folder)
		return nil
	}

	utils.Debug("[" + config.Aliases.GetMsgs + "]: Total files to process in folder '" + folder + "': " + strconv.Itoa(len(changedFiles)))
	utils.UpdateCreativeLoaderPhase("generating")
	utils.UpdateCreativeLoaderMessage(fmt.Sprintf("Generating grouped messages for %d files", len(changedFiles)))

	err = git.BatchProcessWithEmbeddings(changedFiles, folder, clusters)
	if err != nil {
		utils.StopCreativeLoader()
		utils.ShowCompletionMessage("Grouped batch processing failed", false)
		utils.Error("[" + config.Aliases.GetMsgs + "]: Batch processing failed for folder '" + folder + "' - " + err.Error())
		return fmt.Errorf("batch processing failed: %s", err.Error())
	}

	// Stop loader and show success
	utils.StopCreativeLoader()
	utils.ShowCompletionMessage(fmt.Sprintf("Grouped commit message generation completed for folder: %s", folder), true)
	utils.Success("[" + config.Aliases.GetMsgs + "]: Commit message generation completed successfully for folder: " + folder)
	utils.Debug("[" + config.Aliases.GetMsgs + "]: All output: " + fmt.Sprint(output.GetAll()))

	// Update progress in stats if enabled
	if utils.IsStatsEnabled() {
		utils.UpdateOperationProgress("GenerateRootMessages", 90.0)
	}

	output.SaveToFile()
	return nil
}
