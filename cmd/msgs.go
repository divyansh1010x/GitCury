// package cmd

// import (
// 	"GitCury/core"
// 	"GitCury/output"
// 	"GitCury/utils"

// 	"github.com/spf13/cobra"
// )

// var (
// 	numFiles       int
// 	rootFolderName string
// 	allFlag        bool
// )

// var getMsgsCmd = &cobra.Command{
// 	Use:   "getmsgs",
// 	Short: "Generate commit messages for changed files",
// 	Long: `
// This command generates commit messages for changed files.
// You can use the --all flag to generate messages for all changed files in all configured root folders,
// or use the --root flag to specify a particular root folder.
// For example:

// 	getmsgs --all --num 5

// or

// 	getmsgs --root my-folder --num 5
// `,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if allFlag {
// 			err := core.GetAllMsgs(numFiles)
// 			if err != nil {
// 				utils.Error(err.Error())
// 				return
// 			}

// 			allOutput := output.GetAll()
// 			utils.Info("Successfully generated commit messages for all changed files in all configured root folders.")

// 			utils.Print(utils.ToJSON(allOutput))
// 		} else if rootFolderName != "" {
// 			err := core.GetMsgsForRootFolder(rootFolderName, numFiles)
// 			if err != nil {
// 				utils.Error(err.Error())
// 				return
// 			}

// 			rootFolder := output.GetFolder(rootFolderName)
// 			if len(rootFolder.Files) == 0 {
// 				utils.Error("No files to commit in the specified root folder.")
// 				return
// 			}

// 			utils.Info("Successfully generated commit messages for changed files in the specified root folder.")
// 			utils.Print(utils.ToJSON(rootFolder))
// 		} else {
// 			utils.Error("You must specify either --all or --root flag.")
// 		}
// 	},
// }

// func init() {

// 	getMsgsCmd.Flags().IntVarP(&numFiles, "num", "n", 0, "Number of files to commit per folder (overrides config)")
// 	getMsgsCmd.Flags().StringVarP(&rootFolderName, "root", "r", "", "Root folder to commit in (overrides config)")
// 	getMsgsCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Generate commit messages for all changed files in all configured root folders")

// 	rootCmd.AddCommand(getMsgsCmd)
// }

// package cmd

// import (
// 	"GitCury/core"
// 	"GitCury/output"
// 	"GitCury/utils"

// 	"github.com/spf13/cobra"
// )

// var (
// 	numFiles       int
// 	rootFolderName string
// 	allFlag        bool
// )

// var genesisCmd = &cobra.Command{
// 	Use:   "genesis",
// 	Short: "Forge commit messages for altered files",
// 	Long: `
// ╔══════════════════════════════════════════════════════════╗
// ║                  GENESIS: MESSAGE FORGER                 ║
// ╚══════════════════════════════════════════════════════════╝

// [INITIATING]: The Genesis Protocol—crafting commit messages with precision.

// Operational Modes:
// • --all : 🌐 Forge commit messages for all altered files across all root folders.
// • --root <folder> : 📂 Specify a root folder to localize commit message generation.

// Examples:
// • Forge for all folders:
// 	gitcury genesis --all --num 5

// • Target a specific root folder:
// 	gitcury genesis --root my-folder --num 5

// [NOTICE]: Ensure proper configuration of root folders to optimize message crafting.
// `,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if allFlag {
// 			utils.Info("[GENESIS]: 🌌 Forging messages for all root folders.")
// 			err := core.GetAllMsgs(numFiles)
// 			if err != nil {
// 				utils.Error("[GENESIS.FAIL]: ❌ " + err.Error())
// 				return
// 			}

// 			allOutput := output.GetAll()
// 			utils.Success("[GENESIS.SUCCESS]: ✨ Commit messages crafted for all root folders.")
// 			utils.Print(utils.ToJSON(allOutput))
// 		} else if rootFolderName != "" {
// 			utils.Info("[GENESIS]: 📂 Targeting root folder: " + rootFolderName)
// 			err := core.GetMsgsForRootFolder(rootFolderName, numFiles)
// 			if err != nil {
// 				utils.Error("[GENESIS.FAIL]: ❌ " + err.Error())
// 				return
// 			}

// 			rootFolder := output.GetFolder(rootFolderName)
// 			if len(rootFolder.Files) == 0 {
// 				utils.Error("[GENESIS.FAIL]: ⚠️ No altered files detected in the specified root folder.")
// 				return
// 			}

