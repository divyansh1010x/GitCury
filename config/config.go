package config

import (
	"encoding/json"
	"fmt"
	"github.com/lakshyajain-0291/gitcury/api"
	"github.com/lakshyajain-0291/gitcury/utils"
	"os"
	"path/filepath"
	"strings"
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
	// Skip config loading during tests to avoid initialization issues
	if isTestMode() {
		// Set minimal defaults for testing
		Aliases = DefaultAliases
		return
	}

	// Allow config commands to run even with missing critical config
	if isConfigCommand() {
		// Load config but don't exit on critical errors for config commands
		err := LoadConfigForConfigCommands()
		if err != nil {
			// For config commands, just log warnings, don't exit
			utils.Debug("[Config]: " + err.Error())
		}
		return
	}

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

// isTestMode detects if we're running in test mode
func isTestMode() bool {
	// Check for common test environment indicators
	for _, arg := range os.Args {
		if strings.Contains(arg, "test") || strings.Contains(arg, ".test") {
			return true
		}
	}
	// Check if GITCURY_TEST_MODE environment variable is set
	return os.Getenv("GITCURY_TEST_MODE") == "true"
}

// isConfigCommand detects if we're running a config command
func isConfigCommand() bool {
	// Improved detection for config commands
	if len(os.Args) < 2 {
		return false
	}

	// Check if the command is the config command itself or its alias
	for i, arg := range os.Args {
		// Direct config command or its alias
		if arg == "config" || arg == DefaultAliases.Config || arg == "nexus" {
			return true
		}

		// Check for config subcommands
		if i > 0 {
			prevArg := os.Args[i-1]
			if prevArg == "config" || prevArg == DefaultAliases.Config || prevArg == "nexus" {
				// Any subcommand of config should be considered a config command
				return true
			}
		}

		// Also check for flags that belong to config command
		if (arg == "--delete" || arg == "--reset") && i > 0 {
			prevArg := os.Args[i-1]
			if prevArg == "config" || prevArg == DefaultAliases.Config || prevArg == "nexus" {
				return true
			}
		}
	}

	return false
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
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
			"clustering": map[string]interface{}{
				"defaultMethod":                 "directory",
				"enableFallbackMethods":         true,
				"maxFilesForSemanticClustering": 10,
				"confidenceThresholds": map[string]float64{
					"directory": 0.8,
					"pattern":   0.7,
					"cached":    0.6,
					"semantic":  0.5,
				},
				"similarityThresholds": map[string]float64{
					"directory": 0.7,
					"pattern":   0.6,
					"cached":    0.5,
					"semantic":  0.4,
				},
				"methods": map[string]interface{}{
					"directory": map[string]interface{}{
						"enabled": true,
						"weight":  1.0,
					},
					"pattern": map[string]interface{}{
						"enabled": true,
						"weight":  0.8,
					},
					"cached": map[string]interface{}{
						"enabled":          true,
						"weight":           0.6,
						"minCacheHitRatio": 0.4,
						"maxCacheAge":      24, // hours
					},
					"semantic": map[string]interface{}{
						"enabled":                 true,
						"weight":                  0.4,
						"rateLimitDelay":          2000, // milliseconds
						"maxConcurrentEmbeddings": 1,
						"embeddingTimeout":        30, // seconds
					},
				},
				"performance": map[string]interface{}{
					"preferSpeed":          true,
					"maxProcessingTime":    60, // seconds
					"enableBenchmarking":   false,
					"adaptiveOptimization": true,
				},
			},
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

	// Check critical configuration - no longer stops execution for config commands
	criticalMissing := checkCriticalConfig()
	if len(criticalMissing) > 0 && !isConfigCommand() {
		// Only stop execution for non-config commands when API key is missing
		if contains(criticalMissing, "GEMINI_API_KEY") {
			utils.Error("")
			utils.Error("💥 GitCury cannot start due to missing API key.")
			utils.Error("   Please set your GEMINI_API_KEY as shown above and try again.")
			utils.Error("")
			return utils.NewConfigError(
				"Critical configuration missing",
				nil,
				map[string]interface{}{
					"missing_fields": criticalMissing,
					"stop_execution": true,
				},
			)
		}
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

// LoadConfigForConfigCommands loads config specifically for config commands - minimal warnings and no critical errors
func LoadConfigForConfigCommands() error {
	configFilePath := os.Getenv("HOME") + "/.gitcury/config.json"
	utils.Debug("[Config]: Loading config for config commands from " + configFilePath)

	// Ensure config directory exists
	configDir := filepath.Dir(configFilePath)
	if err := os.MkdirAll(configDir, 0750); err != nil {
		utils.Debug("[Config]: Error creating config directory: " + err.Error())
		// Continue with in-memory config
	}

	file, err := os.Open(configFilePath)
	if os.IsNotExist(err) {
		// Create default config without warnings for config commands
		utils.Debug("[Config]: Config file not found, creating basic defaults")
		settings = map[string]interface{}{
			"app_name":         "GitCury",
			"version":          "1.0.0",
			"root_folders":     []string{"."},
			"config_dir":       os.Getenv("HOME") + "/.gitcury",
			"output_file_path": os.Getenv("HOME") + "/.gitcury/output.json",
			"editor":           "nano",
			// "RATE_LIMIT":   15, // Set a default rate limit
			"aliases": map[string]interface{}{
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
			"clustering": map[string]interface{}{
				"defaultMethod":                 "directory",
				"enableFallbackMethods":         true,
				"maxFilesForSemanticClustering": 10,
				"confidenceThresholds": map[string]interface{}{
					"directory": 0.8,
					"pattern":   0.7,
					"cached":    0.6,
					"semantic":  0.5,
				},
				"similarityThresholds": map[string]interface{}{
					"directory": 0.7,
					"pattern":   0.6,
					"cached":    0.5,
					"semantic":  0.4,
				},
				"methods": map[string]interface{}{
					"directory": map[string]interface{}{
						"enabled": true,
						"weight":  1.0,
					},
					"pattern": map[string]interface{}{
						"enabled": true,
						"weight":  0.8,
					},
					"cached": map[string]interface{}{
						"enabled":          true,
						"weight":           0.6,
						"minCacheHitRatio": 0.4,
						"maxCacheAge":      24, // hours
					},
					"semantic": map[string]interface{}{
						"enabled":                 true,
						"weight":                  0.4,
						"rateLimitDelay":          2000, // milliseconds
						"maxConcurrentEmbeddings": 1,
						"embeddingTimeout":        30, // seconds
					},
				},
				"performance": map[string]interface{}{
					"preferSpeed":          true,
					"maxProcessingTime":    60, // seconds
					"enableBenchmarking":   false,
					"adaptiveOptimization": true,
				},
			},
		}

		// Save the default settings silently
		if err := saveConfigToFile(configFilePath); err != nil {
			utils.Debug("[Config]: Could not save default configuration: " + err.Error())
			// Continue with in-memory config
		} else {
			utils.Debug("[Config]: Created default config for config commands")
		}

		// Set up aliases with default values
		Aliases = DefaultAliases
		return nil
	} else if err != nil {
		utils.Debug("[Config]: Could not open config file: " + err.Error() + " - using defaults")
		// Set minimal defaults even if file read fails
		settings = map[string]interface{}{
			"app_name":     "GitCury",
			"version":      "1.0.0",
			"root_folders": []string{"."},
			"editor":       "nano",
			"logLevel":     "info",
			"retries":      3,
			"timeout":      30,
			// "RATE_LIMIT":   15, // Set a default rate limit
		}
		Aliases = DefaultAliases
		return nil
	}
	defer file.Close()

	// Parse the config file
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&settings); err != nil {
		utils.Debug("[Config]: Could not parse config file: " + err.Error() + " - using defaults")
		// Set minimal defaults even if parsing fails
		settings = map[string]interface{}{
			"app_name":     "GitCury",
			"version":      "1.0.0",
			"root_folders": []string{"."},
			"editor":       "nano",
			"logLevel":     "info",
		}
		Aliases = DefaultAliases
		return nil
	}

	// Minimal validation - just ensure basic fields exist
	if _, exists := settings["app_name"]; !exists {
		settings["app_name"] = "GitCury"
	}
	if _, exists := settings["version"]; !exists {
		settings["version"] = "1.0.0"
	}
	if _, exists := settings["root_folders"]; !exists {
		settings["root_folders"] = []string{"."}
	}

	// Set up aliases quietly
	aliasesMap, ok := settings["aliases"].(map[string]interface{})
	if !ok {
		Aliases = DefaultAliases
	} else {
		// Convert the map to our Alias struct with defaults
		Aliases = Alias{
			Commit:  getStringOrDefault(aliasesMap, "commit", DefaultAliases.Commit),
			Push:    getStringOrDefault(aliasesMap, "push", DefaultAliases.Push),
			GetMsgs: getStringOrDefault(aliasesMap, "getmsgs", DefaultAliases.GetMsgs),
			Output:  getStringOrDefault(aliasesMap, "output", DefaultAliases.Output),
			Config:  getStringOrDefault(aliasesMap, "config", DefaultAliases.Config),
			Setup:   getStringOrDefault(aliasesMap, "setup", DefaultAliases.Setup),
			Boom:    getStringOrDefault(aliasesMap, "boom", DefaultAliases.Boom),
		}
	}

	// Set log level if available, but don't complain if not
	if logLevel, ok := settings["logLevel"].(string); ok && logLevel != "" {
		utils.SetLogLevel(logLevel)
	}

	utils.Debug("[Config]: Configuration loaded for config commands")
	return nil
}

// getStringOrDefault is a helper function to safely get string values from a map
func getStringOrDefault(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key].(string); ok && val != "" {
		return val
	}
	return defaultValue
}

// checkCriticalConfig checks for critical configuration values and provides helpful guidance
func checkCriticalConfig() []string {
	var criticalMissing []string
	var hasApiKey bool
	var configChanged bool

	// Check for GEMINI_API_KEY - this is critical for main functionality
	geminiKey, exists := settings["GEMINI_API_KEY"]
	if !exists || geminiKey == "" {
		// Try to get from environment
		envKey := os.Getenv("GEMINI_API_KEY")
		if envKey == "" {
			criticalMissing = append(criticalMissing, "GEMINI_API_KEY")
			hasApiKey = false
		} else {
			utils.Debug("[Config]: Using GEMINI_API_KEY from environment variables")
			settings["GEMINI_API_KEY"] = envKey
			hasApiKey = true
			configChanged = true
		}
	} else {
		hasApiKey = true
	}

	// Check for root_folders - auto-set reasonable defaults
	rootFolders, exists := settings["root_folders"]
	if !exists {
		utils.Debug("[Config]: Setting default root_folders to current directory")
		settings["root_folders"] = []string{"."}
		configChanged = true
	} else {
		// Check if root_folders is empty or invalid
		if folders, ok := rootFolders.([]interface{}); ok {
			if len(folders) == 0 {
				utils.Debug("[Config]: Empty root_folders detected, setting to current directory")
				settings["root_folders"] = []string{"."}
				configChanged = true
			}
		} else {
			utils.Debug("[Config]: Invalid root_folders format detected, setting to current directory")
			settings["root_folders"] = []string{"."}
			configChanged = true
		}
	}

	// Auto-set other important defaults
	if _, exists := settings["numFilesToCommit"]; !exists {
		utils.Debug("[Config]: Setting default numFilesToCommit to 5")
		settings["numFilesToCommit"] = 5
		configChanged = true
	}

	// if _, exists := settings["RATE_LIMIT"]; !exists {
    //     utils.Debug("[Config]: Setting default RATE_LIMIT to 15")
    //     settings["RATE_LIMIT"] = 15
    //     configChanged = true
    // }

	if _, exists := settings["editor"]; !exists {
		utils.Debug("[Config]: Setting default editor to nano")
		settings["editor"] = "nano"
		configChanged = true
	}

	if _, exists := settings["retries"]; !exists {
		utils.Debug("[Config]: Setting default retries to 3")
		settings["retries"] = 3
		configChanged = true
	}

	if _, exists := settings["timeout"]; !exists {
		utils.Debug("[Config]: Setting default timeout to 30")
		settings["timeout"] = 30
		configChanged = true
	}

	if _, exists := settings["logLevel"]; !exists {
		utils.Debug("[Config]: Setting default logLevel to info")
		settings["logLevel"] = "info"
		configChanged = true
	}

	// Auto-set clustering configuration defaults
	if _, exists := settings["clustering"]; !exists {
		utils.Debug("[Config]: Setting default clustering configuration")
		settings["clustering"] = map[string]interface{}{
			"defaultMethod":                 "directory",
			"enableFallbackMethods":         true,
			"maxFilesForSemanticClustering": 10,
			"confidenceThresholds": map[string]interface{}{
				"directory": 0.8,
				"pattern":   0.7,
				"cached":    0.6,
				"semantic":  0.5,
			},
			"similarityThresholds": map[string]interface{}{
				"directory": 0.7,
				"pattern":   0.6,
				"cached":    0.5,
				"semantic":  0.4,
			},
			"methods": map[string]interface{}{
				"directory": map[string]interface{}{
					"enabled": true,
					"weight":  1.0,
				},
				"pattern": map[string]interface{}{
					"enabled": true,
					"weight":  0.8,
				},
				"cached": map[string]interface{}{
					"enabled":          true,
					"weight":           0.6,
					"minCacheHitRatio": 0.4,
					"maxCacheAge":      24, // hours
				},
				"semantic": map[string]interface{}{
					"enabled":                 true,
					"weight":                  0.4,
					"rateLimitDelay":          2000, // milliseconds
					"maxConcurrentEmbeddings": 1,
					"embeddingTimeout":        30, // seconds
				},
			},
			"performance": map[string]interface{}{
				"preferSpeed":          true,
				"maxProcessingTime":    60, // seconds
				"enableBenchmarking":   false,
				"adaptiveOptimization": true,
			},
		}
		configChanged = true
	}

	// Save config if defaults were auto-set
	if configChanged {
		configFilePath := os.Getenv("HOME") + "/.gitcury/config.json"
		if err := saveConfigToFile(configFilePath); err != nil {
			utils.Debug("[Config]: Warning: Failed to save auto-set defaults: " + err.Error())
		} else {
			utils.Debug("[Config]: Auto-set defaults saved to config file")
		}
	}

	// Only show API key guidance if missing and if this is not a config command
	if !hasApiKey && !isConfigCommand() {
		utils.Info("")
		utils.Info("🔑 API Key Setup Required")
		utils.Info("   GitCury needs a Gemini API key for AI-powered features.")
		utils.Info("")
		utils.Info("📝 Quick setup:")
		utils.Info("   gitcury config set --key GEMINI_API_KEY --value YOUR_API_KEY_HERE")
		utils.Info("")
		utils.Info("🌍 Or set environment variable:")
		utils.Info("   export GEMINI_API_KEY=YOUR_API_KEY_HERE")
		utils.Info("")
		utils.Info("📖 Get your free API key:")
		utils.Info("   🔗 https://aistudio.google.com/app/apikey")
		utils.Info("")
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

	// // Ensure RATE_LIMIT is set
	// if _, exists := settings["RATE_LIMIT"]; !exists {
    //     utils.Debug("[Config]: Setting default RATE_LIMIT to 15")
    //     settings["RATE_LIMIT"] = 15
    // }

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
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
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
	if err := os.MkdirAll(dir, 0750); err != nil {
		utils.Error("[" + Aliases.Config + "]: ⚠️ Error creating config directory: " + err.Error())
		return
	}

	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
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

	// Save synchronously to ensure the change is persisted immediately
	configFilePath := os.Getenv("HOME") + "/.gitcury/config.json"

	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		utils.Error("[" + Aliases.Config + "]: ⚠️ Error saving configuration after removal: " + err.Error())
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(settings); err != nil {
		utils.Error("[" + Aliases.Config + "]: ⚠️ Error saving configuration after removal: " + err.Error())
	}
}

func Delete() {
	utils.Debug("[" + Aliases.Config + "]: 🗑️ Deleting configuration file...")
	mu.Lock()
	defer mu.Unlock()
	configFilePath := os.Getenv("HOME") + "/.gitcury/config.json"
	if err := os.Remove(configFilePath); err != nil {
		utils.Error("[" + Aliases.Config + "]: ⚠️ Error deleting configuration file: " + err.Error())
	}
}
