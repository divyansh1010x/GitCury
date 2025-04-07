package cmd

import (
	"GitCury/core"
	"GitCury/output"
	"GitCury/utils"

	"github.com/spf13/cobra"
)

var (
	numFiles       int
	rootFolderName string
	allFlag        bool
)

var getMsgsCmd = &cobra.Command{
	Use:   "getmsgs",
	Short: "Generate commit messages for changed files",
	Long: `
This command generates commit messages for changed files. 
You can use the --all flag to generate messages for all changed files in all configured root folders, 
or use the --root flag to specify a particular root folder.
For example:

	getmsgs --all --num 5

or

	getmsgs --root my-folder --num 5
`,
	Run: func(cmd *cobra.Command, args []string) {
		if allFlag {
			err := core.GetAllMsgs(numFiles)
			if err != nil {
				utils.Error(err.Error())
				return
			}

			allOutput := output.GetAll()
			utils.Info("Successfully generated commit messages for all changed files in all configured root folders.")

			utils.Print(utils.ToJSON(allOutput))
		} else if rootFolderName != "" {
			err := core.GetMsgsForRootFolder(rootFolderName, numFiles)
			if err != nil {
				utils.Error(err.Error())
				return
			}

			rootFolder := output.GetFolder(rootFolderName)
			if len(rootFolder.Files) == 0 {
				utils.Error("No files to commit in the specified root folder.")
				return
			}

			utils.Info("Successfully generated commit messages for changed files in the specified root folder.")
			utils.Print(utils.ToJSON(rootFolder))
		} else {
			utils.Error("You must specify either --all or --root flag.")
		}
	},
}

func init() {

	getMsgsCmd.Flags().IntVarP(&numFiles, "num", "n", 0, "Number of files to commit per folder (overrides config)")
	getMsgsCmd.Flags().StringVarP(&rootFolderName, "root", "r", "", "Root folder to commit in (overrides config)")
	getMsgsCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Generate commit messages for all changed files in all configured root folders")

	rootCmd.AddCommand(getMsgsCmd)
}