// 			utils.Success("[GENESIS.SUCCESS]: ✨ Commit messages crafted for root folder: " + rootFolderName)
// 			utils.Print(utils.ToJSON(rootFolder))
// 		} else {
// 			utils.Error("[GENESIS.FAIL]: ❗ Specify either --all or --root flag to proceed.")
// 		}
// 	},
// }

// func init() {
// 	genesisCmd.Flags().IntVarP(&numFiles, "num", "n", 0, "🔢 Limit the number of files per commit (overrides config)")
// 	genesisCmd.Flags().StringVarP(&rootFolderName, "root", "r", "", "📂 Specify a root folder for localized message crafting")
// 	genesisCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "🌐 Forge messages for all altered files across all root folders")

// 	rootCmd.AddCommand(genesisCmd)
// }

package cmd

import (
	"GitCury/config"
	"GitCury/core"
	"GitCury/output"
	"GitCury/utils"

	"github.com/spf13/cobra"
)

var (
	numFiles           int
	rootFolderName     string
	allFlag            bool
	groupFlag          bool
	customInstructions string
)

var getMsgsCmd = &cobra.Command{
	Use:   "getmsgs",
	Short: "Generate commit messages for changed files and grouping options",
	Long: `
Generate commit messages for changed files and grouping options.

Aliases:
• ` + config.Aliases.GetMsgs + `

Options:
• --all : Generate commit messages for all changed files in all root folders.
• --root <folder> : Generate commit messages for changed files in a specific root folder.
• --num <number> : Limit the number of files per commit (overrides config).
• --group : Group commit messages by file type.
• --help : Display this help message.

Examples:
• Generate messages for all folders:
	gitcury getmsgs --all --num 5

• Generate messages for a specific folder:
	gitcury getmsgs --root my-folder --num 5

• Generate messages for all folders with grouping:
	gitcury getmsgs --all --num 5 --group

• Generate messages for a specific folder with grouping:
	gitcury getmsgs --root my-folder --num 5 --group

• Generate messages with custom instructions:
	gitcury getmsgs --all --instructions "Don't add keywords like 'feat' or others in front of commit msgs and make humanize msgs"

[NOTICE]: Ensure proper configuration of root folders to optimize message generation.
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Start stats tracking if enabled
		if utils.IsStatsEnabled() {
			if allFlag {
				utils.StartOperation("GenerateAllMessages")
				// Set initial progress
				utils.UpdateOperationProgress("GenerateAllMessages", 10.0)
			} else if rootFolderName != "" {
				utils.StartOperation("GenerateRootMessages")
				// Set initial progress
				utils.UpdateOperationProgress("GenerateRootMessages", 10.0)
			}
		}

		// Handle custom instructions temporarily (not saved to config)
		var originalInstructions interface{}
		var hadInstructions bool
		if customInstructions != "" {
			// Store original instructions if they exist
			originalInstructions = config.Get("commit_instructions")
			hadInstructions = originalInstructions != nil && originalInstructions != ""

			// Set custom instructions temporarily
			config.Set("commit_instructions", customInstructions)
			utils.Debug("[" + config.Aliases.GetMsgs + "]: Using custom instructions (temporary): " + customInstructions)

			// Defer cleanup to ensure instructions are removed
			defer func() {
				if hadInstructions {
					config.Set("commit_instructions", originalInstructions)
					utils.Debug("[" + config.Aliases.GetMsgs + "]: Restored original commit instructions")
				} else {
					config.Remove("commit_instructions")
					utils.Debug("[" + config.Aliases.GetMsgs + "]: Removed temporary custom instructions")
				}
			}()
		}

		if allFlag {
			utils.Info("[" + config.Aliases.GetMsgs + "]: Generating messages for all root folders.")
			var err error

			if groupFlag {
				utils.Debug("[" + config.Aliases.GetMsgs + "]: Grouping logic enabled via --group flag for all folders.")
				err = core.GroupAndGetAllMsgs(numFiles)
			} else {
				err = core.GetAllMsgs(numFiles)
			}

			if err != nil {
				if utils.IsStatsEnabled() {
					utils.FailOperation("GenerateAllMessages", err.Error())
				}
				utils.Error("[" + config.Aliases.GetMsgs + "]: Error encountered - " + err.Error())
				return
			}

			if utils.IsStatsEnabled() {
				utils.MarkOperationComplete("GenerateAllMessages")
			}
			allOutput := output.GetAll()
			utils.Success("[" + config.Aliases.GetMsgs + "]: Commit messages generated for all root folders successfully.")
			utils.Print(utils.ToJSON(allOutput))
		} else if rootFolderName != "" {
			utils.Info("[" + config.Aliases.GetMsgs + "]: Targeting root folder: " + rootFolderName + "\n")

			var err error
			if groupFlag {
				utils.Debug("[" + config.Aliases.GetMsgs + "]: Grouping logic enabled via --group flag.")
				err = core.GroupAndGetMsgsForRootFolder(rootFolderName, numFiles)
			} else {
				err = core.GetMsgsForRootFolder(rootFolderName, numFiles)
			}

			if err != nil {
				if utils.IsStatsEnabled() {
					utils.FailOperation("GenerateRootMessages", err.Error())
				}
				utils.Error("[" + config.Aliases.GetMsgs + "]: Error encountered - " + err.Error())
				return
			}

			rootFolder := output.GetFolder(rootFolderName)
			if len(rootFolder.Files) == 0 {
				if utils.IsStatsEnabled() {
					utils.FailOperation("GenerateRootMessages", "No changed files detected")
				}
				utils.Error("[" + config.Aliases.GetMsgs + "]: No changed files detected in the specified root folder.")
				return
			}

			if utils.IsStatsEnabled() {
				utils.MarkOperationComplete("GenerateRootMessages")
			}
			utils.Success("[" + config.Aliases.GetMsgs + "]: Commit messages generated for root folder: " + rootFolderName + " successfully.")
			utils.Print(utils.ToJSON(rootFolder))
		} else {
			if utils.IsStatsEnabled() {
				utils.FailOperation("GenerateMessages", "No operation flag specified")
			}
			utils.Error("[" + config.Aliases.GetMsgs + "]: You must specify either --all or --root flag.")
		}

		// Display stats if enabled
		if utils.IsStatsEnabled() {
			// Capture clustering configuration for getmsgs command
			clusteringConfig := config.GetClusteringConfig()

			// Determine enabled methods
			enabledMethods := []string{}
			if clusteringConfig.Methods.Directory.Enabled {
				enabledMethods = append(enabledMethods, "directory")
			}
			if clusteringConfig.Methods.Pattern.Enabled {
				enabledMethods = append(enabledMethods, "pattern")
			}
			if clusteringConfig.Methods.Cached.Enabled {
				enabledMethods = append(enabledMethods, "cached")
			}
			if clusteringConfig.Methods.Semantic.Enabled {
				enabledMethods = append(enabledMethods, "semantic")
			}

			// Determine performance mode
			performanceMode := "balanced"
			if clusteringConfig.Performance.PreferSpeed {
				if clusteringConfig.Performance.MaxProcessingTime <= 30 {
					performanceMode = "speed"
				} else {
					performanceMode = "balanced"
				}
			} else {
				performanceMode = "quality"
			}

			// Add grouping status to the method information
			methodUsed := clusteringConfig.DefaultMethod
			if groupFlag {
				methodUsed = clusteringConfig.DefaultMethod + " (grouping enabled)"
			}

			// Set clustering information for stats display
			utils.SetClusteringInfo(
				methodUsed,
				enabledMethods,
				clusteringConfig.ConfidenceThresholds,
				clusteringConfig.SimilarityThresholds,
				clusteringConfig.MaxFilesForSemanticClustering,
				clusteringConfig.EnableFallbackMethods,
				performanceMode,
				clusteringConfig.Performance.MaxProcessingTime,
				clusteringConfig.Performance.EnableBenchmarking,
				clusteringConfig.Performance.AdaptiveOptimization,
			)

			utils.PrintStats()
		}
	},
}

func init() {
	getMsgsCmd.Flags().IntVarP(&numFiles, "num", "n", 0, "Limit the number of files per commit (overrides config)")
	getMsgsCmd.Flags().StringVarP(&rootFolderName, "root", "r", "", "Specify a root folder for localized message generation")
	getMsgsCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Generate messages for all changed files across all root folders")
	getMsgsCmd.Flags().BoolVarP(&groupFlag, "group", "g", false, "Group commit messages by file type")
	getMsgsCmd.Flags().StringVarP(&customInstructions, "instructions", "i", "", "Custom instructions for commit message generation (not saved to config)")

	rootCmd.AddCommand(getMsgsCmd)
}
