package cmd

import (
	"GitCury/config"
	"GitCury/utils"
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
)

var deleteConfig bool
var configSetKey string
var configSetValue string

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage GitCury configuration",
	Long:  "Get and set configuration for GitCury including API keys, root folders, and other parameters.",
	Run: func(cmd *cobra.Command, args []string) {
		if deleteConfig {
			config.Delete()
			utils.Info("Configuration deleted.")
			return
		}

		conf := config.GetAll()
		b, _ := json.MarshalIndent(conf, "", "  ")
		utils.Print("\n==================== GitCury Configuration ====================\n")
		utils.Print(string(b))
		utils.Print("\n==============================================================\n")
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a configuration key-value pair",
	Long: `
The 'config set' command allows you to configure GitCury by setting a specific key-value pair. 
This command is essential for customizing the application's behavior and ensuring it operates as per your requirements.

Usage:
	gitcury config set --key <key> --value <value>

Description:
	This command updates the application's configuration by assigning the specified value to the given key. 
	It supports both simple key-value pairs and more complex configurations like lists of paths.

Key Details:
	- GEMINI_API_KEY (Required): The API key for the Gemini service, which is critical for generating AI-powered commit messages.
	- root_folders (Optional): A comma-separated list of root folder paths where Git operations should be scoped. Example: "/path/to/folder1,/path/to/folder2".
	- numFilesToCommit (Optional): The maximum number of files to include in a single commit operation. Default is 5.
	- app_name (Optional): The name of the application. Default is "GitCury".
	- version (Optional): The version of the application. Default is "1.0.0".
	- log_level (Optional): The logging level for the application. Default is "info".
	- editor (Optional): The text editor to use for editing commit messages. Default is "nano".
	- output_file_path (Optional): The path to the output file where generated commit messages are stored. Default is "$HOME/.gitcury/output.json".

Examples:
	- Set a single configuration value:
			gitcury config set --key theme --value dark

	- Set multiple root folders:
			gitcury config set --key root_folders --value /path/to/folder1,/path/to/folder2

Important Notes:
	- Both the --key and --value flags are mandatory. If either is missing, the command will not execute.
	- The "root_folders" key is treated specially and expects a comma-separated list of folder paths, which will be stored as an array of strings.
	- Ensure that the key you are setting is valid and recognized by the application to avoid unexpected behavior.
	- Use this command to configure critical settings like API keys and operational parameters for GitCury.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if configSetKey == "" || configSetValue == "" {
			utils.Error("Both --key and --value flags are required.")
			return
		}

		// Check if the key is "root_folders" to handle it as an array of strings
		if configSetKey == "root_folders" {
			// Split the value by commas to create an array of strings
			values := strings.Split(configSetValue, ",")
			for i := range values {
				values[i] = strings.TrimSpace(values[i]) // Trim spaces around each value
			}
			config.Set(configSetKey, values) // Save as an array of strings
			utils.Info("Configuration updated: " + configSetKey + " = " + utils.ToJSON(values))
		} else {
			// Handle other keys as a single string value
			config.Set(configSetKey, configSetValue)
			utils.Info("Configuration updated: " + configSetKey + " = " + configSetValue)
		}
	},
}

var configRemoveKey string
var configRemoveRoot string

var configRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a configuration key or a specific root folder",
	Long: `
The 'config remove' command allows you to remove a configuration key or a specific root folder from the configuration.

Usage:
	gitcury config remove --key <key>
	gitcury config remove --root <root_folder>

Description:
	- Use the --key flag to remove an entire configuration key and its value.
	- Use the --root flag to remove a specific root folder from the "root_folders" configuration.

Examples:
	- Remove a configuration key:
			gitcury config remove --key theme

	- Remove a specific root folder:
			gitcury config remove --root /path/to/folder1
`,
	Run: func(cmd *cobra.Command, args []string) {
		if configRemoveKey != "" {
			// Remove the entire key from the configuration
			config.Remove(configRemoveKey)
			utils.Info("Configuration key removed: " + configRemoveKey)
		} else if configRemoveRoot != "" {
			// Remove a specific root folder from "root_folders"
			rootFolders, ok := config.Get("root_folders").([]string)
			if !ok {
				utils.Error("'root_folders' is not configured or is not a list.")
				return
			}

			// Filter out the root folder to be removed
			updatedFolders := []string{}
			for _, folder := range rootFolders {
				if folder != configRemoveRoot {
					updatedFolders = append(updatedFolders, folder)
				}
			}

			// Update the configuration
			config.Set("root_folders", updatedFolders)
			utils.Info("Root folder removed: " + configRemoveRoot)
		} else {
			utils.Error("Either --key or --root flag must be provided.")
		}
	},
}

func init() {
	configSetCmd.Flags().StringVarP(&configSetKey, "key", "k", "", "Configuration key to set")
	configSetCmd.Flags().StringVarP(&configSetValue, "value", "v", "", "Configuration value to set")

	configRemoveCmd.Flags().StringVarP(&configRemoveKey, "key", "k", "", "Configuration key to remove")
	configRemoveCmd.Flags().StringVarP(&configRemoveRoot, "root", "r", "", "Specific root folder to remove")

	configCmd.Flags().BoolVarP(&deleteConfig, "delete", "d", false, "Delete the entire configuration")
	configCmd.AddCommand(configRemoveCmd)
	configCmd.AddCommand(configSetCmd)

	rootCmd.AddCommand(configCmd)
}
