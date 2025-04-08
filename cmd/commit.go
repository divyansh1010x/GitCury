// package cmd

// import (
// 	"GitCury/core"
// 	"GitCury/utils"

// 	"github.com/spf13/cobra"
// )

// var (
// 	commitAllFlag bool
// 	folderName    string
// )
// var commitCmd = &cobra.Command{
// 	Use:   "commit",
// 	Short: "Commit files with generated messages",
// 	Long: `
// This command commits files with generated messages.
// You can use the --all flag to commit all files with generated messages,
// or use the --root flag to specify a particular root folder.
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

// func init() {
// 	commitCmd.Flags().BoolVarP(&commitAllFlag, "all", "a", false, "Commit all files with generated messages")
// 	commitCmd.Flags().StringVarP(&folderName, "root", "r", "", "Commit files in the specified root folder with generated messages")
// 	rootCmd.AddCommand(commitCmd)
// }

// package cmd

// import (
// 	"GitCury/core"
// 	"GitCury/utils"
// 	"os"
// 	"os/exec"
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
// This command commits files with generated messages.
// You can use the --all flag to commit all files with generated messages,
// or use the --root flag to specify a particular root folder.
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

// var commitWithDateCmd = &cobra.Command{
// 	Use:   "with-date",
// 	Short: "Commit files with a specified date and time",
// 	Long: `
// This subcommand allows you to temporarily change the system date and time,
// commit files with generated messages, and then restore the original date and time.
// For example:
//   gitcury commit with-date --datetime "2023-01-01T12:00:00" --all
// or
//   gitcury commit with-date --datetime "2023-01-01T12:00:00" --root my-folder`,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if commitDateTime == "" {
// 			utils.Error("You must specify the --datetime flag in the format 'YYYY-MM-DDTHH:MM:SS'.")
// 			return
// 		}

// 		// Parse the provided date-time
// 		_, err := time.Parse("2006-01-02T15:04:05", commitDateTime)
// 		if err != nil {
// 			utils.Error("Invalid date-time format. Use 'YYYY-MM-DDTHH:MM:SS'.")
// 			return
// 		}

// 		// Set environment variables for the commit
// 		utils.Info("Setting commit date and time to: " + commitDateTime)
// 		env := append(os.Environ(),
// 			"GIT_AUTHOR_DATE="+commitDateTime,
// 			"GIT_COMMITTER_DATE="+commitDateTime,
// 		)
// 	},
// }

// func restoreOriginalDateTime(originalDateTime string) {
// 	utils.Info("Restoring original system date and time...")
// 	restoreDateTimeCmd := exec.Command("sudo", "date", "-s", originalDateTime)
// 	if err := restoreDateTimeCmd.Run(); err != nil {
// 		utils.Error("Failed to restore original system date and time: " + err.Error())
// 	} else {
// 		utils.Info("Successfully restored original system date and time.")
// 	}
// }

// func init() {
// 	commitWithDateCmd.Flags().BoolVarP(&commitAllFlag, "all", "a", false, "Commit all files with generated messages")
// 	commitWithDateCmd.Flags().StringVarP(&folderName, "root", "r", "", "Commit files in the specified root folder with generated messages")
// 	commitWithDateCmd.Flags().StringVarP(&commitDateTime, "datetime", "d", "", "Specify the commit date and time in 'YYYY-MM-DDTHH:MM:SS' format")
// 	commitCmd.AddCommand(commitWithDateCmd)
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
	commitDateTime string
	commitAllFlag  bool
	folderName     string
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit files with generated messages",
	Long: `
The 'commit' command allows you to commit files with generated messages.
You can commit all files or specify a root folder for commits.
For example:
  gitcury commit --all
or
  gitcury commit --root my-folder`,
	Run: func(cmd *cobra.Command, args []string) {
		if commitAllFlag {
			err := core.CommitAllRoots()
			if err != nil {
				utils.Error(err.Error())
				return
			}
			utils.Info("Successfully committed all files with generated messages.")
		} else if folderName != "" {
			err := core.CommitOneRoot(folderName)
			if err != nil {
				utils.Error(err.Error())
				return
			}
			utils.Info("Successfully committed files in the specified folder with generated messages.")
		} else {
			utils.Error("You must specify either --all or --root flag.")
		}
	},
}

var withDateCmd = &cobra.Command{
	Use:   "with-date",
	Short: "Commit files with a specified date and time",
	Long: `
The 'with-date' subcommand allows you to commit files with a specified date and time.
This is achieved without changing the system date by using Git's environment variables.
For example:
  gitcury commit with-date --datetime "2023-01-01T12:00:00" --all
or
  gitcury commit with-date --datetime "2023-01-01T12:00:00" --root my-folder`,
	Run: func(cmd *cobra.Command, args []string) {
		if commitDateTime == "" {
			utils.Error("You must specify the --datetime flag in the format 'YYYY-MM-DDTHH:MM:SS'.")
			return
		}

		// Validate the date-time format
		_, err := time.Parse("2006-01-02T15:04:05", commitDateTime)
		if err != nil {
			utils.Error("Invalid date-time format. Use 'YYYY-MM-DDTHH:MM:SS'.")
			return
		}

		// Set Git's environment variables
		env := append(os.Environ(),
			"GIT_AUTHOR_DATE="+commitDateTime,
			"GIT_COMMITTER_DATE="+commitDateTime,
		)

		utils.Info("Setting commit date and time to: " + commitDateTime)

		// Execute the commit logic
		var commitErr error
		if commitAllFlag {
			commitErr = core.CommitAllRoots(env)
		} else if folderName != "" {
			commitErr = core.CommitOneRoot(folderName, env)
		} else {
			utils.Error("You must specify either --all or --root flag.")
			return
		}

		if commitErr != nil {
			utils.Error("Failed to commit: " + commitErr.Error())
			return
		}
		utils.Info("Successfully committed files with the specified date and time.")
	},
}

func init() {
	// Add flags to the with-date subcommand
	withDateCmd.Flags().StringVarP(&commitDateTime, "datetime", "d", "", "Specify the commit date and time in 'YYYY-MM-DDTHH:MM:SS' format")
	withDateCmd.Flags().BoolVarP(&commitAllFlag, "all", "a", false, "Commit all files with generated messages")
	withDateCmd.Flags().StringVarP(&folderName, "root", "r", "", "Commit files in the specified root folder with generated messages")

	// Add the with-date subcommand to the commit command
	commitCmd.AddCommand(withDateCmd)

	commitCmd.Flags().BoolVarP(&commitAllFlag, "all", "a", false, "Commit all files with generated messages")
	commitCmd.Flags().StringVarP(&folderName, "root", "r", "", "Commit files in the specified root folder with generated messages")
	// Add the commit command to the root command
	rootCmd.AddCommand(commitCmd)
}
