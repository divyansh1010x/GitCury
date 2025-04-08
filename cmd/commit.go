// // package cmd

// // import (
// // 	"GitCury/core"
// // 	"GitCury/utils"

// // 	"github.com/spf13/cobra"
// // )

// // var (
// // 	commitAllFlag bool
// // 	folderName    string
// // )
// // var commitCmd = &cobra.Command{
// // 	Use:   "commit",
// // 	Short: "Commit files with generated messages",
// // 	Long: `
// // This command commits files with generated messages.
// // You can use the --all flag to commit all files with generated messages,
// // or use the --root flag to specify a particular root folder.
// // For example:
// //   gitcury commit --all
// // or
// //   gitcury commit --root my-folder`,
// // 	Run: func(cmd *cobra.Command, args []string) {
// // 		if commitAllFlag {
// // 			err := core.CommitAllRoots()
// // 			if err != nil {
// // 				utils.Error(err.Error())
// // 				return
// // 			}
// // 			utils.Info("Successfully committed all files with generated messages.")
// // 		} else if folderName != "" {
// // 			err := core.CommitOneRoot(folderName)
// // 			if err != nil {
// // 				utils.Error(err.Error())
// // 				return
// // 			}
// // 			utils.Info("Successfully committed files in the specified folder with generated messages.")
// // 		} else {
// // 			utils.Error("You must specify either --all or --root flag.")
// // 		}
// // 	},
// // }

// // func init() {
// // 	commitCmd.Flags().BoolVarP(&commitAllFlag, "all", "a", false, "Commit all files with generated messages")
// // 	commitCmd.Flags().StringVarP(&folderName, "root", "r", "", "Commit files in the specified root folder with generated messages")
// // 	rootCmd.AddCommand(commitCmd)
// // }

// // package cmd

// // import (
// // 	"GitCury/core"
// // 	"GitCury/utils"
// // 	"os"
// // 	"os/exec"
// // 	"time"

// // 	"github.com/spf13/cobra"
// // )

// // var (
// // 	commitDateTime string
// // 	commitAllFlag  bool
// // 	folderName     string
// // )

// // var commitCmd = &cobra.Command{
// // 	Use:   "commit",
// // 	Short: "Commit files with generated messages",
// // 	Long: `
// // This command commits files with generated messages.
// // You can use the --all flag to commit all files with generated messages,
// // or use the --root flag to specify a particular root folder.
// // For example:
// //   gitcury commit --all
// // or
// //   gitcury commit --root my-folder`,
// // 	Run: func(cmd *cobra.Command, args []string) {
// // 		if commitAllFlag {
// // 			err := core.CommitAllRoots()
// // 			if err != nil {
// // 				utils.Error(err.Error())
// // 				return
// // 			}
// // 			utils.Info("Successfully committed all files with generated messages.")
// // 		} else if folderName != "" {
// // 			err := core.CommitOneRoot(folderName)
// // 			if err != nil {
// // 				utils.Error(err.Error())
// // 				return
// // 			}
// // 			utils.Info("Successfully committed files in the specified folder with generated messages.")
// // 		} else {
// // 			utils.Error("You must specify either --all or --root flag.")
// // 		}
// // 	},
// // }

// // var commitWithDateCmd = &cobra.Command{
// // 	Use:   "with-date",
// // 	Short: "Commit files with a specified date and time",
// // 	Long: `
// // This subcommand allows you to temporarily change the system date and time,
// // commit files with generated messages, and then restore the original date and time.
// // For example:
// //   gitcury commit with-date --datetime "2023-01-01T12:00:00" --all
// // or
// //   gitcury commit with-date --datetime "2023-01-01T12:00:00" --root my-folder`,
// // 	Run: func(cmd *cobra.Command, args []string) {
// // 		if commitDateTime == "" {
// // 			utils.Error("You must specify the --datetime flag in the format 'YYYY-MM-DDTHH:MM:SS'.")
// // 			return
// // 		}

// // 		// Parse the provided date-time
// // 		_, err := time.Parse("2006-01-02T15:04:05", commitDateTime)
// // 		if err != nil {
// // 			utils.Error("Invalid date-time format. Use 'YYYY-MM-DDTHH:MM:SS'.")
// // 			return
// // 		}

// // 		// Set environment variables for the commit
// // 		utils.Info("Setting commit date and time to: " + commitDateTime)
// // 		env := append(os.Environ(),
// // 			"GIT_AUTHOR_DATE="+commitDateTime,
// // 			"GIT_COMMITTER_DATE="+commitDateTime,
// // 		)
// // 	},
// // }

// // func restoreOriginalDateTime(originalDateTime string) {
// // 	utils.Info("Restoring original system date and time...")
// // 	restoreDateTimeCmd := exec.Command("sudo", "date", "-s", originalDateTime)
// // 	if err := restoreDateTimeCmd.Run(); err != nil {
// // 		utils.Error("Failed to restore original system date and time: " + err.Error())
// // 	} else {
// // 		utils.Info("Successfully restored original system date and time.")
// // 	}
// // }

