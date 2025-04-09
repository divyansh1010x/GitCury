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

type Alias struct {
	Commit  string `json:"commit"`
	Push    string `json:"push"`
	GetMsgs string `json:"getmsgs"`
	Output  string `json:"output"`
	Config  string `json:"config"`
	Setup   string `json:"setup"`
	Boom    string `json:"boom"`
}

var (
	Aliases        Alias
	DefaultAliases = Alias{
		Commit:  "seal",
		Push:    "deploy",
		GetMsgs: "genesis",
		Output:  "trace",
		Config:  "nexus",
		Setup:   "bootstrap",
		Boom:    "cascade",
	}
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
			"aliases": map[string]string{
				"getmsgs": DefaultAliases.GetMsgs,
				"commit":  DefaultAliases.Commit,
				"push":    DefaultAliases.Push,
				"output":  DefaultAliases.Output,
				"config":  DefaultAliases.Config,
				"setup":   DefaultAliases.Setup,
				"boom":    DefaultAliases.Boom,
			},
			"retries": 3,
			"timeout": 30,
		}

		// Save the default settings to the file
		if err := saveConfigToFile(configFilePath); err != nil {
			utils.Error("[" + Aliases.Config + "]: ⚠️ Failed to save default configuration: " + err.Error())
		} else {
			utils.Debug("[" + Aliases.Config + "]: ⚙️ Config file not found. Using default settings and saving to file: \n" + utils.ToJSON(settings))
		}
		return
	} else if err != nil {
		utils.Error("[" + Aliases.Config + "]: 🚨 Error opening configuration file: " + err.Error())
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&settings); err != nil {
		utils.Error("[Config]: 🚨 Error decoding configuration file: " + err.Error())
		panic(err)
	}

	// Ensure necessary configurations are set to default if missing
	defaults := map[string]interface{}{
		"app_name":         "GitCury",
		"version":          "1.0.0",
		"config_dir":       os.Getenv("HOME") + "/.gitcury",
		"output_file_path": os.Getenv("HOME") + "/.gitcury/output.json",
		"editor":           "nano",
		"aliases": map[string]string{
			"getmsgs": DefaultAliases.GetMsgs,
			"commit":  DefaultAliases.Commit,
			"push":    DefaultAliases.Push,
			"output":  DefaultAliases.Output,
			"config":  DefaultAliases.Config,
			"setup":   DefaultAliases.Setup,
			"boom":    DefaultAliases.Boom,
		},
		"retries":      3,
		"timeout":      30,
		"root_folders": []string{"."},
	}
	for key, value := range defaults {
		if _, exists := settings[key]; !exists {
			settings[key] = value
			utils.Warning("[" + Aliases.Config + "]: ⚠️ Missing configuration '" + key + "'. Setting to default: " + fmt.Sprintf("%v", value))
		}
	}

	// utils.Debug("[Config]: 🔧 Loaded configuration successfully: \n" + utils.ToJSON(settings))
	// Set aliases
	if aliasMap, ok := settings["aliases"].(map[string]interface{}); ok {
		Aliases = Alias{
			Commit:  fmt.Sprintf("%v", aliasMap["commit"]),
			Push:    fmt.Sprintf("%v", aliasMap["push"]),
			GetMsgs: fmt.Sprintf("%v", aliasMap["getmsgs"]),
			Output:  fmt.Sprintf("%v", aliasMap["output"]),
			Config:  fmt.Sprintf("%v", aliasMap["config"]),
			Setup:   fmt.Sprintf("%v", aliasMap["setup"]),
			Boom:    fmt.Sprintf("%v", aliasMap["boom"]),
		}
	}

	// utils.Debug("[Config]: 🔧 Aliases loaded successfully: \n" + utils.ToJSON(Aliases))
	// Check for important configurations and warn if empty
	if rootFolders, ok := settings["root_folders"].([]interface{}); !ok || len(rootFolders) == 0 {
		utils.Warning("[" + Aliases.Config + "]: ⚠️ 'root_folders' is empty or missing. Please configure it to ensure proper functionality.")
	}
	if apiKey, ok := settings["GEMINI_API_KEY"].(string); !ok || apiKey == "" {
		utils.Warning("[" + Aliases.Config + "]: ⚠️ 'GEMINI_API_KEY' is missing. Please set it to enable AI-powered commit messages.")
	}

	// Set log level if available
	if level, ok := settings["log_level"].(string); ok {
		utils.SetLogLevel(level)
		utils.Debug("[" + Aliases.Config + "]: 🔧 Log level set to: " + level)
	}

	retries, ok1 := settings["retries"].(float64) // JSON unmarshals numbers as float64
	timeout, ok2 := settings["timeout"].(float64)
	utils.Debug("[" + Aliases.Config + "]: 🔧 Retries: " + fmt.Sprintf("%v", retries) + ", Timeout: " + fmt.Sprintf("%v", timeout))
	if !ok1 || retries <= 0 || !ok2 || timeout <= 0 {
		utils.Warning("[" + Aliases.Config + "]: ⚠️ 'retries' or 'timeout' is missing or invalid. Setting to default values.")
		settings["retries"] = 3
		settings["timeout"] = 30
		retries = 3
		timeout = 30
	}
	utils.SetTimeoutVar(int(retries), int(timeout)) // Convert float64 to int

	saveConfigToFile(configFilePath)
	utils.Debug("[" + Aliases.Config + "]: 🔧 Loaded configuration successfully: \n" + utils.ToJSON(settings))
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

	utils.Debug("[" + Aliases.Config + "]: 💾 Saving updated configuration...")

	configFilePath := os.Getenv("HOME") + "/.gitcury/ json"

	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		utils.Error("[" + Aliases.Config + "]: ⚠️ Error saving configuration: " + err.Error())
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(settings); err != nil {
		utils.Error("[" + Aliases.Config + "]: ⚠️ Error saving configuration: " + err.Error())
	}
}

func Get(key string) interface{} {
	mu.RLock()
	defer mu.RUnlock()
	utils.Debug("[" + Aliases.Config + "]: 🔍 Retrieving configuration key: " + key + " with value: " + utils.ToJSON(settings[key]))
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

		configFilePath := os.Getenv("HOME") + "/.gitcury/ json"

		file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			utils.Error("[" + Aliases.Config + "]: ⚠️ Error saving configuration after removal: " + err.Error())
			return
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		if err := encoder.Encode(settings); err != nil {
			utils.Error("[" + Aliases.Config + "]: ⚠️ Error saving configuration after removal: " + err.Error())
		}
	}()
}

func Delete() {
	utils.Debug("[" + Aliases.Config + "]: 🗑️ Deleting configuration file...")
	mu.Lock()
	defer mu.Unlock()
	configFilePath := os.Getenv("HOME") + "/.gitcury/ json"
	if err := os.Remove(configFilePath); err != nil {
		utils.Error("[" + Aliases.Config + "]: ⚠️ Error deleting configuration file: " + err.Error())
	}
}
