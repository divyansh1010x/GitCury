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

	numFilesToCommit := 10 // Default value
	if configValue := config.Get("numFilesToCommit"); configValue != "" {
		if configValueFloat, ok := configValue.(float64); ok {
			numFilesToCommit = int(configValueFloat)
		}
	}

	utils.Debug("Number of files to prepare commit messages for: " + strconv.Itoa(numFilesToCommit))

	rootFolders, ok := config.Get("root_folders").([]interface{})
	if !ok {
		utils.Error("Invalid or missing root_folders configuration")
		http.Error(w, "Invalid or missing root_folders configuration", http.StatusInternalServerError)
		return
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

			if len(changedFiles) > numFilesToCommit {
				changedFiles = changedFiles[:numFilesToCommit]
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
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Errors occurred during batch processing",
			"errors":  errors,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output.GetAll())
}

func CommitAllFiles(w http.ResponseWriter, r *http.Request) {
	rootFolders := output.GetAll().Folders

	var rootFolderWg sync.WaitGroup
	var mu sync.Mutex
	var errors []string

	for _, rootFolder := range rootFolders {
		rootFolderWg.Add(1)

		go func(rootFolder output.Folder) {
			defer rootFolderWg.Done()
			utils.Debug("Root folder to commit in: " + rootFolder.Name)

			err := git.CommitBatch(rootFolder)
			if err != nil {
				utils.Error("Failed to commit batch: " + err.Error())
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, Error: %s", rootFolder.Name, err.Error()))
				mu.Unlock()
				return
			}
		}(rootFolder)
	}

	rootFolderWg.Wait()

	if len(errors) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Errors occurred during batch processing",
			"errors":  errors,
		})
		return
	}

	output.Clear()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Files committed successfully and output.json deleted"))
}

func CommitFolder(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		RootFolderName string `json:"rootFolder"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil || requestBody.RootFolderName == "" {
		utils.Error("Invalid or missing root folder name in request body")
		http.Error(w, "Invalid or missing root folder name in request body", http.StatusBadRequest)
		return
	}

	rootFolderName := requestBody.RootFolderName

	rootFolder := output.GetAllFiles(rootFolderName)
	if len(rootFolder.Files) == 0 {
		utils.Error("Root folder not found: " + rootFolderName)
		http.Error(w, "Root folder not found", http.StatusNotFound)
		return
	}

	err := git.CommitBatch(rootFolder)
	if err != nil {
		utils.Error("Failed to commit batch: " + err.Error())
		http.Error(w, "Failed to commit batch", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Files committed successfully"))
}