// // func init() {
// // 	commitWithDateCmd.Flags().BoolVarP(&commitAllFlag, "all", "a", false, "Commit all files with generated messages")
// // 	commitWithDateCmd.Flags().StringVarP(&folderName, "root", "r", "", "Commit files in the specified root folder with generated messages")
// // 	commitWithDateCmd.Flags().StringVarP(&commitDateTime, "datetime", "d", "", "Specify the commit date and time in 'YYYY-MM-DDTHH:MM:SS' format")
// // 	commitCmd.AddCommand(commitWithDateCmd)
// // 	rootCmd.AddCommand(commitCmd)
// // }

// package cmd

// import (
// 	"GitCury/core"
// 	"GitCury/utils"
// 	"os"
// 	"time"

// 	"github.com/spf13/cobra"
// )

// var (
// 	commitDateTime string
// 	commitAllFlag  bool
// 	folderName     string
// )

// var commitCmd = &cobra.Command{
// 	Use:   "commit",
// 	Short: "Commit files with generated messages",
// 	Long: `
// The 'commit' command allows you to commit files with generated messages.
// You can commit all files or specify a root folder for commits.
// For example:
//   gitcury commit --all
// or
//   gitcury commit --root my-folder`,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if commitAllFlag {
// 			err := core.CommitAllRoots()
// 			if err != nil {
// 				utils.Error(err.Error())
// 				return
// 			}
// 			utils.Info("Successfully committed all files with generated messages.")
// 		} else if folderName != "" {
// 			err := core.CommitOneRoot(folderName)
// 			if err != nil {
// 				utils.Error(err.Error())
// 				return
// 			}
// 			utils.Info("Successfully committed files in the specified folder with generated messages.")
// 		} else {
// 			utils.Error("You must specify either --all or --root flag.")
// 		}
// 	},
// }

// var withDateCmd = &cobra.Command{
// 	Use:   "with-date",
// 	Short: "Commit files with a specified date and time",
// 	Long: `
// The 'with-date' subcommand allows you to commit files with a specified date and time.
// This is achieved without changing the system date by using Git's environment variables.
// For example:
//   gitcury commit with-date --datetime "2023-01-01T12:00:00" --all
// or
//   gitcury commit with-date --datetime "2023-01-01T12:00:00" --root my-folder`,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if commitDateTime == "" {
// 			utils.Error("You must specify the --datetime flag in the format 'YYYY-MM-DDTHH:MM:SS'.")
// 			return
// 		}

// 		// Validate the date-time format
// 		_, err := time.Parse("2006-01-02T15:04:05", commitDateTime)
// 		if err != nil {
// 			utils.Error("Invalid date-time format. Use 'YYYY-MM-DDTHH:MM:SS'.")
// 			return
// 		}

// 		// Set Git's environment variables
// 		env := append(os.Environ(),
// 			"GIT_AUTHOR_DATE="+commitDateTime,
// 			"GIT_COMMITTER_DATE="+commitDateTime,
// 		)

// 		utils.Info("Setting commit date and time to: " + commitDateTime)

// 		// Execute the commit logic
// 		var commitErr error
// 		if commitAllFlag {
// 			commitErr = core.CommitAllRoots(env)
// 		} else if folderName != "" {
// 			commitErr = core.CommitOneRoot(folderName, env)
// 		} else {
// 			utils.Error("You must specify either --all or --root flag.")
// 			return
// 		}

// 		if commitErr != nil {
// 			utils.Error("Failed to commit: " + commitErr.Error())
// 			return
// 		}
// 		utils.Info("Successfully committed files with the specified date and time.")
// 	},
// }

// func init() {
// 	// Add flags to the with-date subcommand
// 	withDateCmd.Flags().StringVarP(&commitDateTime, "datetime", "d", "", "Specify the commit date and time in 'YYYY-MM-DDTHH:MM:SS' format")
// 	withDateCmd.Flags().BoolVarP(&commitAllFlag, "all", "a", false, "Commit all files with generated messages")
// 	withDateCmd.Flags().StringVarP(&folderName, "root", "r", "", "Commit files in the specified root folder with generated messages")

// 	// Add the with-date subcommand to the commit command
// 	commitCmd.AddCommand(withDateCmd)

// 	commitCmd.Flags().BoolVarP(&commitAllFlag, "all", "a", false, "Commit all files with generated messages")
// 	commitCmd.Flags().StringVarP(&folderName, "root", "r", "", "Commit files in the specified root folder with generated messages")
// 	// Add the commit command to the root command
// 	rootCmd.AddCommand(commitCmd)
// }


package cmd

import (
	"GitCury/core"
	"GitCury/utils"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	sealDateTime string
	sealAllFlag  bool
	folderName   string
)

