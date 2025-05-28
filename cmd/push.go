// package cmd

// import (
// 	"GitCury/core"
// 	"GitCury/utils"

// 	"github.com/spf13/cobra"
// )

// var (
// 	folder      string
// 	pushAllFlag bool
// 	branchName  string
// )

// var pushCmd = &cobra.Command{
// 	Use:   "push",
// 	Short: "Push changes to remote repository",
// 	Long: `
// This command pushes the committed changes to the remote repository.
// It requires that the files have been committed using the 'commit' command.
// For example:
//   gitcury push --all --branch main
// or
//   gitcury push --root my-folder --branch main`,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if pushAllFlag {
// 			err := core.PushAllRoots(branchName)
// 			if err != nil {
// 				utils.Error(err.Error())
// 				return
// 			}
// 			utils.Info("Successfully pushed all changes to remote repository.")
// 		} else if folder != "" {
// 			err := core.PushOneRoot(folder, branchName)
// 			if err != nil {
// 				utils.Error(err.Error())
// 				return
// 			}
// 			utils.Info("Successfully pushed changes in the specified folder to remote repository.")
// 		} else {
// 			utils.Error("You must specify either --all or --root flag.")
// 		}
// 	},
// }

// func init() {
// 	pushCmd.Flags().BoolVarP(&pushAllFlag, "all", "a", false, "Push all changes to remote repository")
// 	pushCmd.Flags().StringVarP(&folder, "root", "r", "", "Push changes in the specified root folder to remote repository")
// 	pushCmd.Flags().StringVarP(&branchName, "branch", "b", "", "Specify the branch to push to (default: current branch)")
// 	rootCmd.AddCommand(pushCmd)
// }

// package cmd

// import (
// 	"GitCury/core"
// 	"GitCury/utils"

// 	"github.com/spf13/cobra"
// )

// var (
// 	targetFolder string
// 	deployAll    bool
// 	targetBranch string
// )

// var deployCmd = &cobra.Command{
// 	Use:   "deploy",
// 	Short: "Transmit changes to the remote repository",
// 	Long: `
// ╔══════════════════════════════════════════════════╗
// ║              DEPLOY: REMOTE TRANSMISSION         ║
// ╚══════════════════════════════════════════════════╝

// [INITIATING]: The Deploy Protocol—synchronizing your committed changes with the remote repository.

// Operational Modes:
// • --all : Transmit all changes across all root folders.
// • --root <folder> : Specify a root folder for localized transmission.

// Examples:
// • Transmit all changes:
// 	gitcury deploy --all --branch main

// • Target a specific root folder:
// 	gitcury deploy --root my-folder --branch dev

// [NOTICE]: Ensure all necessary commits are sealed using the 'seal' command before deployment.
// `,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if deployAll {
// 			utils.Info("[DEPLOY]: Initiating transmission for all root folders.")
// 			err := core.PushAllRoots(targetBranch)
// 			if err != nil {
// 				utils.Error("[DEPLOY.FAIL]: ⚠️ Error transmitting all changes: " + err.Error())
// 				return
// 			}
// 			utils.Success("[DEPLOY.SUCCESS]: 🌐 All changes successfully transmitted to the remote repository.")
// 		} else if targetFolder != "" {
// 			utils.Info("[DEPLOY]: Targeting root folder: " + targetFolder)
// 			err := core.PushOneRoot(targetFolder, targetBranch)
// 			if err != nil {
// 				utils.Error("[DEPLOY.FAIL]: 🚨 Error transmitting changes for folder '" + targetFolder + "': " + err.Error())
// 				return
// 			}
// 			utils.Success("[DEPLOY.SUCCESS]: 📂 Changes from folder '" + targetFolder + "' successfully transmitted to the remote repository.")
// 		} else {
// 			utils.Error("[DEPLOY.FAIL]: ❗ You must specify either --all or --root flag.")
// 		}
// 	},
// }

// func init() {
// 	deployCmd.Flags().BoolVarP(&deployAll, "all", "a", false, "🌐 Transmit all changes to the remote repository")
// 	deployCmd.Flags().StringVarP(&targetFolder, "root", "r", "", "📂 Transmit changes from the specified folder to the remote repository")
// 	deployCmd.Flags().StringVarP(&targetBranch, "branch", "b", "", "🌿 Specify the branch to transmit to (default: current branch)")
// 	rootCmd.AddCommand(deployCmd)
// }

package cmd

import (
	"GitCury/config"
	"GitCury/core"
	"GitCury/utils"

	"github.com/spf13/cobra"
)

var (
	targetFolder string
	deployAll    bool
	targetBranch string
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push changes to the remote repository",
	Long: `
Push committed changes to the remote repository.

Aliases:
• ` + config.Aliases.Push + `

Options:
• --all : Push all changes across all root folders.
• --root <folder> : Push changes in a specific root folder.

Examples:
• Push all changes:
	gitcury push --all --branch main

• Push changes in a folder:
	gitcury push --root my-folder --branch dev
`,
	Run: func(cmd *cobra.Command, args []string) {
		if deployAll {
			utils.Info("[" + config.Aliases.Push + "]: Pushing all changes to the remote repository.")
			
			// Start stats tracking if enabled
			if utils.IsStatsEnabled() {
				utils.StartOperation("PushAllRoots")
				// Set initial progress
				utils.UpdateOperationProgress("PushAllRoots", 10.0)
			}
			
			err := core.PushAllRoots(targetBranch)
			if err != nil {
				if utils.IsStatsEnabled() {
					utils.FailOperation("PushAllRoots", err.Error())
				}
				utils.Error("[" + config.Aliases.Push + "]: Error pushing all changes: " + err.Error())
				return
			}
			
			if utils.IsStatsEnabled() {
				utils.MarkOperationComplete("PushAllRoots")
			}
			utils.Success("[" + config.Aliases.Push + "]: All changes pushed successfully.")
		} else if targetFolder != "" {
			utils.Info("[" + config.Aliases.Push + "]: Pushing changes from folder: " + targetFolder)
			
			// Start stats tracking if enabled
			if utils.IsStatsEnabled() {
				utils.StartOperation("PushOneRoot")
				// Set initial progress
				utils.UpdateOperationProgress("PushOneRoot", 10.0)
			}
			
			err := core.PushOneRoot(targetFolder, targetBranch)
			if err != nil {
				if utils.IsStatsEnabled() {
					utils.FailOperation("PushOneRoot", err.Error())
				}
				utils.Error("[" + config.Aliases.Push + "]: Error pushing changes from folder '" + targetFolder + "': " + err.Error())
				return
			}
			
			if utils.IsStatsEnabled() {
				utils.MarkOperationComplete("PushOneRoot")
			}
			utils.Success("[" + config.Aliases.Push + "]: Changes from folder '" + targetFolder + "' pushed successfully.")
		} else {
			utils.Error("[" + config.Aliases.Push + "]: You must specify either --all or --root flag.")
		}
	},
}

func init() {
	pushCmd.Flags().BoolVarP(&deployAll, "all", "a", false, "Push all changes to the remote repository")
	pushCmd.Flags().StringVarP(&targetFolder, "root", "r", "", "Push changes from the specified folder to the remote repository")
	pushCmd.Flags().StringVarP(&targetBranch, "branch", "b", "", "Specify the branch to push to (default: current branch)")
	rootCmd.AddCommand(pushCmd)
}
