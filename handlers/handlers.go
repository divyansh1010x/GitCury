package handlers

import (
	"GitCury/config"
	"GitCury/git"
	"GitCury/output"
	"GitCury/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	var settings map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		utils.Error("Error decoding request: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for key, value := range settings {
		utils.Info("Setting " + key + " to " + fmt.Sprintf("%v", value))
		config.Set(key, value)
	}

	json.NewEncoder(w).Encode(config.GetAll())
}

func PrepareCommitMessagesHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.Error("Error decoding request: " + err.Error())
		// http.Error(w, err.Error(), http.StatusBadRequest)
		// return
	}

	numFilesToCommit := 10 // Default value
	if value, ok := requestBody["numFilesToCommit"].(float64); ok {
		numFilesToCommit = int(value)
	} else if configValue := config.Get("numFilesToCommit"); configValue != "" {
		numFilesToCommit = int(configValue.(float64))
	}

	utils.Debug("Number of files to prepare commit messages for: " + strconv.Itoa(numFilesToCommit))

	changedFiles, err := git.GetAllChangedFiles()
	if err != nil {
		utils.Error("Failed to get changed files: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(changedFiles))

	for i, file := range changedFiles {
		if i >= numFilesToCommit {
			break
		}

		wg.Add(1)
		go func(file string) {
			defer wg.Done()

			message, err := git.GenCommitMessage(file)
			if err != nil {
				utils.Error("Failed to generate commit message: " + err.Error())
				errChan <- err
				return
			}

			utils.Debug("Generated commit message for file: " + file + " - " + message)
			output.Set(file, message)
		}(file)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		http.Error(w, "One or more errors occurred while preparing commit messages", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output.GetAll())
}

func CommitPreparedFilesHandler(w http.ResponseWriter, r *http.Request) {
	commitMessages := output.GetAll()
	if len(commitMessages) == 0 {
		http.Error(w, "No prepared commit messages found in output.json", http.StatusBadRequest)
		return
	}

	for file, message := range commitMessages {
		utils.Debug("Adding file to commit: " + file)
		if _, err := git.RunGitCmd("add", file); err != nil {
			utils.Error("Failed to add file to commit: " + err.Error())
			http.Error(w, "Failed to add file to commit: "+err.Error(), http.StatusInternalServerError)
			return
		}

		utils.Debug("Committing file: " + file + " with message: " + message)
		if _, err := git.RunGitCmd("commit", "-m", message); err != nil {
			utils.Error("Failed to commit file: " + err.Error())
			http.Error(w, "Failed to commit file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		output.Delete(file)
	}

	output.Clear()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Files committed successfully and output.json deleted"))
}
