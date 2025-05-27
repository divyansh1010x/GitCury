// package cmd

// import (
// 	"GitCury/config"
// 	"GitCury/utils"
// 	"encoding/json"
// 	"strings"

// 	"github.com/spf13/cobra"
// )

// var deleteConfig bool
// var configSetKey string
// var configSetValue string

// var configCmd = &cobra.Command{
// 	Use:   "config",
// 	Short: "Manage GitCury configuration",
// 	Long:  "Get and set configuration for GitCury including API keys, root folders, and other parameters.",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if deleteConfig {
// 			config.Delete()
// 			utils.Info("Configuration deleted.")
// 			return
// 		}

// 		conf := config.GetAll()
// 		b, _ := json.MarshalIndent(conf, "", "  ")
// 		utils.Print("\n==================== GitCury Configuration ====================\n")
// 		utils.Print(string(b))
// 		utils.Print("\n==============================================================\n")
// 	},
// }

// var configSetCmd = &cobra.Command{
// 	Use:   "set",
// 	Short: "Set a configuration key-value pair",
// 	Long: `
// The 'config set' command allows you to configure GitCury by setting a specific key-value pair.
// This command is essential for customizing the application's behavior and ensuring it operates as per your requirements.

// Usage:
// 	gitcury config set --key <key> --value <value>

// Description:
// 	This command updates the application's configuration by assigning the specified value to the given key.
// 	It supports both simple key-value pairs and more complex configurations like lists of paths.

// Key Details:
// 	- GEMINI_API_KEY (Required): The API key for the Gemini service, which is critical for generating AI-powered commit messages.
// 	- root_folders (Optional): A comma-separated list of root folder paths where Git operations should be scoped. Example: "/path/to/folder1,/path/to/folder2".
// 	- numFilesToCommit (Optional): The maximum number of files to include in a single commit operation. Default is 5.
// 	- app_name (Optional): The name of the application. Default is "GitCury".
// 	- version (Optional): The version of the application. Default is "1.0.0".
// 	- log_level (Optional): The logging level for the application. Default is "info".
// 	- editor (Optional): The text editor to use for editing commit messages. Default is "nano".
// 	- output_file_path (Optional): The path to the output file where generated commit messages are stored. Default is "$HOME/.gitcury/output.json".

// Examples:
// 	- Set a single configuration value:
// 			gitcury config set --key theme --value dark

// 	- Set multiple root folders:
// 			gitcury config set --key root_folders --value /path/to/folder1,/path/to/folder2

// Important Notes:
// 	- Both the --key and --value flags are mandatory. If either is missing, the command will not execute.
// 	- The "root_folders" key is treated specially and expects a comma-separated list of folder paths, which will be stored as an array of strings.
// 	- Ensure that the key you are setting is valid and recognized by the application to avoid unexpected behavior.
// 	- Use this command to configure critical settings like API keys and operational parameters for GitCury.
// `,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if configSetKey == "" || configSetValue == "" {
// 			utils.Error("Both --key and --value flags are required.")
// 			return
// 		}

// 		// Check if the key is "root_folders" to handle it as an array of strings
// 		if configSetKey == "root_folders" {
// 			// Split the value by commas to create an array of strings
// 			values := strings.Split(configSetValue, ",")
// 			for i := range values {
// 				values[i] = strings.TrimSpace(values[i]) // Trim spaces around each value
// 			}
// 			config.Set(configSetKey, values) // Save as an array of strings
// 			utils.Info("Configuration updated: " + configSetKey + " = " + utils.ToJSON(values))
// 		} else {
// 			// Handle other keys as a single string value
// 			config.Set(configSetKey, configSetValue)
// 			utils.Info("Configuration updated: " + configSetKey + " = " + configSetValue)
// 		}
// 	},
// }

// var configRemoveKey string
// var configRemoveRoot string

// var configRemoveCmd = &cobra.Command{
// 	Use:   "remove",
// 	Short: "Remove a configuration key or a specific root folder",
// 	Long: `
// The 'config remove' command allows you to remove a configuration key or a specific root folder from the configuration.

// Usage:
// 	gitcury config remove --key <key>
// 	gitcury config remove --root <root_folder>

