package cmd

import (
	"GitCury/core"
	"GitCury/utils"

	"github.com/spf13/cobra"
)

var (
	commitAllFlag bool
	folderName    string
)
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit files with generated messages",
	Long: `
This command commits files with generated messages.
You can use the --all flag to commit all files with generated messages,
or use the --root flag to specify a particular root folder.
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

func init() {
	commitCmd.Flags().BoolVarP(&commitAllFlag, "all", "a", false, "Commit all files with generated messages")
	commitCmd.Flags().StringVarP(&folderName, "root", "r", "", "Commit files in the specified root folder with generated messages")
	rootCmd.AddCommand(commitCmd)
}