var sealCmd = &cobra.Command{
	Use:   "seal",
	Short: "Seal changes with autogenerated commit messages",
	Long: `
╔══════════════════════════════════════════════════╗
║              SEAL: COMMIT MESSAGE ENGRAVER       ║
╚══════════════════════════════════════════════════╝

[INITIATING]: The Seal Protocol—engraving changes with precision.

Operational Modes:
• --all : Seal all changes across all root folders.
• --root <folder> : Specify a root folder for localized sealing.

Examples:
• Seal all changes:
	gitcury seal --all

• Target a specific root folder:
	gitcury seal --root my-folder

[NOTICE]: Ensure commit messages are generated before sealing.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if sealAllFlag {
			utils.Info("[SEAL]: Engraving all changes across root folders.")
			err := core.CommitAllRoots()
			if err != nil {
				utils.Error("[SEAL.FAIL]: ❌ Error encountered - " + err.Error())
				return
			}
			utils.Success("[SEAL.SUCCESS]: 🔒 All changes sealed successfully.")
		} else if folderName != "" {
			utils.Info("[SEAL]: Targeting root folder: " + folderName)
			err := core.CommitOneRoot(folderName)
			if err != nil {
				utils.Error("[SEAL.FAIL]: ❌ Error encountered - " + err.Error())
				return
			}
			utils.Success("[SEAL.SUCCESS]: 🔒 Changes in the specified folder sealed successfully.")
		} else {
			utils.Error("[SEAL.FAIL]: ❗ You must specify either --all or --root flag.")
		}
	},
}

var withDateCmd = &cobra.Command{
	Use:   "with-date",
	Short: "⏳ Seal changes with a specified timestamp",
	Long: `
╔══════════════════════════════════════════════════╗
║        SEAL WITH DATE: PRECISION ENGRAVING       ║
╚══════════════════════════════════════════════════╝

[INITIATING]: The Seal with Date Protocol—engraving changes with a specific timestamp.

Operational Modes:
• --all : Seal all changes across all root folders with a timestamp.
• --root <folder> : Specify a root folder for localized sealing with a timestamp.

Examples:
• Seal all changes with a timestamp:
	gitcury seal with-date --datetime "2025-01-01T12:00:00" --all

• Target a specific root folder with a timestamp:
	gitcury seal with-date --datetime "2025-01-01T12:00:00" --root my-folder

[NOTICE]: Ensure the timestamp is in the format 'YYYY-MM-DDTHH:MM:SS'.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if sealDateTime == "" {
			utils.Error("[SEAL.FAIL]: ❗ You must specify the --datetime flag in the format 'YYYY-MM-DDTHH:MM:SS'.")
			return
		}

		// Validate date-time format
		_, err := time.Parse("2006-01-02T15:04:05", sealDateTime)
		if err != nil {
			utils.Error("[SEAL.FAIL]: ❌ Invalid date-time format. Use 'YYYY-MM-DDTHH:MM:SS'.")
			return
		}

		// Set Git environment variables for the commit
		env := append(os.Environ(),
			"GIT_AUTHOR_DATE="+sealDateTime,
			"GIT_COMMITTER_DATE="+sealDateTime,
		)
		utils.Info("[SEAL]: Setting commit date and time to: " + sealDateTime)

		// Execute commit logic
		var commitErr error
		if sealAllFlag {
			commitErr = core.CommitAllRoots(env)
		} else if folderName != "" {
			commitErr = core.CommitOneRoot(folderName, env)
		} else {
			utils.Error("[SEAL.FAIL]: ❗ You must specify either --all or --root flag.")
			return
		}

		if commitErr != nil {
			utils.Error("[SEAL.FAIL]: ❌ Failed to seal changes - " + commitErr.Error())
			return
		}
		utils.Success("[SEAL.SUCCESS]: 🔒 Changes sealed with the specified timestamp successfully.")
	},
}

func init() {
	// Add flags for the with-date subcommand
	withDateCmd.Flags().StringVarP(&sealDateTime, "datetime", "d", "", "⏳ Specify the commit date and time in 'YYYY-MM-DDTHH:MM:SS' format")
	withDateCmd.Flags().BoolVarP(&sealAllFlag, "all", "a", false, "🔒 Seal all changes with autogenerated messages")
	withDateCmd.Flags().StringVarP(&folderName, "root", "r", "", "📂 Seal changes in the specified root folder with autogenerated messages")

	// Add the with-date subcommand to the seal command
	sealCmd.AddCommand(withDateCmd)

	// Add flags to the main seal command
	sealCmd.Flags().BoolVarP(&sealAllFlag, "all", "a", false, "🔒 Seal all changes with autogenerated messages")
	sealCmd.Flags().StringVarP(&folderName, "root", "r", "", "📂 Seal changes in the specified root folder with autogenerated messages")

	// Add the seal command to the root command
	rootCmd.AddCommand(sealCmd)
}
