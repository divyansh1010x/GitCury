package config

import (
	"GitCury/api"
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
	err := LoadConfig()
	if err != nil {
		// Check if it's a critical config error
		if configErr, ok := err.(*utils.StructuredError); ok && configErr.Type == utils.ConfigError {
			if stop, exists := configErr.Context["stop_execution"].(bool); exists && stop {
				// Exit immediately for critical missing configs
				utils.Error("")
				utils.Error("💥 GitCury cannot start due to missing critical configuration.")
				utils.Error("   Please fix the configuration issues shown above and try again.")
				utils.Error("")
				os.Exit(1)
			}
		}
		// For non-critical errors, just log them
		utils.Warning("[Config]: " + err.Error())
	}
}

func LoadConfig() error {
	configFilePath := os.Getenv("HOME") + "/.gitcury/config.json"
	file, err := os.Open(configFilePath)
	if os.IsNotExist(err) {
		// Prompt user about creating default config
		utils.Info("🔧 Configuration file not found at " + configFilePath)
		utils.Info("📝 Creating default configuration file with recommended settings...")

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
			"retries":       3,
			"timeout":       30,
			"maxConcurrent": 5,
			"logLevel":      "info",
		}

		// Save the default settings to the file
		if err := saveConfigToFile(configFilePath); err != nil {
			utils.Error("[Config]: Failed to save default configuration: " + err.Error())
			return utils.NewConfigError(
				"Failed to save default configuration",
				err,
				map[string]interface{}{
					"configPath": configFilePath,
				},
			)
		}

		utils.Success("✅ Default configuration file created at " + configFilePath)
		utils.Warning("⚠️  IMPORTANT: You need to set your GEMINI_API_KEY to use AI features!")
		utils.Info("🔑 To set your API key, run this command:")
		utils.Print("    gitcury config set --key GEMINI_API_KEY --value YOUR_API_KEY_HERE")
		utils.Info("📖 Get your API key from: https://aistudio.google.com/app/apikey")
		utils.Info("💡 You can also set the environment variable: export GEMINI_API_KEY=your_key_here")

		utils.Debug("[Config]: Config file not found. Using default settings and saving to file: \n" + utils.ToJSON(settings))

		// Set up aliases with default values
		Aliases = DefaultAliases
		return nil
	} else if err != nil {
		utils.Error("[Config]: Error opening configuration file: " + err.Error())
		return utils.NewConfigError(
			"Error opening configuration file",
			err,
			map[string]interface{}{
				"configPath": configFilePath,
			},
		)
	}
	defer file.Close()

	// Parse the config file
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&settings); err != nil {
		utils.Error("[Config]: Error parsing configuration file: " + err.Error())
		return utils.NewConfigError(
			"Error parsing configuration file",
			err,
			map[string]interface{}{
				"configPath": configFilePath,
			},
		)
	}

	// Validate required settings
	if err := validateConfig(); err != nil {
		return err
	}

	// Check critical configuration and stop if missing
	criticalMissing := checkCriticalConfig()
	if len(criticalMissing) > 0 {
		utils.Error("[Config]: Critical configuration missing: " + fmt.Sprint(criticalMissing))
		return utils.NewConfigError(
			"Critical configuration missing",
			nil,
			map[string]interface{}{
				"missing_fields": criticalMissing,
				"stop_execution": true,
			},
		)
	}

	// Set up aliases
	aliasesMap, ok := settings["aliases"].(map[string]interface{})
	if !ok {
		utils.Warning("[Config]: No aliases configuration found. Using defaults.")
		Aliases = DefaultAliases
	} else {
		// Convert the map to our Alias struct
		commitAlias, _ := aliasesMap["commit"].(string)
		if commitAlias == "" {
			commitAlias = DefaultAliases.Commit
		}

		pushAlias, _ := aliasesMap["push"].(string)
		if pushAlias == "" {
			pushAlias = DefaultAliases.Push
		}

		getMsgsAlias, _ := aliasesMap["getmsgs"].(string)
		if getMsgsAlias == "" {
			getMsgsAlias = DefaultAliases.GetMsgs
		}

		outputAlias, _ := aliasesMap["output"].(string)
		if outputAlias == "" {
			outputAlias = DefaultAliases.Output
		}

		configAlias, _ := aliasesMap["config"].(string)
		if configAlias == "" {
			configAlias = DefaultAliases.Config
		}

		setupAlias, _ := aliasesMap["setup"].(string)
		if setupAlias == "" {
			setupAlias = DefaultAliases.Setup
		}

		boomAlias, _ := aliasesMap["boom"].(string)
		if boomAlias == "" {
			boomAlias = DefaultAliases.Boom
		}

		Aliases = Alias{
			Commit:  commitAlias,
			Push:    pushAlias,
			GetMsgs: getMsgsAlias,
			Output:  outputAlias,
			Config:  configAlias,
			Setup:   setupAlias,
			Boom:    boomAlias,
		}
	}

	// Set log level if available
	logLevel, ok := settings["logLevel"].(string)
	if ok && logLevel != "" {
		utils.SetLogLevel(logLevel)
	}

	// Initialize API configuration
	api.LoadConfig(settings)

	utils.Debug("[Config]: Configuration loaded successfully: \n" + utils.ToJSON(settings))
	utils.Debug("[Config]: Aliases loaded successfully: \n" + utils.ToJSON(Aliases))

	return nil
}

