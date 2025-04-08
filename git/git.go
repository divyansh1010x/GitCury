package git

import (
	"GitCury/config"
	"GitCury/output"
	"GitCury/utils"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

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

func GenCommitMessage(file string, dir string) (string, error) {
	var output string
	var err error
	var fileType string

	cacheMu.RLock()
	status, cached := changedFilesCache[file]
	cacheMu.RUnlock()

	if cached && strings.HasPrefix(status, "D") {
		fileType = "deleted"
		context := map[string]string{
			"file": file,
			"type": fileType,
		}

		utils.Debug("[GIT.COMMIT.MSG]: File marked as deleted: '" + file + "'")

		apiKey := config.Get("GEMINI_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("GEMINI_API_KEY")
			if apiKey == "" {
				return "", err
			}
		}
		message, err := utils.SendToGemini(context, apiKey.(string))
		if err != nil {
			utils.Error("[GEMINI.FAIL]: Error sending to Gemini: " + err.Error())
			return "Automated commit message: deleted " + file, nil
		}

		return message, nil
	}

	output, err = RunGitCmd(dir, nil, "diff", "--", file)
	if err != nil {
		utils.Error("[GIT.DIFF.FAIL]: Error running git diff for unstaged changes in file '" + file + "': " + err.Error())
		return "", err
	}

	if strings.TrimSpace(output) == "" {
		output, err = RunGitCmd(dir, nil, "diff", "--cached", "--", file)
		if err != nil {
			utils.Error("[GIT.DIFF.FAIL]: Error running git diff for staged changes in file '" + file + "': " + err.Error())
			return "", err
		}
	}

	if strings.TrimSpace(output) == "" {
		fileContent, err := os.ReadFile(file)
		if err != nil {
			utils.Error("[GIT.FILE.READ.FAIL]: Error reading untracked file '" + file + "': " + err.Error())
			return "", err
		}

		output = string(fileContent)
		fileType = "new"
	} else {
		fileType = "updated"
	}

	context := map[string]string{
		"file": file,
		"diff": output,
		"type": fileType,
	}

	utils.Debug("[GIT.COMMIT.MSG]: Generated diff for file '" + file + "'")

	apiKey := config.Get("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return "", err
		}
	}
	message, err := utils.SendToGemini(context, apiKey.(string))
	if err != nil {
		utils.Error("[GEMINI.FAIL]: Error sending to Gemini: " + err.Error())
		return "Automated commit message: changes made to " + file, nil
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
			message, err := GenCommitMessage(file, rootFolder)
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

	// utils.Info("[GIT.BATCH.SUCCESS]: Batch processing of commit messages completed successfully")
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

	for _, commit := range commitMessagesList {
		utils.Debug("[GIT.COMMIT]: Adding file to commit: " + commit.Name)
		envMap := make(map[string]string)
		if len(env) > 0 {
			for _, pair := range env[0] {
				parts := strings.SplitN(pair, "=", 2)
				if len(parts) == 2 {
					envMap[parts[0]] = parts[1]
				}
			}
		}

		if _, err := RunGitCmd(rootFolder.Name, envMap, "add", commit.Name); err != nil {
			utils.Error("[GIT.COMMIT.FAIL]: Failed to add file to commit: " + err.Error())
			return fmt.Errorf("failed to add file to commit: %s", err.Error())
		}

		utils.Debug("[GIT.COMMIT]: Committing file: " + commit.Name + " with message: " + commit.Message)
		if _, err := RunGitCmd(rootFolder.Name, envMap, "commit", "-m", commit.Message); err != nil {
			utils.Error("[GIT.COMMIT.FAIL]: Failed to commit file: " + err.Error())
			return fmt.Errorf("failed to commit file: %s", err.Error())
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
