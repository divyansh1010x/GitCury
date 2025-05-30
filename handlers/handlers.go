package handlers

import (
	"github.com/lakshyajain-0291/gitcury/config"
	"github.com/lakshyajain-0291/gitcury/core"
	"github.com/lakshyajain-0291/gitcury/output"
	"github.com/lakshyajain-0291/gitcury/utils"
	"encoding/json"
	"fmt"
	"net/http"
)

func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Handle GET request to return the current configuration
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(config.GetAll()); err != nil {
			utils.Error("Error encoding config response: " + err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
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
		if err := json.NewEncoder(w).Encode(config.GetAll()); err != nil {
			utils.Error("Error encoding config response: " + err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Handle unsupported methods
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func PrepareCommitMessagesHandler(w http.ResponseWriter, r *http.Request) {
	err := core.GetAllMsgs()
	if err != nil {
		utils.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(output.GetAll()); err != nil {
		utils.Error("Error encoding commit messages response: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func PrepareCommitMessagesOne(w http.ResponseWriter, r *http.Request) {
	folder := r.URL.Query().Get("rootFolder")
	if folder == "" {
		utils.Error("Missing root folder name in query parameter")
		http.Error(w, "Missing root folder name in query parameter", http.StatusBadRequest)
		return
	}

	err := core.GetMsgsForRootFolder(folder)
	if err != nil {
		utils.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(output.GetFolder(folder)); err != nil {
		utils.Error("Error encoding folder response: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func CommitAllFiles(w http.ResponseWriter, r *http.Request) {

	err := core.CommitAllRoots()
	if err != nil {
		utils.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Files committed successfully and output.json deleted")); err != nil {
		utils.Error("Error writing response: " + err.Error())
	}
}

func CommitFolder(w http.ResponseWriter, r *http.Request) {
	rootFolderName := r.URL.Query().Get("rootFolder")
	if rootFolderName == "" {
		utils.Error("Missing root folder name in query parameter")
		http.Error(w, "Missing root folder name in query parameter", http.StatusBadRequest)
		return
	}

	err := core.CommitOneRoot(rootFolderName)
	if err != nil {
		utils.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Files committed successfully")); err != nil {
		utils.Error("Error writing response: " + err.Error())
	}
}

func PushAll(w http.ResponseWriter, r *http.Request) {
	branchName := r.URL.Query().Get("branch")
	if branchName == "" {
		utils.Error("Missing branch name in query parameter")
		http.Error(w, "Missing branch name in query parameter", http.StatusBadRequest)
		return
	}

	err := core.PushAllRoots(branchName)
	if err != nil {
		utils.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("All folders pushed successfully")); err != nil {
		utils.Error("Error writing response: " + err.Error())
	}
}

func PushOne(w http.ResponseWriter, r *http.Request) {
	rootFolderName := r.URL.Query().Get("rootFolder")
	branchName := r.URL.Query().Get("branch")

	if rootFolderName == "" || branchName == "" {
		utils.Error("Missing root folder name or branch name in query parameters")
		http.Error(w, "Missing root folder name or branch name in query parameters", http.StatusBadRequest)
		return
	}

	err := core.PushOneRoot(rootFolderName, branchName)
	if err != nil {
		utils.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Files pushed successfully")); err != nil {
		utils.Error("Error writing response: " + err.Error())
	}
}