// Description:
// 	- Use the --key flag to remove an entire configuration key and its value.
// 	- Use the --root flag to remove a specific root folder from the "root_folders" configuration.

// Examples:
// 	- Remove a configuration key:
// 			gitcury config remove --key theme

// 	- Remove a specific root folder:
// 			gitcury config remove --root /path/to/folder1
// `,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if configRemoveKey != "" {
// 			// Remove the entire key from the configuration
// 			config.Remove(configRemoveKey)
// 			utils.Info("Configuration key removed: " + configRemoveKey)
// 		} else if configRemoveRoot != "" {
// 			// Remove a specific root folder from "root_folders"
// 			rootFolders, ok := config.Get("root_folders").([]string)
// 			if !ok {
// 				utils.Error("'root_folders' is not configured or is not a list.")
// 				return
// 			}

// 			// Filter out the root folder to be removed
// 			updatedFolders := []string{}
// 			for _, folder := range rootFolders {
// 				if folder != configRemoveRoot {
// 					updatedFolders = append(updatedFolders, folder)
// 				}
// 			}

// 			// Update the configuration
// 			config.Set("root_folders", updatedFolders)
// 			utils.Info("Root folder removed: " + configRemoveRoot)
// 		} else {
// 			utils.Error("Either --key or --root flag must be provided.")
// 		}
// 	},
// }

// func init() {
// 	configSetCmd.Flags().StringVarP(&configSetKey, "key", "k", "", "Configuration key to set")
// 	configSetCmd.Flags().StringVarP(&configSetValue, "value", "v", "", "Configuration value to set")

// 	configRemoveCmd.Flags().StringVarP(&configRemoveKey, "key", "k", "", "Configuration key to remove")
// 	configRemoveCmd.Flags().StringVarP(&configRemoveRoot, "root", "r", "", "Specific root folder to remove")

// 	configCmd.Flags().BoolVarP(&deleteConfig, "delete", "d", false, "Delete the entire configuration")
// 	configCmd.AddCommand(configRemoveCmd)
// 	configCmd.AddCommand(configSetCmd)

// 	rootCmd.AddCommand(configCmd)
// }

// package cmd

// import (
// 	"GitCury/config"
// 	"GitCury/utils"
// 	"encoding/json"
// 	"strings"

// 	"github.com/spf13/cobra"
// )

// var deleteConfig bool
// var configSetKey string
// var configSetValue string
// var configRemoveKey string
// var configRemoveRoot string

// var nexusCmd = &cobra.Command{
// 	Use:   "nexus",
// 	Short: "Access the central configuration nexus",
// 	Long: `
// ╔══════════════════════════════════════════════════════════╗
// ║                  "+ config.Aliases.Config +": CONFIGURATION CORE               ║
// ╚══════════════════════════════════════════════════════════╝

// [INITIATING]: The Nexus Protocol—manage critical system parameters.

// Capabilities:
// • 🔑 API authentication protocols
// • 📂 File system access points
// • 🧠 Neural network parameters
// • 🛠️ System memory allocation

// Configuration Keys:
// • GEMINI_API_KEY (Required): API key for Gemini service.
// • root_folders (Optional): Comma-separated list of root folder paths.
// • numFilesToCommit (Optional): Max number of files per commit (default: 5).
// • app_name (Optional): Application name (default: "GitCury").
// • version (Optional): Application version (default: "1.0.0").
// • log_level (Optional): Logging level (default: "info").
// • editor (Optional): Text editor for editing commit messages (default: "nano").
// • output_file_path (Optional): Path to output file (default: "$HOME/.gitcury/output.json").
// • retries (Optional): Number of retries for operations (default: 3).
// • timeout (Optional): Timeout duration for operations (default: 30 seconds).

// [NOTICE]: Unauthorized changes may destabilize the system.
// `,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if deleteConfig {
// 			config.Delete()
// 			utils.Success("[" + config.Aliases.Config + "]: 🗑️ Configuration nexus obliterated.")
// 			return
// 		}

// 		conf := config.GetAll()
// 		b, _ := json.MarshalIndent(conf, "", "  ")
// 		utils.Print("\n======== " + config.Aliases.Config + " CONFIGURATION STATUS ========\n")
// 		utils.Print(string(b))
// 		utils.Print("\n============================================\n")
// 	},
// }

