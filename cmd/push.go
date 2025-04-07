package cmd

import (
	"GitCury/core"
	"GitCury/utils"

	"github.com/spf13/cobra"
)

var (
	folder      string
	pushAllFlag bool
	branchName  string
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push changes to remote repository",
	Long: `
This command pushes the committed changes to the remote repository.
It requires that the files have been committed using the 'commit' command.
For example:
  gitcury push --all --branch main
or
  gitcury push --root my-folder --branch main`,
	Run: func(cmd *cobra.Command, args []string) {
		if pushAllFlag {
			err := core.PushAllRoots(branchName)
			if err != nil {
				utils.Error(err.Error())
				return
			}
			utils.Info("Successfully pushed all changes to remote repository.")
		} else if folder != "" {
			err := core.PushOneRoot(folder, branchName)
			if err != nil {
				utils.Error(err.Error())
				return
			}
			utils.Info("Successfully pushed changes in the specified folder to remote repository.")
		} else {
			utils.Error("You must specify either --all or --root flag.")
		}
	},
}

func init() {
	pushCmd.Flags().BoolVarP(&pushAllFlag, "all", "a", false, "Push all changes to remote repository")
	pushCmd.Flags().StringVarP(&folder, "root", "r", "", "Push changes in the specified root folder to remote repository")
	pushCmd.Flags().StringVarP(&branchName, "branch", "b", "", "Specify the branch to push to (default: current branch)")
	rootCmd.AddCommand(pushCmd)
}
