package git

import (
	"GitCury/config"
	"GitCury/di"
	"GitCury/embeddings"
	"GitCury/output"
	"GitCury/utils"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type FileEmbedding struct {
	Pa	// Handle binary files first
	if len(binaryFiles) > 0 {
		// Stop the loader to prevent interference with user prompt
		utils.StopCreativeLoader()
		
		utils.Info(fmt.Sprintf("🔍 Detected %d binary file(s):", len(binaryFiles)))
		for _, file := range binaryFiles {
			fileType := utils.GetBinaryFileType(file)
			utils.Info(fmt.Sprintf("  • %s (%s)", filepath.Base(file), fileType))
		}
		fmt.Println()
		
		// Create detailed prompt showing binary files
		promptMessage := fmt.Sprintf("Generate automated commit messages for these %d binary file(s)?", len(binaryFiles))
		generateBinaryMessages := utils.ConfirmAction(promptMessage, true)Diff      string
	Embedding []float32
}

func RunGitCmd(dir string, envVars map[string]string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	// Append custom environment variables to the existing environment
	if envVars != nil {
		env := cmd.Env
		for key, value := range envVars {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		utils.Error(fmt.Sprintf(
			"[GIT.EXEC.FAIL]: Command failed: %s\nStdout: %s\nStderr: %s\n",
			err,
			stdout.String(),
			stderr.String(),
		))
		return "", err
	}

	utils.Debug("[GIT.EXEC.SUCCESS]: Command executed successfully in directory '" + dir + "': git " + strings.Join(args, " "))
	return stdout.String(), nil
}

var changedFilesCache = make(map[string]string)
var cacheMu sync.RWMutex

func GetAllChangedFiles(dir string) ([]string, error) {
	output, err := RunGitCmd(dir, nil, "status", "--porcelain")
	if err != nil {
		utils.Error("[GIT.STATUS.FAIL]: Failed to get git status: " + err.Error())
		return nil, err
	}

	if strings.TrimSpace(output) == "" {
		utils.Debug("[GIT.STATUS]: No changed files detected in directory: " + dir)
		return nil, nil
	}

	var changedFiles []string
	lines := strings.Split(output, "\n")

	cacheMu.Lock()
	defer cacheMu.Unlock()

	for _, line := range lines {
		if len(line) < 4 {
			continue
		}

		status := strings.TrimSpace(line[:2])
		relativePath := strings.TrimSpace(line[3:])
		absolutePath := filepath.Join(dir, relativePath)
		abs, err := filepath.Abs(absolutePath)
		if err != nil {
			utils.Error("[GIT.PATH.FAIL]: Failed to resolve absolute path for '" + relativePath + "': " + err.Error())
			continue
		}

		changedFilesCache[abs] = status

		if strings.HasPrefix(status, "D") {
			utils.Debug("[GIT.FILE.DELETED]: File marked as deleted: " + abs)
			changedFiles = append(changedFiles, abs)
			continue
		}

		info, err := os.Stat(abs)
		if err != nil {
			if os.IsNotExist(err) {
				utils.Debug("[GIT.FILE.MISSING]: File does not exist (possibly deleted): " + abs)
				changedFiles = append(changedFiles, abs)
				continue
			}
			utils.Error("[GIT.STAT.FAIL]: Failed to stat path '" + abs + "': " + err.Error())
			return nil, err
		}

		if info.IsDir() && status == "??" {
			innerOutput, err := RunGitCmd(dir, nil, "ls-files", "--others", "--exclude-standard", relativePath)
			if err != nil {
				utils.Error("[GIT.UNTRACKED.FAIL]: Failed to list files in untracked dir '" + relativePath + "': " + err.Error())
				return nil, err
			}

			for _, inner := range strings.Split(innerOutput, "\n") {
				if strings.TrimSpace(inner) == "" {
					continue
				}
				fullPath := filepath.Join(dir, inner)
				absInner, err := filepath.Abs(fullPath)
				if err == nil {
					changedFiles = append(changedFiles, absInner)
					changedFilesCache[absInner] = "??"
				}
			}
		} else {
			changedFiles = append(changedFiles, abs)
		}
	}

	utils.Debug("[GIT.CHANGED.FILES]: " + strings.Join(changedFiles, ", "))
	return changedFiles, nil
}

func GenCommitMessage(files []string, dir string) (string, error) {
	contextData := make(map[string]map[string]string)

	apiKey := config.Get("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return "", fmt.Errorf("Gemini API key not found in config or env")
		}
	}

	for _, file := range files {
		var fileType, diffOutput string

		cacheMu.RLock()
		status, cached := changedFilesCache[file]
		cacheMu.RUnlock()

		if cached && strings.HasPrefix(status, "D") {
			fileType = "deleted"
			contextData[file] = map[string]string{
				"type": fileType,
				"diff": "file deleted",
			}
			utils.Debug("[GIT.COMMIT.MSG]: File marked as deleted: '" + file + "'")
			continue
		}

		diffOutput, err := RunGitCmd(dir, nil, "diff", "--", file)
		if err != nil {
			utils.Error(fmt.Sprintf("[GIT.DIFF.FAIL]: Error running git diff for '%s': %s", file, err.Error()))
			return "", err
		}

		if strings.TrimSpace(diffOutput) == "" {
			diffOutput, err = RunGitCmd(dir, nil, "diff", "--cached", "--", file)
			if err != nil {
				utils.Error(fmt.Sprintf("[GIT.DIFF.FAIL]: Error running git diff --cached for '%s': %s", file, err.Error()))
				return "", err
			}
		}

		if strings.TrimSpace(diffOutput) == "" {
			contentBytes, err := os.ReadFile(file)
			if err != nil {
				utils.Error(fmt.Sprintf("[GIT.FILE.READ.FAIL]: Error reading new file '%s': %s", file, err.Error()))
				return "", err
			}
			diffOutput = string(contentBytes)
			fileType = "new"
		} else {
			fileType = "updated"
		}

		contextData[file] = map[string]string{
			"type": fileType,
			"diff": diffOutput,
		}

		utils.Debug("[GIT.COMMIT.MSG]: Processed file '" + file + "' as " + fileType)
	}

	message, err := di.GetGeminiRunner().SendToGemini(contextData, apiKey.(string))
	if err != nil {
		utils.Error("[GEMINI.FAIL]: Error generating group commit message: " + err.Error())
		return "", err
	}

	return message, nil
}

func BatchProcessGetMessages(allChangedFiles []string, rootFolder string) error {
	utils.Debug("[GIT.BATCH]: Starting batch processing of commit messages")
	
	// Separate binary and text files
	var binaryFiles []string
	var textFiles []string
	
	for _, file := range allChangedFiles {
		if utils.IsBinaryFile(file) {
			binaryFiles = append(binaryFiles, file)
		} else {
			textFiles = append(textFiles, file)
		}
	}
	
	// Handle binary files first
	if len(binaryFiles) > 0 {
		// Stop the creative loader to allow user interaction
		utils.StopCreativeLoader()
		
		utils.Info(fmt.Sprintf("🔍 Detected %d binary file(s):", len(binaryFiles)))
		for _, file := range binaryFiles {
			fileType := utils.GetBinaryFileType(file)
			utils.Info(fmt.Sprintf("  • %s (%s)", filepath.Base(file), fileType))
		}
		fmt.Println()
		
		// Create detailed prompt showing binary files
		promptMessage := fmt.Sprintf("Generate automated commit messages for these %d binary file(s)?", len(binaryFiles))
		generateBinaryMessages := utils.ConfirmAction(promptMessage, true)
		
		// Restart the creative loader after user input
		utils.StartCreativeLoader("Processing files", utils.ProcessingAnimation)
		utils.UpdateCreativeLoaderPhase("processing")
		
		if generateBinaryMessages {
			utils.Info("📝 Generating automated commit messages for binary files...")
			for _, file := range binaryFiles {
				// Get git status for this file
				cacheMu.RLock()
				status, cached := changedFilesCache[file]
				cacheMu.RUnlock()
				
				if !cached {
					// If not in cache, assume it's modified
					status = "M"
				}
				
				message := utils.GenerateBinaryCommitMessage(file, status)
				utils.Debug("[GIT.BATCH.BINARY]: Generated message for binary file: " + file + " - " + message)
				output.Set(file, rootFolder, message)
			}
			utils.Success(fmt.Sprintf("✅ Generated commit messages for %d binary file(s)", len(binaryFiles)))
		} else {
			utils.Info("⏭️ Skipping binary files. They will not have commit messages generated.")
		}
		fmt.Println()
	}
	
	// Process text files with Gemini API (only if there are any)
	if len(textFiles) == 0 {
		utils.Info("ℹ️ No text files to process with AI.")
		return nil
	}
	
	utils.Info(fmt.Sprintf("🤖 Processing %d text file(s) with AI...", len(textFiles)))
	
	var fileWg sync.WaitGroup
	var fileErrors []error
	fileMu := sync.Mutex{}

	for _, file := range textFiles {
		fileWg.Add(1)
		go func(file string) {
			defer fileWg.Done()

			utils.Debug("[GIT.BATCH]: Processing text file: " + file)
			message, err := GenCommitMessage([]string{file}, rootFolder) // <-- wrapped in slice
			if err != nil {
				utils.Error("[GIT.BATCH.FAIL]: Failed to generate commit message for file: " + file + " - " + err.Error())
				fileMu.Lock()
				fileErrors = append(fileErrors, err)
				fileMu.Unlock()
				return
			}

			utils.Debug("[GIT.BATCH.SUCCESS]: Generated commit message for file: " + file + " - " + message)
			output.Set(file, rootFolder, message)
		}(file)
	}

	fileWg.Wait()

	if len(fileErrors) > 0 {
		utils.Error("[GIT.BATCH.FAIL]: Batch processing completed with errors")
		return fmt.Errorf("one or more errors occurred while preparing commit messages")
	}

	utils.Success(fmt.Sprintf("✅ Generated AI commit messages for %d text file(s)", len(textFiles)))
	return nil
}

func CommitBatch(rootFolder output.Folder, env ...[]string) error {
	commitMessagesList := rootFolder.Files
	if len(commitMessagesList) == 0 {
		utils.Debug("[GIT.COMMIT]: No commit messages found for root folder: " + rootFolder.Name)
		return fmt.Errorf("no commit messages found for root folder: %s", rootFolder.Name)
	}

	utils.Debug("[GIT.COMMIT]: Starting batch commit in folder: " + rootFolder.Name)
	utils.Debug("[GIT.COMMIT]: Total files to commit: " + fmt.Sprint(len(commitMessagesList)))

	envMap := make(map[string]string)
	if len(env) > 0 {
		for _, pair := range env[0] {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}
	}

	messageToFiles := make(map[string][]string)
	for _, entry := range commitMessagesList {
		utils.Debug("[GIT.COMMIT]: Staging file for grouping: " + entry.Name + " with message: " + entry.Message)
		messageToFiles[entry.Message] = append(messageToFiles[entry.Message], entry.Name)
	}

	for message, files := range messageToFiles {
		for _, file := range files {
			utils.Debug("[GIT.COMMIT]: Adding file to commit: " + file)
			if _, err := RunGitCmd(rootFolder.Name, envMap, "add", file); err != nil {
				utils.Error("[GIT.COMMIT.FAIL]: Failed to add file to commit: " + err.Error())
				return fmt.Errorf("failed to add file to commit: %s", err.Error())
			}
		}

		utils.Debug(fmt.Sprintf("[GIT.COMMIT]: Committing %d file(s) with message: %s", len(files), message))
		if _, err := RunGitCmd(rootFolder.Name, envMap, "commit", "-m", message); err != nil {
			utils.Error("[GIT.COMMIT.FAIL]: Failed to commit files with message '" + message + "': " + err.Error())
			return fmt.Errorf("failed to commit files: %s", err.Error())
		}
	}

	output.RemoveFolder(rootFolder.Name)
	utils.Info("[GIT.COMMIT.SUCCESS]: Batch commit completed successfully and folder removed: " + rootFolder.Name)
	return nil
}

func PushBranch(rootFolderName string, branch string) error {
	if branch == "" {
		utils.Debug("[GIT.PUSH]: Branch name is empty, defaulting to 'main'")
		branch = "main"
	}

	utils.Debug("[GIT.PUSH]: Pushing branch: " + branch + " in folder: " + rootFolderName)
	if _, err := RunGitCmd(rootFolderName, nil, "push", "origin", branch); err != nil {
		utils.Error("[GIT.PUSH.FAIL]: Failed to push branch: " + err.Error())
		return fmt.Errorf("failed to push branch: %s", err.Error())
	}

	utils.Info("[GIT.PUSH.SUCCESS]: Branch pushed successfully")
	return nil
}

func GetFileDiff(filePath string, rootFolder string) (string, error) {
	cmdStatus := exec.Command("git", "-C", rootFolder, "status", "--porcelain", "--untracked-files=all", "--", filePath)

	var statusOut bytes.Buffer
	cmdStatus.Stdout = &statusOut
	cmdStatus.Stderr = &statusOut

	err := cmdStatus.Run()
	if err != nil {
		return "", fmt.Errorf("error checking status for file %s: %v", filePath, err)
	}

	if statusOut.String() != "" {
		return fmt.Sprintf("New untracked file: %s", filePath), nil
	}

	cmd := exec.Command("git", "-C", rootFolder, "diff", "--", filePath)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error getting diff for file %s: %v", filePath, err)
	}

	return out.String(), nil
}

func BatchProcessWithEmbeddings(allChangedFiles []string, rootFolder string, numClusters int) error {
	utils.Debug("[GIT.BATCH]: Starting batch processing with embeddings and clustering")

	// Separate binary and text files
	var binaryFiles []string
	var textFiles []string
	
	for _, file := range allChangedFiles {
		if utils.IsBinaryFile(file) {
			binaryFiles = append(binaryFiles, file)
		} else {
			textFiles = append(textFiles, file)
		}
	}
	
	// Handle binary files first
	if len(binaryFiles) > 0 {
		// Stop the creative loader to allow user interaction
		utils.StopCreativeLoader()
		
		utils.Info(fmt.Sprintf("🔍 Detected %d binary file(s) for grouped processing:", len(binaryFiles)))
		for _, file := range binaryFiles {
			fileType := utils.GetBinaryFileType(file)
			utils.Info(fmt.Sprintf("  • %s (%s)", filepath.Base(file), fileType))
		}
		fmt.Println()
		
		// Create detailed prompt showing binary files
		promptMessage := fmt.Sprintf("Generate automated commit messages for these %d binary file(s)?", len(binaryFiles))
		generateBinaryMessages := utils.ConfirmAction(promptMessage, true)
		
		// Restart the creative loader after user input
		utils.StartCreativeLoader("Processing files with clustering", utils.BrailleAnimation)
		utils.UpdateCreativeLoaderPhase("clustering")
		
		if generateBinaryMessages {
			utils.Info("📝 Generating automated commit messages for binary files...")
			for _, file := range binaryFiles {
				// Get git status for this file
				cacheMu.RLock()
				status, cached := changedFilesCache[file]
				cacheMu.RUnlock()
				
				if !cached {
					// If not in cache, assume it's modified
					status = "M"
				}
				
				message := utils.GenerateBinaryCommitMessage(file, status)
				utils.Debug("[GIT.BATCH.BINARY]: Generated message for binary file: " + file + " - " + message)
				output.Set(file, rootFolder, message)
			}
			utils.Success(fmt.Sprintf("✅ Generated commit messages for %d binary file(s)", len(binaryFiles)))
		} else {
			utils.Info("⏭️ Skipping binary files. They will not have commit messages generated.")
		}
		fmt.Println()
	}
	
	// Process text files with embeddings and clustering (only if there are any)
	if len(textFiles) == 0 {
		utils.Info("ℹ️ No text files to process with AI clustering.")
		return nil
	}
	
	utils.Info(fmt.Sprintf("🤖 Processing %d text file(s) with AI clustering...", len(textFiles)))

	var fileData []FileEmbedding
	var fileErrors []error
	var fileMu sync.Mutex

	for _, file := range textFiles {
		diff, err := GetFileDiff(file, rootFolder)
		if err != nil || strings.TrimSpace(diff) == "" {
			utils.Error("[GIT.BATCH]: Could not get diff for file: " + file)
			continue
		}

		embed, err := embeddings.GenerateEmbedding(diff)
		if err != nil {
			utils.Error("[GIT.BATCH]: Could not generate embedding for file: " + file)
			fileMu.Lock()
			fileErrors = append(fileErrors, err)
			fileMu.Unlock()
			continue
		}

		fileData = append(fileData, FileEmbedding{
			Path:      file,
			Diff:      diff,
			Embedding: embed,
		})
	}

	if len(fileData) == 0 {
		return fmt.Errorf("no valid diffs or embeddings generated")
	}

	utils.Debug(fmt.Sprintf("[GIT.BATCH]: Clustering %d files by embeddings", len(fileData)))

	vectors := make([][]float32, len(fileData))
	for i, f := range fileData {
		vectors[i] = f.Embedding
	}

	labels, err := embeddings.KMeans(vectors, numClusters, 10)
	if err != nil {
		return fmt.Errorf("clustering failed: %v", err)
	}

	groupMap := make(map[int][]FileEmbedding)
	for i, label := range labels {
		groupMap[label] = append(groupMap[label], fileData[i])
	}

	type CommitGroup struct {
		Message string   `json:"message"`
		Files   []string `json:"files"`
	}

	var commitGroups []CommitGroup
	var commitMu sync.Mutex
	var fileWg sync.WaitGroup

	for label, group := range groupMap {
		fileWg.Add(1)
		go func(label int, group []FileEmbedding) {
			defer fileWg.Done()

			utils.Debug(fmt.Sprintf("[GIT.BATCH]: Generating commit message for group %d with %d files", label, len(group)))

			var filePaths []string
			for _, f := range group {
				filePaths = append(filePaths, f.Path)
			}

			message, err := GenCommitMessage(filePaths, rootFolder)
			if err != nil {
				utils.Error(fmt.Sprintf("[GIT.BATCH]: Commit message generation failed for group %d - %s", label, err.Error()))
				fileMu.Lock()
				fileErrors = append(fileErrors, err)
				fileMu.Unlock()
				return
			}

			commitMu.Lock()
			commitGroups = append(commitGroups, CommitGroup{
				Message: message,
				Files:   filePaths,
			})
			commitMu.Unlock()

			for _, f := range group {
				utils.Debug("[GIT.BATCH.SUCCESS]: Generated commit message for file: " + f.Path + " - " + message)
				output.Set(f.Path, rootFolder, message)
			}
		}(label, group)
	}

	fileWg.Wait()

	if len(fileErrors) > 0 {
		utils.Error("[GIT.BATCH.FAIL]: Batch processing completed with errors")
		return fmt.Errorf("one or more errors occurred while preparing commit messages")
	}

	utils.Success(fmt.Sprintf("✅ Generated AI commit messages for %d text file(s) using clustering", len(textFiles)))
	return nil
}

// RunGitCmdWithTimeout executes a Git command with a timeout
func RunGitCmdWithTimeout(dir string, envVars map[string]string, timeout time.Duration, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	// Append custom environment variables to the existing environment
	if envVars != nil {
		env := cmd.Env
		for key, value := range envVars {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Create a channel to signal completion
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	// Wait for command completion or timeout
	select {
	case err := <-done:
		if err != nil {
			utils.Error(fmt.Sprintf(
				"[GIT.EXEC.TIMEOUT.FAIL]: Command failed: %s\nStdout: %s\nStderr: %s\n",
				err,
				stdout.String(),
				stderr.String(),
			))
			return "", err
		}
		utils.Debug("[GIT.EXEC.TIMEOUT.SUCCESS]: Command executed successfully in directory '" + dir + "': git " + strings.Join(args, " "))
		return stdout.String(), nil
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		err := fmt.Errorf("command timeout after %v", timeout)
		utils.Error("[GIT.EXEC.TIMEOUT]: Command timed out: git " + strings.Join(args, " "))
		return "", err
	}
}

// GetChangedFiles gets changed files from multiple root folders with concurrency control
func GetChangedFiles(rootFolders []string, maxConcurrency int, env ...[]string) ([]output.Folder, error) {
	if maxConcurrency <= 0 {
		maxConcurrency = 5 // Default concurrency
	}

	var folders []output.Folder
	folderChan := make(chan output.Folder, len(rootFolders))
	errorChan := make(chan error, len(rootFolders))

	// Use a semaphore to limit concurrency
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for _, rootFolder := range rootFolders {
		wg.Add(1)
		go func(folder string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			changedFiles, err := GetAllChangedFiles(folder)
			if err != nil {
				errorChan <- fmt.Errorf("failed to get changed files for %s: %w", folder, err)
				return
			}

			if len(changedFiles) > 0 {
				var fileEntries []output.FileEntry
				for _, file := range changedFiles {
					fileEntries = append(fileEntries, output.FileEntry{
						Name:    file,
						Message: "", // Message will be generated later
					})
				}

				folderChan <- output.Folder{
					Name:  folder,
					Files: fileEntries,
				}
			}
		}(rootFolder)
	}

	// Close channels when all goroutines are done
	go func() {
		wg.Wait()
		close(folderChan)
		close(errorChan)
	}()

	// Collect results
	for folder := range folderChan {
		folders = append(folders, folder)
	}

	// Check for errors
	select {
	case err := <-errorChan:
		return folders, err
	default:
		return folders, nil
	}
}

// Status gets the git status for multiple root paths
func Status(rootPaths []string) ([]output.Folder, error) {
	var folders []output.Folder

	for _, rootPath := range rootPaths {
		statusOutput, err := RunGitCmd(rootPath, nil, "status", "--porcelain")
		if err != nil {
			utils.Error("[GIT.STATUS.FAIL]: Failed to get status for " + rootPath + ": " + err.Error())
			continue
		}

		if strings.TrimSpace(statusOutput) == "" {
			utils.Debug("[GIT.STATUS]: No changes in " + rootPath)
			continue
		}

		var fileEntries []output.FileEntry
		lines := strings.Split(statusOutput, "\n")

		for _, line := range lines {
			if len(line) < 4 {
				continue
			}

			relativePath := strings.TrimSpace(line[3:])
			absolutePath := filepath.Join(rootPath, relativePath)

			fileEntries = append(fileEntries, output.FileEntry{
				Name:    absolutePath,
				Message: "", // Status doesn't include commit message
			})
		}

		if len(fileEntries) > 0 {
			folders = append(folders, output.Folder{
				Name:  rootPath,
				Files: fileEntries,
			})
		}
	}

	return folders, nil
}

// ProcessOneFile processes a single file with the given commit message
func ProcessOneFile(filePath, commitMessage string, env ...[]string) error {
	// Get the directory containing the file
	dir := filepath.Dir(filePath)
	fileName := filepath.Base(filePath)

	// Convert env slice to map
	envMap := make(map[string]string)
	if len(env) > 0 {
		for _, pair := range env[0] {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}
	}

	// Add the file to git
	_, err := RunGitCmd(dir, envMap, "add", fileName)
	if err != nil {
		utils.Error("[GIT.PROCESS.FAIL]: Failed to add file: " + err.Error())
		return fmt.Errorf("failed to add file %s: %w", filePath, err)
	}

	// Commit the file
	_, err = RunGitCmd(dir, envMap, "commit", "-m", commitMessage)
	if err != nil {
		utils.Error("[GIT.PROCESS.FAIL]: Failed to commit file: " + err.Error())
		return fmt.Errorf("failed to commit file %s: %w", filePath, err)
	}

	utils.Debug("[GIT.PROCESS.SUCCESS]: Successfully processed file: " + filePath)
	return nil
}

// IsGitRepository checks if the given path is a Git repository
func IsGitRepository(path string) bool {
	_, err := RunGitCmd(path, nil, "rev-parse", "--git-dir")
	return err == nil
}

// GetGitConfigValue gets a Git configuration value
func GetGitConfigValue(key string, env ...[]string) (string, error) {
	// Convert env slice to map
	envMap := make(map[string]string)
	if len(env) > 0 {
		for _, pair := range env[0] {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}
	}

	// Use current directory as default
	dir := "."

	output, err := RunGitCmd(dir, envMap, "config", "--get", key)
	if err != nil {
		return "", fmt.Errorf("failed to get git config value for %s: %w", key, err)
	}

	return strings.TrimSpace(output), nil
}

// SetGitConfigValue sets a Git configuration value
func SetGitConfigValue(key, value string, env ...[]string) error {
	// Convert env slice to map
	envMap := make(map[string]string)
	if len(env) > 0 {
		for _, pair := range env[0] {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}
	}

	// Use current directory as default
	dir := "."

	_, err := RunGitCmd(dir, envMap, "config", "--local", key, value)
	if err != nil {
		return fmt.Errorf("failed to set git config value %s=%s: %w", key, value, err)
	}

	utils.Debug("[GIT.CONFIG.SET]: Successfully set " + key + "=" + value)
	return nil
}