// var injectCmd = &cobra.Command{
// 	Use:   "inject",
// 	Short: "💉 Inject key-value pairs into the nexus",
// 	Long: `
// ╔══════════════════════════════════════════════════╗
// ║              INJECT: CONFIGURATION UPDATE        ║
// ╚══════════════════════════════════════════════════╝

// [INITIATING]: The Inject Protocol—update or add directives to the configuration nexus.

// Examples:
// • Inject a new directive:
// 	gitcury inject --key GEMINI_API_KEY --value YOUR_API_KEY

// • Update root folders:
// 	gitcury inject --key root_folders --value /path/to/folder1,/path/to/folder2
// `,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if configSetKey == "" || configSetValue == "" {
// 			utils.Error("[" + config.Aliases.Config + "]: ❌ Injection failed. Missing --key or --value.")
// 			return
// 		}

// 		if configSetKey == "root_folders" {
// 			values := strings.Split(configSetValue, ",")
// 			for i := range values {
// 				values[i] = strings.TrimSpace(values[i])
// 			}
// 			config.Set(configSetKey, values)
// 			utils.Success("[" + config.Aliases.Config + "]: ✅ Directive injected: " + configSetKey + " = " + utils.ToJSON(values))
// 		} else {
// 			config.Set(configSetKey, configSetValue)
// 			utils.Success("[" + config.Aliases.Config + "]: ✅ Directive injected: " + configSetKey + " = " + configSetValue)
// 		}
// 	},
// }

// var purgeCmd = &cobra.Command{
// 	Use:   "purge",
// 	Short: "🗑️ Purge directives from the nexus",
// 	Long: `
// ╔══════════════════════════════════════════════════╗
// ║              PURGE: CONFIGURATION CLEANUP        ║
// ╚══════════════════════════════════════════════════╝

// [INITIATING]: The Purge Protocol—remove directives or root folders from the nexus.

// Examples:
// • Purge a configuration key:
// 	gitcury purge --key theme

// • Purge a specific root folder:
// 	gitcury purge --root /path/to/folder1
// `,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if configRemoveKey != "" {
// 			config.Remove(configRemoveKey)
// 			utils.Success("[" + config.Aliases.Config + "]: 🗑️ Directive purged: " + configRemoveKey)
// 		} else if configRemoveRoot != "" {
// 			rootFolders, ok := config.Get("root_folders").([]string)
// 			if !ok {
// 				utils.Error("[" + config.Aliases.Config + "]: ❌ Root folders directive missing or corrupted.")
// 				return
// 			}

// 			updatedFolders := []string{}
// 			for _, folder := range rootFolders {
// 				if folder != configRemoveRoot {
// 					updatedFolders = append(updatedFolders, folder)
// 				}
// 			}

// 			config.Set("root_folders", updatedFolders)
// 			utils.Success("[" + config.Aliases.Config + "]: 🗑️ Root folder purged: " + configRemoveRoot)
// 		} else {
// 			utils.Error("[" + config.Aliases.Config + "]: ❌ Specify either --key or --root for purge operation.")
// 		}
// 	},
// }

// func init() {
// 	injectCmd.Flags().StringVarP(&configSetKey, "key", "k", "", "🔑 Directive key to inject")
// 	injectCmd.Flags().StringVarP(&configSetValue, "value", "v", "", "📄 Directive value to inject")

// 	purgeCmd.Flags().StringVarP(&configRemoveKey, "key", "k", "", "🔑 Directive key to purge")
// 	purgeCmd.Flags().StringVarP(&configRemoveRoot, "root", "r", "", "📂 Specific root folder to purge")

// 	nexusCmd.Flags().BoolVarP(&deleteConfig, "delete", "d", false, "🗑️ Obliterate all directives from the nexus")
// 	nexusCmd.AddCommand(purgeCmd)
// 	nexusCmd.AddCommand(injectCmd)

// 	rootCmd.AddCommand(nexusCmd)
// }

package cmd

