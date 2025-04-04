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
	if r.Method == http.MethodGet {
		// Handle GET request to return the current configuration
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config.GetAll())
		return
	}

	if r.Method == http.MethodPost {
		// Handle POST request to update the configuration
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config.GetAll())
		return
	}

	// Handle unsupported methods
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

func PrepareCommitMessagesOne(w http.ResponseWriter, r *http.Request) {
	folder := r.URL.Query().Get("rootFolder")
	if folder == "" {
		utils.Error("Missing root folder name in query parameter")
		http.Error(w, "Missing root folder name in query parameter", http.StatusBadRequest)
		return
	}

	numFilesToCommit := 10 // Default value
	if configValue := config.Get("numFilesToCommit"); configValue != "" {
		if configValueFloat, ok := configValue.(float64); ok {
			numFilesToCommit = int(configValueFloat)
		}
	}

	utils.Debug("Number of files to prepare commit messages for: " + strconv.Itoa(numFilesToCommit))

	utils.Debug("Root folder to get messages : " + folder)

	changedFiles, err := git.GetAllChangedFiles(folder)
	if err != nil {
		utils.Error("Failed to get changed files: " + err.Error())
		http.Error(w, "Failed to get changed files", http.StatusInternalServerError)
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
		utils.Error("Batch processing failed: " + err.Error())
		http.Error(w, "Batch processing failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output.GetAllFiles(folder))
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
	rootFolderName := r.URL.Query().Get("rootFolder")
	if rootFolderName == "" {
		utils.Error("Missing root folder name in query parameter")
		http.Error(w, "Missing root folder name in query parameter", http.StatusBadRequest)
		return
	}

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

func PushAll(w http.ResponseWriter, r *http.Request) {
	branchName := r.URL.Query().Get("branch")
	if branchName == "" {
		utils.Error("Missing branch name in query parameter")
		http.Error(w, "Missing branch name in query parameter", http.StatusBadRequest)
		return
	}

	rootFolders := output.GetAll().Folders
	if len(rootFolders) == 0 {
		utils.Error("No root folders found")
		http.Error(w, "No root folders found", http.StatusNotFound)
		return
	}

	var rootFolderWg sync.WaitGroup
	var mu sync.Mutex
	var errors []string

	for _, rootFolder := range rootFolders {
		rootFolderWg.Add(1)

		go func(rootFolder output.Folder) {
			defer rootFolderWg.Done()
			utils.Debug("Root folder to push: " + rootFolder.Name)

			err := git.PushBranch(rootFolder, branchName)
			if err != nil {
				utils.Error("Failed to push branch: " + err.Error())
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Folder: %s, Error: %s", rootFolder.Name, err.Error()))
				mu.Unlock()
				return
			}
		}(rootFolder)
	}

	rootFolderWg.Wait()
	if len(errors) > 0 {
		utils.Error("Errors occurred during push operation")
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Errors occurred during push operation",
			"errors":  errors,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("All folders pushed successfully"))
}

func PushOne(w http.ResponseWriter, r *http.Request) {
	rootFolderName := r.URL.Query().Get("rootFolder")
	branchName := r.URL.Query().Get("branch")

	if rootFolderName == "" || branchName == "" {
		utils.Error("Missing root folder name or branch name in query parameters")
		http.Error(w, "Missing root folder name or branch name in query parameters", http.StatusBadRequest)
		return
	}

	rootFolder := output.GetAllFiles(rootFolderName)
	if len(rootFolder.Files) == 0 {
		utils.Error("Root folder not found: " + rootFolderName)
		http.Error(w, "Root folder not found", http.StatusNotFound)
		return
	}

	err := git.PushBranch(rootFolder, branchName)
	if err != nil {
		utils.Error("Failed to push branch: " + err.Error())
		http.Error(w, "Failed to push branch", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Files pushed successfully"))
}
