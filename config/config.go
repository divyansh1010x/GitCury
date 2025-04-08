package config

import (
	"GitCury/utils"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	settings = make(map[string]interface{})
	mu       sync.RWMutex
)

func init() {
	LoadConfig()
}

func LoadConfig() {
	configFilePath := os.Getenv("HOME") + "/.gitcury/config.json"
	file, err := os.Open(configFilePath)
	if os.IsNotExist(err) {
		// Set default settings if the file does not exist
		settings = map[string]interface{}{
			"app_name":         "GitCury",
			"version":          "1.0.0",
			"root_folders":     []string{"."},
			"config_dir":       os.Getenv("HOME") + "/.gitcury",
			"output_file_path": os.Getenv("HOME") + "/.gitcury/output.json",
			"editor":           "nano",
		}

		// Save the default settings to the file
		if err := saveConfigToFile(configFilePath); err != nil {
			utils.Error("[NEXUS]: ⚠️ Failed to save default configuration: " + err.Error())
		} else {
			utils.Debug("[NEXUS]: ⚙️ Config file not found. Using default settings and saving to file: \n" + utils.ToJSON(settings))
		}
		return
	} else if err != nil {
		utils.Error("[NEXUS]: 🚨 Error opening configuration file: " + err.Error())
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&settings); err != nil {
		utils.Error("[NEXUS]: 🚨 Error decoding configuration file: " + err.Error())
		panic(err)
	}

	// Ensure necessary configurations are set to default if missing
	defaults := map[string]interface{}{
		"app_name":         "GitCury",
		"version":          "1.0.0",
		"config_dir":       os.Getenv("HOME") + "/.gitcury",
		"output_file_path": os.Getenv("HOME") + "/.gitcury/output.json",
		"editor":           "nano",
	}
	for key, value := range defaults {
		if _, exists := settings[key]; !exists {
			settings[key] = value
			utils.Warning("[NEXUS]: ⚠️ Missing configuration '" + key + "'. Setting to default: " + fmt.Sprintf("%v", value))
		}
	}

	// Check for important configurations and warn if empty
	if rootFolders, ok := settings["root_folders"].([]interface{}); !ok || len(rootFolders) == 0 {
		utils.Warning("[NEXUS]: ⚠️ 'root_folders' is empty or missing. Please configure it to ensure proper functionality.")
	}
	if apiKey, ok := settings["GEMINI_API_KEY"].(string); !ok || apiKey == "" {
		utils.Warning("[NEXUS]: ⚠️ 'GEMINI_API_KEY' is missing. Please set it to enable AI-powered commit messages.")
	}

	// Set log level if available
	if level, ok := settings["log_level"].(string); ok {
		utils.SetLogLevel(level)
		utils.Debug("[NEXUS]: 🔧 Log level set to: " + level)
	}

	saveConfigToFile(configFilePath)
	utils.Debug("[NEXUS]: 🔧 Loaded configuration successfully: \n" + utils.ToJSON(settings))
}

func saveConfigToFile(configFilePath string) error {
	// Ensure the directory exists
	dir := filepath.Dir(configFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(settings)
}

func Set(key string, value any) {
	mu.Lock()
	defer mu.Unlock()
	settings[key] = value

	utils.Debug("[NEXUS]: 💾 Saving updated configuration...")

	configFilePath := os.Getenv("HOME") + "/.gitcury/config.json"

	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		utils.Error("[NEXUS]: ⚠️ Error saving configuration: " + err.Error())
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(settings); err != nil {
		utils.Error("[NEXUS]: ⚠️ Error saving configuration: " + err.Error())
	}
}

func Get(key string) interface{} {
	mu.RLock()
	defer mu.RUnlock()
	utils.Debug("[NEXUS]: 🔍 Retrieving configuration key: " + key + " with value: " + utils.ToJSON(settings[key]))
	return settings[key]
}

func GetAll() map[string]interface{} {
	mu.RLock()
	defer mu.RUnlock()
	copy := make(map[string]interface{})
	for key, value := range settings {
		copy[key] = value
	}
	return copy
}

func Remove(key string) {
	mu.Lock()
	defer mu.Unlock()
	delete(settings, key)

	go func() {
		mu.RLock()
		defer mu.RUnlock()

		configFilePath := os.Getenv("HOME") + "/.gitcury/config.json"

		file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			utils.Error("[NEXUS]: ⚠️ Error saving configuration after removal: " + err.Error())
			return
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		if err := encoder.Encode(settings); err != nil {
			utils.Error("[NEXUS]: ⚠️ Error saving configuration after removal: " + err.Error())
		}
	}()
}

func Delete() {
	utils.Debug("[NEXUS]: 🗑️ Deleting configuration file...")
	mu.Lock()
	defer mu.Unlock()
	configFilePath := os.Getenv("HOME") + "/.gitcury/config.json"
	if err := os.Remove(configFilePath); err != nil {
		utils.Error("[NEXUS]: ⚠️ Error deleting configuration file: " + err.Error())
	}
}
