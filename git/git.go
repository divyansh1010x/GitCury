package git

import (
	"GitCury/config"
	"GitCury/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RunGitCmd(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.Error(string(output))
		return "", err
	}

	utils.Info("Successfully ran git command: git " + strings.Join(args, " "))
	return string(output), nil
}

func GetAllChangedFiles() ([]string, error) {
	output, err := RunGitCmd("status", "--porcelain")
	if err != nil {
		utils.Error("Failed to get git status: " + err.Error())
		return nil, err
	}

	var changedFiles []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) > 3 {
			relativePath := line[3:]
			absolutePath, err := filepath.Abs(relativePath)
			if err != nil {
				utils.Error("Failed to get absolute path for file '" + relativePath + "': " + err.Error())
				return nil, err
			}

			// Check if the path is a directory
			info, err := os.Stat(absolutePath)
			if err != nil {
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

func GenCommitMessage(file string) (string, error) {
	var output string
	var err error
	var fileType string

	output, err = RunGitCmd("diff", "--", file)
	if err != nil {
		utils.Error("Error running git diff for unstaged changes in file '" + file + "': " + err.Error())
		return "", err
	}

	if strings.TrimSpace(output) == "" {
		output, err = RunGitCmd("diff", "--cached", "--", file)
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