// checkCriticalConfig checks for critical configuration values and prompts user with commands if missing
func checkCriticalConfig() []string {
	var criticalMissing []string

	// Check for GEMINI_API_KEY - this is critical for main functionality
	geminiKey, exists := settings["GEMINI_API_KEY"]
	if !exists || geminiKey == "" {
		// Try to get from environment
		envKey := os.Getenv("GEMINI_API_KEY")
		if envKey == "" {
			criticalMissing = append(criticalMissing, "GEMINI_API_KEY")
			utils.Error("")
			utils.Error("🚫 CRITICAL: GEMINI_API_KEY is required but not configured!")
			utils.Error("")
			utils.Error("💡 To fix this issue, run one of these commands:")
			utils.Error("   📝 Set via config file:")
			utils.Error("      ./gitcury config set --key GEMINI_API_KEY --value YOUR_API_KEY_HERE")
			utils.Error("")
			utils.Error("   🌍 Set via environment variable:")
			utils.Error("      export GEMINI_API_KEY=YOUR_API_KEY_HERE")
			utils.Error("")
			utils.Error("📖 Get your free API key from:")
			utils.Error("   🔗 https://aistudio.google.com/app/apikey")
			utils.Error("")
			utils.Error("⚠️  GitCury cannot function without this API key.")
		} else {
			utils.Debug("[Config]: Using GEMINI_API_KEY from environment variables")
			settings["GEMINI_API_KEY"] = envKey
		}
	}

	// Check for root_folders - critical for knowing where to work
	rootFolders, exists := settings["root_folders"]
	if !exists {
		criticalMissing = append(criticalMissing, "root_folders")
		utils.Error("")
		utils.Error("🚫 CRITICAL: root_folders configuration is missing!")
		utils.Error("")
		utils.Error("💡 To fix this, add your project directories:")
		utils.Error("   📝 Add current directory:")
		utils.Error("      ./gitcury config set --key root_folders --value '[\".\"]'")
		utils.Error("")
		utils.Error("   📁 Add specific directories:")
		utils.Error("      ./gitcury config set --key root_folders --value '[\"/path/to/project1\",\"/path/to/project2\"]'")
		utils.Error("")
	} else {
		// Check if root_folders is empty or invalid
		if folders, ok := rootFolders.([]interface{}); ok {
			if len(folders) == 0 {
				criticalMissing = append(criticalMissing, "root_folders")
				utils.Error("")
				utils.Error("🚫 CRITICAL: root_folders is empty!")
				utils.Error("")
				utils.Error("💡 Add at least one project directory:")
				utils.Error("   ./gitcury config set --key root_folders --value '[\".\"]'")
				utils.Error("")
			}
		} else {
			criticalMissing = append(criticalMissing, "root_folders")
			utils.Error("")
			utils.Error("🚫 CRITICAL: root_folders has invalid format!")
			utils.Error("")
			utils.Error("💡 Fix the format:")
			utils.Error("   ./gitcury config set --key root_folders --value '[\".\"]'")
			utils.Error("")
		}
	}

	// Check for numFilesToCommit - not critical but important for functionality
	if _, exists := settings["numFilesToCommit"]; !exists {
		utils.Warning("")
		utils.Warning("⚠️  RECOMMENDED: numFilesToCommit not set, using default (5)")
		utils.Warning("")
		utils.Warning("💡 To set a custom value:")
		utils.Warning("   ./gitcury config set --key numFilesToCommit --value 10")
		utils.Warning("")
		// Set a default value
		settings["numFilesToCommit"] = 5
	}

	return criticalMissing
}

