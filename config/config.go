package config

import (
	"GitCury/utils"
	"encoding/json"
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
			"output_file_path": os.Getenv("HOME") + "/.gitcury/output.json",
			"editor":           "nano",
		}

		// Save the default settings to the file
		if err := saveConfigToFile(configFilePath); err != nil {
			utils.Error("Error saving default config: " + err.Error())
		} else {
			utils.Error("Config file not found. Using default settings and saving to file: \n" + utils.ToJSON(settings))
		}
		return
	} else if err != nil {
		utils.Error("Error opening config file: " + err.Error())
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&settings); err != nil {
		utils.Error("Error decoding config file: " + err.Error())
		panic(err)
	}

	if level, ok := settings["log_level"].(string); ok {
		utils.SetLogLevel(level)
	}
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

	utils.Debug("Saving config file...")

	configFilePath := os.Getenv("HOME") + "/.gitcury/config.json"

	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		utils.Error("Error saving config: " + err.Error())
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(settings); err != nil {
		utils.Error("Error saving config: " + err.Error())
	}
}

func Get(key string) interface{} {
	mu.RLock()
	defer mu.RUnlock()
	utils.Debug("Getting config key: " + key + " with value: " + utils.ToJSON(settings[key]))
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
			utils.Error("Error saving config: " + err.Error())
			return
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		if err := encoder.Encode(settings); err != nil {
			utils.Error("Error saving config: " + err.Error())
		}
	}()
}

func Delete() {
	utils.Info("Deleting config file...")
	mu.Lock()
	defer mu.Unlock()
	configFilePath := os.Getenv("HOME") + "/.gitcury/config.json"
	if err := os.Remove(configFilePath); err != nil {
		utils.Error("Error deleting config file: " + err.Error())
	}
}