import (
	"GitCury/config"
	"GitCury/utils"
	"encoding/json"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var deleteConfig bool // Flag to reset configuration to defaults
var configSetKey string
var configSetValue string
var configRemoveKey string
var configRemoveRoot string

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage GitCury configuration",
	Long: `
Manage GitCury configuration settings.

Aliases:
• ` + config.Aliases.Config + `

Configuration Keys:
• GEMINI_API_KEY (Required): API key for Gemini service
• root_folders (Optional): Comma-separated list of root folder paths
• numFilesToCommit (Optional): Max number of files per commit (default: 5)
• app_name (Optional): Application name (default: "GitCury")
• version (Optional): Application version (default: "1.0.0")
• log_level (Optional): Logging level (default: "info")
• editor (Optional): Text editor for editing commit messages (default: "nano")
• output_file_path (Optional): Path to output file (default: "$HOME/.gitcury/output.json")
• retries (Optional): Number of retries for operations (default: 3)
• timeout (Optional): Timeout duration for operations (default: 30 seconds)

Examples:
• View current configuration:
	gitcury config

• Set API key:
	gitcury config set --key GEMINI_API_KEY --value YOUR_API_KEY_HERE

• Update root folders:
	gitcury config set --key root_folders --value /path/to/folder1,/path/to/folder2
`,
	Run: func(cmd *cobra.Command, args []string) {
		if deleteConfig {
			utils.Info("Resetting all configuration to defaults...")
			config.Delete()
			utils.Success("Configuration reset successfully.")
			return
		}

		// Always show config, even if there are missing values
		conf := config.GetAll()

		// Ensure we have at least basic config structure
		if conf == nil || len(conf) == 0 {
			conf = map[string]interface{}{
				"app_name":         "GitCury",
				"version":          "1.0.0",
				"root_folders":     []string{"."},
				"numFilesToCommit": 5,
				"editor":           "nano",
				"retries":          3,
				"timeout":          30,
				"logLevel":         "info",
			}

			// Save this basic config
			for key, value := range conf {
				config.Set(key, value)
			}

			utils.Info("📝 Created basic configuration with default values")
		}

		// Check if API key is missing and provide helpful guidance
		apiKey, hasApiKey := conf["GEMINI_API_KEY"]
		envKey := os.Getenv("GEMINI_API_KEY")

		if !hasApiKey || apiKey == "" {
			if envKey == "" {
				utils.Info("📋 Current Configuration (⚠️  API key missing)")
				utils.Info("════════════════════════════════════════")
			} else {
				utils.Info("📋 Current Configuration (✅ API key from environment)")
				utils.Info("════════════════════════════════════════")
				conf["GEMINI_API_KEY"] = "[FROM ENVIRONMENT: " + envKey[:10] + "...]"
			}
		} else {
			// Mask the API key for security
			if keyStr, ok := apiKey.(string); ok && len(keyStr) > 10 {
				conf["GEMINI_API_KEY"] = keyStr[:10] + "..." + " (configured)"
			}
			utils.Info("📋 Current Configuration (✅ API key configured)")
			utils.Info("════════════════════════════════════════")
		}

		// Display config in a user-friendly format
		b, _ := json.MarshalIndent(conf, "", "  ")
		utils.Print(string(b))
		utils.Print("")

		// Provide helpful guidance if API key is missing
		if (!hasApiKey || apiKey == "") && envKey == "" {
			utils.Info("🔑 Next Steps:")
			utils.Info("   To use GitCury's AI features, set your Gemini API key:")
			utils.Info("")
			utils.Info("   gitcury config set --key GEMINI_API_KEY --value YOUR_API_KEY_HERE")
			utils.Info("")
			utils.Info("📖 Get your free API key:")
			utils.Info("   🔗 https://aistudio.google.com/app/apikey")
			utils.Info("")
			utils.Info("💡 Tip: You can also set the environment variable:")
			utils.Info("   export GEMINI_API_KEY=your_key_here")
			utils.Info("")
		} else {
			utils.Info("✅ Configuration looks good! You're ready to use GitCury.")
			utils.Info("")
			utils.Info("💡 Try these commands:")
			utils.Info("   gitcury getmsgs    # Generate AI commit messages")
			utils.Info("   gitcury commit     # Commit changes")
			utils.Info("   gitcury --help     # See all available commands")
			utils.Info("")
		}
	},
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration values",
	Long: `
Set or update configuration values.

Examples:
• Set API key:
	gitcury config set --key GEMINI_API_KEY --value YOUR_API_KEY_HERE

• Update root folders:
	gitcury config set --key root_folders --value /path/to/folder1,/path/to/folder2
	
• Set numeric value:
	gitcury config set --key numFilesToCommit --value 10
`,
	Run: func(cmd *cobra.Command, args []string) {
		if configSetKey == "" || configSetValue == "" {
			utils.Error("Both --key and --value are required.")
			utils.Info("Example: gitcury config set --key GEMINI_API_KEY --value YOUR_API_KEY")
			return
		}

		// Special handling for root_folders
		if configSetKey == "root_folders" {
			values := strings.Split(configSetValue, ",")
			for i := range values {
				values[i] = strings.TrimSpace(values[i])
			}
			config.Set(configSetKey, values)
			utils.Success("✅ Configuration updated: " + configSetKey + " = " + utils.ToJSON(values))
		} else if utils.IsNumeric(configSetValue) {
			// Handle numeric values
			intValue, err := utils.ParseInt(configSetValue)
			if err == nil {
				config.Set(configSetKey, intValue)
				utils.Success("✅ Configuration updated: " + configSetKey + " = " + configSetValue)
			} else {
				// Fall back to string if conversion fails
				config.Set(configSetKey, configSetValue)
				utils.Success("✅ Configuration updated: " + configSetKey + " = " + configSetValue)
			}
		} else {
			config.Set(configSetKey, configSetValue)
			utils.Success("✅ Configuration updated: " + configSetKey + " = " + configSetValue)
		}

		// Provide extra guidance for API key
		if configSetKey == "GEMINI_API_KEY" {
			utils.Info("")
			utils.Info("🎉 API key configured! You can now use GitCury's AI features.")
			utils.Info("💡 Try: gitcury commit --help")
		}
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove configuration values",
	Long: `
Remove configuration keys or specific root folders.

Examples:
• Remove a configuration key:
	gitcury config remove --key theme

• Remove a specific root folder:
	gitcury config remove --root /path/to/folder1
`,
	Run: func(cmd *cobra.Command, args []string) {
		if configRemoveKey != "" {
			utils.Info("Removing configuration key: " + configRemoveKey)
			config.Remove(configRemoveKey)
			utils.Success("✅ Configuration key removed: " + configRemoveKey)
		} else if configRemoveRoot != "" {
			rootFolders, ok := config.Get("root_folders").([]string)
			if !ok {
				utils.Error("Root folders configuration missing or has invalid format.")
				return
			}

			updatedFolders := []string{}
			found := false
			for _, folder := range rootFolders {
				if folder != configRemoveRoot {
					updatedFolders = append(updatedFolders, folder)
				} else {
					found = true
				}
			}

			if !found {
				utils.Warning("Root folder not found in configuration: " + configRemoveRoot)
				return
			}

			config.Set("root_folders", updatedFolders)
			utils.Success("✅ Root folder removed: " + configRemoveRoot)
		} else {
			utils.Error("Specify either --key or --root for remove operation.")
			utils.Info("Examples:")
			utils.Info("  gitcury config remove --key SOME_KEY")
			utils.Info("  gitcury config remove --root /path/to/folder")
		}
	},
}

func init() {
	setCmd.Flags().StringVarP(&configSetKey, "key", "k", "", "Configuration key to set")
	setCmd.Flags().StringVarP(&configSetValue, "value", "v", "", "Configuration value to set")

	removeCmd.Flags().StringVarP(&configRemoveKey, "key", "k", "", "Configuration key to remove")
	removeCmd.Flags().StringVarP(&configRemoveRoot, "root", "r", "", "Specific root folder to remove")

	// Avoid shorthand flag to prevent conflicts with other commands
	configCmd.Flags().BoolVar(&deleteConfig, "reset", false, "Reset all configuration to defaults")
	configCmd.AddCommand(removeCmd)
	configCmd.AddCommand(setCmd)

	rootCmd.AddCommand(configCmd)
}
