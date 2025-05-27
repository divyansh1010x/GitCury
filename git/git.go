package git

import (
	"GitCury/config"
	"GitCury/embeddings"
	"GitCury/output"
	"GitCury/utils"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type FileEmbedding struct {
	Path      string
	Diff      string
	Embedding []float32
}

// Repository-level mutex system to prevent concurrent Git operations on the same repository
var (
	repoMutexes      = make(map[string]*sync.Mutex)
	repoMutexMapLock sync.RWMutex
)

// getRepoMutex returns a mutex for the given repository directory
func getRepoMutex(repoDir string) *sync.Mutex {
	// Clean the path to ensure consistency
	cleanPath := filepath.Clean(repoDir)

	repoMutexMapLock.RLock()
	mutex, exists := repoMutexes[cleanPath]
	repoMutexMapLock.RUnlock()

	if exists {
		return mutex
	}

	// If mutex doesn't exist, create it
	repoMutexMapLock.Lock()
	defer repoMutexMapLock.Unlock()

	// Double-check in case another goroutine created it
	if mutex, exists := repoMutexes[cleanPath]; exists {
		return mutex
	}

	// Create new mutex for this repository
	mutex = &sync.Mutex{}
	repoMutexes[cleanPath] = mutex
	utils.Debug(fmt.Sprintf("[GIT.MUTEX]: Created repository mutex for: %s", cleanPath))
	return mutex
}

// withRepoLock executes a function while holding the repository-level lock
func withRepoLock(repoDir string, operation string, fn func() error) error {
	mutex := getRepoMutex(repoDir)

	utils.Debug(fmt.Sprintf("[GIT.MUTEX]: Acquiring lock for %s in repo: %s", operation, repoDir))
	mutex.Lock()
	defer func() {
		mutex.Unlock()
		utils.Debug(fmt.Sprintf("[GIT.MUTEX]: Released lock for %s in repo: %s", operation, repoDir))
	}()

	return fn()
}

func RunGitCmd(dir string, envVars map[string]string, args ...string) (string, error) {
	return RunGitCmdWithTimeout(dir, envVars, 30*time.Second, args...)
}

func RunGitCmdWithTimeout(dir string, envVars map[string]string, timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	// Create environment with existing env vars
	env := os.Environ()

	// Append custom environment variables
	if envVars != nil {
		for key, value := range envVars {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	commandStr := "git " + strings.Join(args, " ")
	utils.Debug("[GIT.EXEC]: Running git command in '" + dir + "': " + commandStr)

	err := cmd.Run()

	// Check if the error is due to timeout
	if ctx.Err() == context.DeadlineExceeded {
		return "", utils.NewGitError(
			"Git command timed out after "+timeout.String(),
			ctx.Err(),
			map[string]interface{}{
				"directory": dir,
				"command":   commandStr,
				"timeout":   timeout.String(),
			},
			dir,
		)
	}

	if err != nil {
		errOutput := stderr.String()
		stdOutput := stdout.String()

		utils.Error(fmt.Sprintf(
			"[GIT.EXEC.FAIL]: Command failed: %s\nStdout: %s\nStderr: %s\n",
			err,
			stdOutput,
			errOutput,
		), dir)

		// Create structured error with context
		return "", utils.NewGitError(
			"Git command failed",
			err,
			map[string]interface{}{
				"directory": dir,
				"command":   commandStr,
				"stdout":    stdOutput,
				"stderr":    errOutput,
			},
			dir,
		)
	}

	utils.Debug("[GIT.EXEC.SUCCESS]: Command executed successfully in directory '" + dir + "': " + commandStr)
	return stdout.String(), nil
}

var changedFilesCache = make(map[string]string)
var cacheMu sync.RWMutex

func GetAllChangedFiles(dir string) ([]string, error) {
	output, err := RunGitCmd(dir, nil, "status", "--porcelain")
	if err != nil {
		utils.Error("[GIT.STATUS.FAIL]: Failed to get git status: "+err.Error(), dir)
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
					// Check if file should be ignored
					if IsIgnoredFile(absInner) {
						utils.Debug("[GIT.IGNORED]: Skipping ignored file: " + absInner)
						continue
					}

					// Check if file is binary
					if IsBinaryFile(absInner) {
						utils.Debug("[GIT.BINARY]: Skipping binary file: " + absInner)
						continue
					}

					changedFiles = append(changedFiles, absInner)
					changedFilesCache[absInner] = "??"
				}
			}
		} else {
			// Check if file should be ignored
			if IsIgnoredFile(abs) {
				utils.Debug("[GIT.IGNORED]: Skipping ignored file: " + abs)
				continue
			}

			// Check if file is binary (only for existing files)
			if !strings.HasPrefix(status, "D") && IsBinaryFile(abs) {
				utils.Debug("[GIT.BINARY]: Skipping binary file: " + abs)
				continue
			}

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
			utils.Error(fmt.Sprintf("[GIT.DIFF.FAIL]: Error running git diff for '%s': %s", file, err.Error()), file)
			return "", err
		}

		if strings.TrimSpace(diffOutput) == "" {
			diffOutput, err = RunGitCmd(dir, nil, "diff", "--cached", "--", file)
			if err != nil {
				utils.Error(fmt.Sprintf("[GIT.DIFF.FAIL]: Error running git diff --cached for '%s': %s", file, err.Error()), file)
				return "", err
			}
		}

		if strings.TrimSpace(diffOutput) == "" {
			contentBytes, err := os.ReadFile(file)
			if err != nil {
				utils.Error(fmt.Sprintf("[GIT.FILE.READ.FAIL]: Error reading new file '%s': %s", file, err.Error()), file)
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

	// Check for custom commit instructions
	var commitInstructions string
	if instructions, ok := config.Get("commit_instructions").(string); ok {
		commitInstructions = instructions
		utils.Debug("[GIT.COMMIT.MSG]: Using custom commit instructions from config")
	}

	message, err := utils.SendToGemini(contextData, apiKey.(string), commitInstructions)
	if err != nil {
		// Enhanced error reporting with file details
		filesList := make([]string, 0, len(files))
		fileDetails := make([]string, 0, len(files))

		for _, file := range files {
			fileName := filepath.Base(file)
			filesList = append(filesList, fileName)

			// Add file diagnostic information
			if info, statErr := os.Stat(file); statErr == nil {
				size := info.Size()
				isBinary := IsBinaryFile(file)
				isIgnored := IsIgnoredFile(file)
				fileDetails = append(fileDetails, fmt.Sprintf("%s (size: %d bytes, binary: %v, ignored: %v)",
					fileName, size, isBinary, isIgnored))
			} else {
				fileDetails = append(fileDetails, fmt.Sprintf("%s (stat error: %v)", fileName, statErr))
			}
		}

		filesInfo := strings.Join(filesList, ", ")
		utils.Error("[GEMINI.FAIL]: Error generating group commit message: "+err.Error(), filesInfo)
		utils.Debug("[GEMINI.FAIL.DETAILS]: File details: " + strings.Join(fileDetails, "; "))

		return "", err
	}

	return message, nil
}

// GenCommitMessageWithContext generates a commit message for multiple files with enhanced context
func GenCommitMessageWithContext(files []string, dir string, contextPrompt string) (string, error) {
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
			utils.Debug(fmt.Sprintf("[GIT.DIFF.FAIL]: Error running git diff for '%s': %s", file, err.Error()))
			// Continue with other methods instead of failing
		}

		if strings.TrimSpace(diffOutput) == "" {
			diffOutput, err = RunGitCmd(dir, nil, "diff", "--cached", "--", file)
			if err != nil {
				utils.Debug(fmt.Sprintf("[GIT.DIFF.FAIL]: Error running git diff --cached for '%s': %s", file, err.Error()))
			}
		}

		if strings.TrimSpace(diffOutput) == "" {
			contentBytes, err := os.ReadFile(file)
			if err != nil {
				utils.Debug(fmt.Sprintf("[GIT.FILE.READ.FAIL]: Error reading new file '%s': %s", file, err.Error()))
				// Use filename as fallback context
				diffOutput = "new file: " + filepath.Base(file)
			} else {
				diffOutput = string(contentBytes)
			}
			fileType = "new"
		} else {
			fileType = "updated"
		}

		// Limit diff size to prevent overwhelming the API
		const maxDiffSize = 5000
		if len(diffOutput) > maxDiffSize {
			diffOutput = diffOutput[:maxDiffSize] + "\n... (truncated)"
		}

		contextData[file] = map[string]string{
			"type": fileType,
			"diff": diffOutput,
		}

		utils.Debug("[GIT.COMMIT.MSG]: Processed file '" + file + "' as " + fileType)
	}

	// Check for custom commit instructions
	var commitInstructions string
	if instructions, ok := config.Get("commit_instructions").(string); ok {
		commitInstructions = instructions + "\n\nAdditional context: " + contextPrompt
		utils.Debug("[GIT.COMMIT.MSG]: Using enhanced custom commit instructions with grouping context")
	} else {
		commitInstructions = contextPrompt
	}

	message, err := utils.SendToGemini(contextData, apiKey.(string), commitInstructions)
	if err != nil {
		// Enhanced error reporting with file details
		filesList := make([]string, 0, len(files))
		fileDetails := make([]string, 0, len(files))

		for _, file := range files {
			fileName := filepath.Base(file)
			filesList = append(filesList, fileName)

			// Add file diagnostic information
			if info, statErr := os.Stat(file); statErr == nil {
				size := info.Size()
				isBinary := IsBinaryFile(file)
				isIgnored := IsIgnoredFile(file)
				fileDetails = append(fileDetails, fmt.Sprintf("%s (size: %d bytes, binary: %v, ignored: %v)",
					fileName, size, isBinary, isIgnored))
			} else {
				fileDetails = append(fileDetails, fmt.Sprintf("%s (stat error: %v)", fileName, statErr))
			}
		}

		filesInfo := strings.Join(filesList, ", ")
		utils.Error("[GEMINI.FAIL]: Error generating context-enhanced commit message: "+err.Error(), filesInfo)
		utils.Debug("[GEMINI.FAIL.DETAILS]: File details: " + strings.Join(fileDetails, "; "))

		return "", err
	}

	return message, nil
}

func BatchProcessGetMessages(allChangedFiles []string, rootFolder string) error {
	utils.Debug("[GIT.BATCH]: Starting batch processing of commit messages")
	var fileWg sync.WaitGroup
	var fileErrors []error
	fileMu := sync.Mutex{}

	for _, file := range allChangedFiles {
		fileWg.Add(1)
		go func(file string) {
			defer fileWg.Done()

			utils.Debug("[GIT.BATCH]: Processing file: " + file)

			// Add pre-processing file validation
			if info, statErr := os.Stat(file); statErr == nil {
				utils.Debug(fmt.Sprintf("[GIT.BATCH.FILE]: %s (size: %d bytes, binary: %v, ignored: %v)",
					filepath.Base(file), info.Size(), IsBinaryFile(file), IsIgnoredFile(file)))
			}

			message, err := GenCommitMessage([]string{file}, rootFolder) // <-- wrapped in slice
			if err != nil {
				// Enhanced error reporting with file analysis
				fileInfo := filepath.Base(file)
				if info, statErr := os.Stat(file); statErr == nil {
					fileInfo = fmt.Sprintf("%s (size: %d bytes, binary: %v)",
						filepath.Base(file), info.Size(), IsBinaryFile(file))
				}

				utils.Error("[GIT.BATCH.FAIL]: Failed to generate commit message for file: "+file+" - "+err.Error(), file)
				utils.Debug("[GIT.BATCH.FAIL.ANALYSIS]: File analysis: " + fileInfo)
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
		// Collect file information from errors
		errorFileNames := make([]string, 0, len(fileErrors))
		for _, err := range fileErrors {
			if structErr, ok := err.(*utils.StructuredError); ok && structErr.ProcessedFile != "" {
				// Use basename to make the message cleaner
				errorFileNames = append(errorFileNames, filepath.Base(structErr.ProcessedFile))
			}
		}

		filesInfo := "multiple files"
		if len(errorFileNames) > 0 {
			filesInfo = strings.Join(errorFileNames, ", ")
		}

		utils.Error("[GIT.BATCH.FAIL]: Batch processing completed with errors", filesInfo)
		return fmt.Errorf("one or more errors occurred while preparing commit messages")
	}

	return nil
}

func CommitBatch(rootFolder output.Folder, env ...[]string) error {
	commitMessagesList := rootFolder.Files
	if len(commitMessagesList) == 0 {
		utils.Debug("[GIT.COMMIT]: No commit messages found for root folder: " + rootFolder.Name)
		return utils.NewValidationError(
			"No commit messages found for root folder",
			nil,
			map[string]interface{}{
				"folderName": rootFolder.Name,
			},
			rootFolder.Name,
		)
	}

	utils.Debug("[GIT.COMMIT]: Starting batch commit in folder: " + rootFolder.Name)
	utils.Debug("[GIT.COMMIT]: Total files to commit: " + fmt.Sprint(len(commitMessagesList)))

	// Convert environment slice to map
	envMap := make(map[string]string)
	if len(env) > 0 {
		for _, pair := range env[0] {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}
	}

	// Group files by commit message for efficiency
	messageToFiles := make(map[string][]string)
	for _, entry := range commitMessagesList {
		utils.Debug("[GIT.COMMIT]: Staging file for grouping: " + entry.Name + " with message: " + entry.Message)
		messageToFiles[entry.Message] = append(messageToFiles[entry.Message], entry.Name)
	}

	// Process commits sequentially to avoid Git index conflicts
	var errMu sync.Mutex
	var commitErrs []error

	// Process each commit sequentially within the repository to avoid Git conflicts
	for message, files := range messageToFiles {
		utils.Debug(fmt.Sprintf("[GIT.COMMIT]: Processing commit with %d files: %s", len(files), message))

		// Use repository-level locking to serialize Git operations within the same repo
		err := withRepoLock(rootFolder.Name, "stage_and_commit", func() error {
			// Stage all files for this commit sequentially
			for _, file := range files {
				utils.Debug("[GIT.COMMIT]: Adding file to commit: " + file)

				// Use SafeGitOperation to handle index.lock and other recovery scenarios
				err := SafeGitOperation(rootFolder.Name, "add file", func() error {
					_, gitErr := RunGitCmdWithTimeout(rootFolder.Name, envMap, 15*time.Second, "add", file)
					return gitErr
				})

				if err != nil {
					utils.Error("[GIT.COMMIT.FAIL]: Failed to add file to commit: "+err.Error(), file)
					return utils.NewGitError(
						"Failed to stage file",
						err,
						map[string]interface{}{
							"file":   file,
							"folder": rootFolder.Name,
						},
						file,
					)
				}
			}

			// Perform the actual commit after all files are staged
			utils.Debug(fmt.Sprintf("[GIT.COMMIT]: Committing %d file(s) with message: %s", len(files), message))

			// Use SafeGitOperation to handle index.lock and other recovery scenarios
			err := SafeGitOperation(rootFolder.Name, "commit", func() error {
				_, gitErr := RunGitCmdWithTimeout(rootFolder.Name, envMap, 30*time.Second, "commit", "-m", message)
				return gitErr
			})

			if err != nil {
				// Join filenames for error context
				filesList := strings.Join(files, ", ")
				utils.Error("[GIT.COMMIT.FAIL]: Failed to commit files with message '"+message+"': "+err.Error(), filesList)
				return utils.NewGitError(
					"Failed to commit files",
					err,
					map[string]interface{}{
						"message":   message,
						"folder":    rootFolder.Name,
						"fileCount": len(files),
					},
					filesList,
				)
			}

			return nil
		})

		if err != nil {
			errMu.Lock()
			commitErrs = append(commitErrs, err)
			errMu.Unlock()
		}
	}

	// Check if there were any errors
	if len(commitErrs) > 0 {
		// Collect file information from errors
		errorFileNames := make([]string, 0, len(commitErrs))
		for _, err := range commitErrs {
			if structErr, ok := err.(*utils.StructuredError); ok && structErr.ProcessedFile != "" {
				// Use basename to make the message cleaner
				errorFileNames = append(errorFileNames, filepath.Base(structErr.ProcessedFile))
			}
		}

		filesInfo := rootFolder.Name
		if len(errorFileNames) > 0 {
			filesInfo = strings.Join(errorFileNames, ", ")
		}

		return utils.NewGitError(
			"Failed to stage one or more files for commit",
			fmt.Errorf("%d errors occurred during staging", len(commitErrs)),
			map[string]interface{}{
				"folder": rootFolder.Name,
				"errors": commitErrs,
			},
			filesInfo,
		)
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

	// Use SafeGitOperation to handle index.lock and other recovery scenarios
	err := SafeGitOperation(rootFolderName, "push", func() error {
		_, gitErr := RunGitCmd(rootFolderName, nil, "push", "origin", branch)
		return gitErr
	})

	if err != nil {
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
	utils.Debug("[GIT.BATCH]: Starting intelligent batch processing with semantic grouping")

	// First, filter out files that shouldn't be processed
	filteredFiles := make([]string, 0, len(allChangedFiles))
	for _, file := range allChangedFiles {
		if IsIgnoredFile(file) || IsBinaryFile(file) {
			utils.Debug(fmt.Sprintf("[GIT.BATCH]: Skipping ignored/binary file: %s", file))
			continue
		}
		filteredFiles = append(filteredFiles, file)
	}

	if len(filteredFiles) == 0 {
		utils.Warning("[GIT.BATCH]: No suitable files found for processing after filtering")
		return nil
	}

	utils.Debug(fmt.Sprintf("[GIT.BATCH]: Processing %d filtered files (from %d total)", len(filteredFiles), len(allChangedFiles)))

	// Try semantic grouping first, fallback to simple processing if it fails
	err := attemptSemanticGrouping(filteredFiles, rootFolder, numClusters)
	if err != nil {
		utils.Warning(fmt.Sprintf("[GIT.BATCH]: Semantic grouping failed (%s), falling back to simple processing", err.Error()))
		return fallbackToSimpleProcessing(filteredFiles, rootFolder)
	}

	return nil
}

// attemptSemanticGrouping tries to group files using embeddings and semantic similarity
func attemptSemanticGrouping(files []string, rootFolder string, numClusters int) error {
	utils.Debug("[GIT.BATCH]: Attempting semantic grouping with embeddings")

	var fileData []FileEmbedding
	var processedFiles []string
	var embeddingErrors []error

	// Rate limiter for embedding generation to prevent API quota issues
	const maxConcurrentEmbeddings = 1 // Reduced from 2 to prevent rate limits
	semaphore := make(chan struct{}, maxConcurrentEmbeddings)
	var embeddingWg sync.WaitGroup
	var dataMu sync.Mutex

	for _, file := range files {
		embeddingWg.Add(1)
		go func(file string) {
			defer embeddingWg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Increased delay to prevent overwhelming the API (rate limiting)
			time.Sleep(5000 * time.Millisecond) // Increased from 2000ms to 5000ms

			diff, err := GetFileDiff(file, rootFolder)
			if err != nil || strings.TrimSpace(diff) == "" {
				utils.Debug(fmt.Sprintf("[GIT.BATCH]: Skipping file with no diff: %s", file))
				return
			}

			// Limit diff size for embedding generation
			const maxDiffSize = 10000 // 10KB limit
			if len(diff) > maxDiffSize {
				diff = diff[:maxDiffSize] + "\n... (truncated for embedding)"
			}

			embed, err := embeddings.GenerateEmbedding(diff)
			if err != nil {
				// Check for rate limit errors specifically
				if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "quota") || strings.Contains(err.Error(), "rate limit") {
					utils.Warning(fmt.Sprintf("[GIT.BATCH]: Rate limit hit for file %s, may need to reduce processing speed", file))
				}
				utils.Debug(fmt.Sprintf("[GIT.BATCH]: Embedding failed for file %s: %s", file, err.Error()))
				dataMu.Lock()
				embeddingErrors = append(embeddingErrors, err)
				dataMu.Unlock()
				return
			}

			dataMu.Lock()
			fileData = append(fileData, FileEmbedding{
				Path:      file,
				Diff:      diff,
				Embedding: embed,
			})
			processedFiles = append(processedFiles, file)
			dataMu.Unlock()

			utils.Debug(fmt.Sprintf("[GIT.BATCH]: Successfully generated embedding for: %s", file))
		}(file)
	}

	embeddingWg.Wait()

	// If we have too few successful embeddings, fall back to simple processing
	minFilesForClustering := 3
	if len(fileData) < minFilesForClustering {
		return fmt.Errorf("insufficient files for semantic grouping (%d/%d successful, need at least %d)",
			len(fileData), len(files), minFilesForClustering)
	}

	// Adjust cluster count based on actual data
	actualClusters := numClusters
	if len(fileData) < numClusters {
		actualClusters = len(fileData)
		utils.Debug(fmt.Sprintf("[GIT.BATCH]: Adjusting clusters from %d to %d based on available files", numClusters, actualClusters))
	}

	// Perform semantic grouping
	utils.Debug(fmt.Sprintf("[GIT.BATCH]: Clustering %d files into %d groups", len(fileData), actualClusters))

	vectors := make([][]float32, len(fileData))
	for i, f := range fileData {
		vectors[i] = f.Embedding
	}

	labels, err := embeddings.KMeans(vectors, actualClusters, 20)
	if err != nil {
		return fmt.Errorf("clustering failed: %v", err)
	}

	// Group files by cluster
	groupMap := make(map[int][]FileEmbedding)
	for i, label := range labels {
		groupMap[label] = append(groupMap[label], fileData[i])
	}

	// Generate commit messages for each group
	return generateGroupCommitMessages(groupMap, rootFolder)
}

// generateGroupCommitMessages generates commit messages for clustered file groups
func generateGroupCommitMessages(groupMap map[int][]FileEmbedding, rootFolder string) error {
	type CommitGroup struct {
		Message string   `json:"message"`
		Files   []string `json:"files"`
	}

	var commitGroups []CommitGroup
	var commitMu sync.Mutex
	var fileWg sync.WaitGroup
	var groupErrors []error
	var errorMu sync.Mutex

	// Rate limiter for commit message generation
	const maxConcurrentMessages = 3
	semaphore := make(chan struct{}, maxConcurrentMessages)

	for label, group := range groupMap {
		fileWg.Add(1)
		go func(label int, group []FileEmbedding) {
			defer fileWg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Add delay between requests
			time.Sleep(500 * time.Millisecond)

			utils.Debug(fmt.Sprintf("[GIT.BATCH]: Generating commit message for group %d with %d files", label, len(group)))

			var filePaths []string
			var groupContext strings.Builder
			for i, f := range group {
				filePaths = append(filePaths, f.Path)
				if i < 3 { // Include diff context for first 3 files
					groupContext.WriteString(fmt.Sprintf("File: %s\nDiff excerpt: %s\n\n",
						filepath.Base(f.Path), truncateString(f.Diff, 200)))
				}
			}

			// Enhanced prompt for grouped files
			contextPrompt := fmt.Sprintf(`Generate a commit message for this group of related files:
Files: %s

Context: %s

The files were grouped together based on semantic similarity of their changes.`,
				strings.Join(getFileBasenames(filePaths), ", "),
				groupContext.String())

			message, err := GenCommitMessageWithContext(filePaths, rootFolder, contextPrompt)
			if err != nil {
				utils.Error(fmt.Sprintf("[GIT.BATCH]: Commit message generation failed for group %d - %s", label, err.Error()))
				errorMu.Lock()
				groupErrors = append(groupErrors, err)
				errorMu.Unlock()
				return
			}

			commitMu.Lock()
			commitGroups = append(commitGroups, CommitGroup{
				Message: message,
				Files:   filePaths,
			})
			commitMu.Unlock()

			// Set the same commit message for all files in the group
			for _, f := range group {
				utils.Debug(fmt.Sprintf("[GIT.BATCH.SUCCESS]: Generated grouped commit message for file: %s - %s", f.Path, message))
				output.Set(f.Path, rootFolder, message)
			}

			utils.Success(fmt.Sprintf("[GIT.BATCH]: Successfully processed group %d (%d files): %s",
				label, len(group), truncateString(message, 60)))
		}(label, group)
	}

	fileWg.Wait()

	if len(groupErrors) > 0 {
		return fmt.Errorf("failed to generate commit messages for %d groups", len(groupErrors))
	}

	utils.Success(fmt.Sprintf("[GIT.BATCH]: Successfully generated %d grouped commit messages", len(commitGroups)))
	return nil
}

// fallbackToSimpleProcessing processes files individually when semantic grouping fails
func fallbackToSimpleProcessing(files []string, rootFolder string) error {
	utils.Debug("[GIT.BATCH]: Using simple processing fallback")

	var fileWg sync.WaitGroup
	var fileErrors []error
	var errorMu sync.Mutex

	// Rate limiter for simple processing
	const maxConcurrentSimple = 1 // Reduced from 5 to prevent rate limits
	semaphore := make(chan struct{}, maxConcurrentSimple)

	for _, file := range files {
		fileWg.Add(1)
		go func(file string) {
			defer fileWg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Increased delay for rate limiting
			time.Sleep(3000 * time.Millisecond) // Increased from 100ms to 3000ms

			utils.Debug(fmt.Sprintf("[GIT.BATCH]: Processing file individually: %s", file))

			message, err := GenCommitMessage([]string{file}, rootFolder)
			if err != nil {
				utils.Error(fmt.Sprintf("[GIT.BATCH]: Failed to generate commit message for file: %s - %s", file, err.Error()))
				errorMu.Lock()
				fileErrors = append(fileErrors, err)
				errorMu.Unlock()
				return
			}

			utils.Debug(fmt.Sprintf("[GIT.BATCH.SUCCESS]: Generated individual commit message for file: %s - %s", file, message))
			output.Set(file, rootFolder, message)
		}(file)
	}

	fileWg.Wait()

	if len(fileErrors) > 0 {
		utils.Warning(fmt.Sprintf("[GIT.BATCH]: Simple processing completed with %d errors out of %d files", len(fileErrors), len(files)))
		// Don't return error for simple processing - some success is better than none
	}

	return nil
}

// Helper functions for the enhanced grouping system
func getFileBasenames(filePaths []string) []string {
	basenames := make([]string, len(filePaths))
	for i, path := range filePaths {
		basenames[i] = filepath.Base(path)
	}
	return basenames
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// IsBinaryFile checks if a file is binary by reading the first few bytes
func IsBinaryFile(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false // If we can't open it, assume it's not binary for now
	}
	defer file.Close()

	// Read first 512 bytes to check for binary content
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	// Check for null bytes which indicate binary content
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return true
		}
	}

	// Check for high percentage of non-printable characters
	nonPrintable := 0
	for i := 0; i < n; i++ {
		if buffer[i] < 32 && buffer[i] != 9 && buffer[i] != 10 && buffer[i] != 13 {
			nonPrintable++
		}
	}

	// If more than 30% of the characters are non-printable, consider it binary
	return float64(nonPrintable)/float64(n) > 0.3
}

// IsIgnoredFile checks if a file should be ignored based on its extension or name
func IsIgnoredFile(filePath string) bool {
	fileName := filepath.Base(filePath)
	ext := filepath.Ext(filePath)

	// Common binary file extensions
	binaryExtensions := []string{
		".exe", ".bin", ".dll", ".so", ".dylib", ".a", ".o", ".obj",
		".zip", ".tar", ".gz", ".rar", ".7z",
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".ico", ".svg",
		".mp3", ".mp4", ".avi", ".mov", ".wmv", ".flv",
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
	}

	for _, binExt := range binaryExtensions {
		if strings.EqualFold(ext, binExt) {
			return true
		}
	}

	// Ignore common build artifacts and executables without extensions
	ignoreNames := []string{
		"gitcury", "gencli.exe", "gitcury.exe",
	}

	for _, ignoreName := range ignoreNames {
		if strings.EqualFold(fileName, ignoreName) {
			return true
		}
	}

	return false
}