// validateConfig ensures that required configuration values are present and valid
func validateConfig() error {
	// Check for required configuration fields
	requiredFields := []string{"app_name", "version", "root_folders"}
	missingFields := []string{}

	for _, field := range requiredFields {
		if _, exists := settings[field]; !exists {
			missingFields = append(missingFields, field)
		}
	}

	if len(missingFields) > 0 {
		utils.Warning("[Config]: Missing required configuration fields: " + fmt.Sprint(missingFields))
		// Add default values for missing fields
		if _, exists := settings["app_name"]; !exists {
			settings["app_name"] = "GitCury"
		}
		if _, exists := settings["version"]; !exists {
			settings["version"] = "1.0.0"
		}
		if _, exists := settings["root_folders"]; !exists {
			settings["root_folders"] = []string{"."}
		}
	}

	// Validate root_folders
	rootFolders, ok := settings["root_folders"].([]interface{})
	if !ok {
		utils.Warning("[Config]: Invalid root_folders configuration. Must be an array.")
		settings["root_folders"] = []string{"."}
	} else if len(rootFolders) == 0 {
		utils.Warning("[Config]: Empty root_folders configuration. Adding current directory.")
		settings["root_folders"] = []string{"."}
	}

	// Ensure output_file_path is set
	if _, exists := settings["output_file_path"]; !exists {
		settings["output_file_path"] = os.Getenv("HOME") + "/.gitcury/output.json"
	}

	// Ensure config_dir is set
	if _, exists := settings["config_dir"]; !exists {
		settings["config_dir"] = os.Getenv("HOME") + "/.gitcury"
	}

	// Check for API key - this is critical for most functionality
	if geminiKey, exists := settings["GEMINI_API_KEY"]; !exists || geminiKey == "" {
		// Try to get from environment
		envKey := os.Getenv("GEMINI_API_KEY")
		if envKey != "" {
			utils.Debug("[Config]: Using GEMINI_API_KEY from environment variables")
			settings["GEMINI_API_KEY"] = envKey
		} else {
			utils.Warning("[Config]: GEMINI_API_KEY not found in config or environment. Some features may not work correctly.")
		}
	}

	return nil
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

	configFilePath := os.Getenv("HOME") + "/.gitcury/config.json"

	dir := os.Getenv("HOME") + "/.gitcury"
	if err := os.MkdirAll(dir, 0755); err != nil {
		utils.Error("[" + Aliases.Config + "]: ⚠️ Error creating config directory: " + err.Error())
		return
	}

	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		utils.Error("[" + Aliases.Config + "]: ⚠️ Error saving configuration: " + err.Error())
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(settings); err != nil {
		utils.Error("[" + Aliases.Config + "]: ⚠️ Error encoding configuration: " + err.Error())
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
