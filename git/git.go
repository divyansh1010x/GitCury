package git

import (
	"GitCury/config"
	"GitCury/output"
	"GitCury/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

func RunGitCmd(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.Error("Error in running command 'git " + strings.Join(args, " ") + "' in directory '" + dir + "': " + err.Error())
		return "", err
	}

	utils.Info("Successfully ran git command in directory '" + dir + "': git " + strings.Join(args, " "))
	return string(output), nil
}

var changedFilesCache = make(map[string]string)
var cacheMu sync.RWMutex

func GetAllChangedFiles(dir string) ([]string, error) {
	output, err := RunGitCmd(dir, "status", "--porcelain")
	if err != nil {
		utils.Error("Failed to get git status: " + err.Error())
		return nil, err
	}

	if strings.TrimSpace(output) == "" {
		utils.Info("No changed files detected in directory: " + dir)
		return nil, nil
	}

	var changedFiles []string
	lines := strings.Split(output, "\n")
	cacheMu.Lock()
	defer cacheMu.Unlock()
	for _, line := range lines {
		if len(line) > 3 {
			status := line[:2]
			relativePath := line[3:]
			absolutePath := filepath.Join(dir, relativePath)

			changedFilesCache[absolutePath] = status

			if strings.HasPrefix(status, "D") {
				// Handle deleted files
				utils.Debug("File marked as deleted: " + absolutePath)
				changedFiles = append(changedFiles, absolutePath)
				continue
			}

			// Check if the path is a directory or file
			info, err := os.Stat(absolutePath)
			if err != nil {
				if os.IsNotExist(err) {
					utils.Debug("File does not exist (possibly deleted): " + absolutePath)
					changedFiles = append(changedFiles, absolutePath)
					continue
				}
				utils.Error("Failed to stat path '" + absolutePath + "': " + err.Error())
				return nil, err
			}

			if info.IsDir() {
				// Walk through the directory and collect all files
				err = filepath.Walk(absolutePath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						utils.Error("Error walking through directory '" + absolutePath + "': " + err.Error())
						return err
					}
					if !info.IsDir() {
						changedFiles = append(changedFiles, path)
					}
					return nil
				})
				if err != nil {
					return nil, err
				}
			} else {
				changedFiles = append(changedFiles, absolutePath)
			}
		}
	}

	utils.Debug("Changed files: " + strings.Join(changedFiles, ", "))
	return changedFiles, nil
}

func GenCommitMessage(file string, dir string) (string, error) {
	var output string
	var err error
	var fileType string

	cacheMu.RLock()
	status, cached := changedFilesCache[file]
	cacheMu.RUnlock()

	if cached && strings.HasPrefix(status, "D") {
		fileType = "deleted" // File is marked as deleted
		context := map[string]string{
			"file": file,
			"type": fileType,
		}

		utils.Debug("File marked as deleted: '" + file + "'")

		apiKey := config.Get("GEMINI_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("GEMINI_API_KEY")
			if apiKey == "" {
				return "", err
			}
		}
		message, err := utils.SendToGemini(context, apiKey.(string))
		if err != nil {
			utils.Error("Error sending to Gemini: " + err.Error())
			return "Automated commit message: deleted " + file, nil
		}

		return message, nil
	}

	// Handle other file types (modified, new, etc.)
	output, err = RunGitCmd(dir, "diff", "--", file)
	if err != nil {
		utils.Error("Error running git diff for unstaged changes in file '" + file + "': " + err.Error())
		return "", err
	}

	if strings.TrimSpace(output) == "" {
		output, err = RunGitCmd(dir, "diff", "--cached", "--", file)
		if err != nil {
			utils.Error("Error running git diff for staged changes in file '" + file + "': " + err.Error())
			return "", err
		}
	}

	if strings.TrimSpace(output) == "" {
		fileContent, err := os.ReadFile(file)
		if err != nil {
			utils.Error("Error reading untracked file '" + file + "': " + err.Error())
			return "", err
		}

		output = string(fileContent)
		fileType = "new" // File is untracked, hence it's new
	} else {
		fileType = "updated" // File has changes, hence it's updated
	}

	context := map[string]string{
		"file": file,
		"diff": output,
		"type": fileType,
	}

	utils.Debug("Generated diff for file '" + file + "'")

	apiKey := config.Get("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return "", err
		}
	}
	message, err := utils.SendToGemini(context, apiKey.(string))
	if err != nil {
		utils.Error("Error sending to Gemini: " + err.Error())
		return "Automated commit message: changes made to " + file, nil
	}

	return message, nil
}

func BatchProcessGetMessages(allChangedFiles []string, rootFolder string) error {
	utils.Info("Starting batch processing of commit messages")
	var fileWg sync.WaitGroup
	var fileErrors []error
	fileMu := sync.Mutex{}

	for _, file := range allChangedFiles {
		fileWg.Add(1)
		go func(file string) {
			defer fileWg.Done()

			utils.Debug("Processing file: " + file)
			message, err := GenCommitMessage(file, rootFolder)
			if err != nil {
				utils.Error("Failed to generate commit message for file: " + file + " - " + err.Error())
				fileMu.Lock()
				fileErrors = append(fileErrors, err)
				fileMu.Unlock()
				return
			}

			utils.Debug("Generated commit message for file: " + file + " - " + message)
			output.Set(file, rootFolder, message)
		}(file)
	}

	fileWg.Wait()

	if len(fileErrors) > 0 {
		utils.Error("Batch processing completed with errors")
		return fmt.Errorf("one or more errors occurred while preparing commit messages")
	}

	utils.Info("Batch processing of commit messages completed successfully")
	return nil
}

func CommitBatch(rootFolder output.Folder) error {
	commitMessagesList := rootFolder.Files
	if len(commitMessagesList) == 0 {
		utils.Info("No commit messages found for root folder: " + rootFolder.Name)
		return fmt.Errorf("no commit messages found for root folder: %s", rootFolder.Name)
	}
	for _, commit := range commitMessagesList {
		utils.Debug("Adding file to commit: " + commit.Name)
		if _, err := RunGitCmd(rootFolder.Name, "add", commit.Name); err != nil {
			utils.Error("Failed to add file to commit: " + err.Error())
			return fmt.Errorf("failed to add file to commit: %s", err.Error())
		}

		utils.Debug("Committing file: " + commit.Name + " with message: " + commit.Message)
		if _, err := RunGitCmd(rootFolder.Name, "commit", "-m", commit.Message); err != nil {
			utils.Error("Failed to commit file: " + err.Error())
			return fmt.Errorf("failed to commit file: %s", err.Error())
		}

		output.Delete(commit.Name, rootFolder.Name)
	}

	utils.Info("Batch commit completed successfully")
	return nil
}

func PushBranch(rootFolderName string, branch string) error {
	if branch == "" {
		branch = "main"
	}

	utils.Debug("Pushing branch: " + branch + " in folder: " + rootFolderName)
	if _, err := RunGitCmd(rootFolderName, "push", "origin", branch); err != nil {
		utils.Error("Failed to push branch: " + err.Error())
		return fmt.Errorf("failed to push branch: %s", err.Error())
	}

	utils.Info("Branch pushed successfully")
	return nil
}
