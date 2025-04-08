// package cmd

// import (
// 	"GitCury/config"
// 	"GitCury/output"
// 	"GitCury/utils"
// 	"os"
// 	"os/exec"
// 	"strings"

// 	"github.com/spf13/cobra"
// )

// var (
// 	deleteFlag bool
// 	logFlag    bool
// 	editFlag   bool
// )

// var outputCmd = &cobra.Command{
// 	Use:   "output",
// 	Short: "Generated messages output and their related cmds for gitcury",
// 	Long: `
// The 'output' command provides options to display and manage the generated commit messages and their related commands for GitCury.
// You can use this command to view the generated commit messages and their associated commands in a structured format.
// For example:
//   gitcury output --log
// or
//   gitcury output --edit
// `,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if deleteFlag {
// 			output.Clear()
// 			utils.Info("Successfully deleted all generated commit messages and their related commands.")
// 			return
// 		}
// 		if logFlag {
// 			// Call the function to display logs
// 			output := output.GetAll()

// 			utils.Info("Successfully retrieved all generated commit messages and their related commands.")
// 			utils.Print(utils.ToJSON(output))
// 		} else if editFlag {
// 			editor, ok := config.Get("editor").(string)
// 			if !ok || editor == "" {
// 				editor = os.Getenv("EDITOR")
// 				if editor == "" {
// 					editor = "nano" // Default to nano if no editor is set
// 				}
// 			}

// 			outputFile, ok := config.Get("output_file_path").(string)
// 			if !ok || outputFile == "" {

// 				outputFile = os.Getenv("HOME") + "/output.json" // Default to the user's home directory for a more reliable path
// 			} else {
// 				outputFile = strings.TrimSpace(outputFile)
// 			}
// 			if _, err := os.Stat(outputFile); os.IsNotExist(err) {
// 				utils.Error("Output file does not exist. Please generate the output first.")
// 				return
// 			}

// 			cmd := exec.Command(editor, outputFile)
// 			cmd.Stdin = os.Stdin
// 			cmd.Stdout = os.Stdout
// 			cmd.Stderr = os.Stderr

// 			if err := cmd.Run(); err != nil {
// 				utils.Error("Failed to open the editor: " + err.Error())
// 				return
// 			}

// 			utils.Info("Successfully opened and edited the output file.")
// 		} else {
// 			cmd.Help()
// 		}
// 	},
// }

// func init() {
// 	outputCmd.Flags().BoolVarP(&deleteFlag, "delete", "d", false, "Delete all generated commit messages and their related commands")
// 	outputCmd.Flags().BoolVarP(&logFlag, "log", "l", false, "Display all generated commit messages and their related commands")
// 	outputCmd.Flags().BoolVarP(&editFlag, "edit", "e", false, "Edit the output file with the configured editor")
// 	rootCmd.AddCommand(outputCmd)
// }

package cmd

import (
	"GitCury/config"
	"GitCury/output"
	"GitCury/utils"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	purgeFlag bool
	viewFlag  bool
	editFlag  bool
)

var traceCmd = &cobra.Command{
	Use:   "trace",
	Short: "Manage and inspect generated commit traces",
	Long: `
╔══════════════════════════════════════════════════╗
║              TRACE: COMMIT TRACE MANAGER         ║
╚══════════════════════════════════════════════════╝

[INITIATING]: The Trace Protocol—interact with and manage generated commit traces.

Operational Modes:
• --view : Display all generated commit traces.
• --edit : Open the trace file for manual editing.
• --purge : Obliterate all commit traces and related commands.

Examples:
• View all traces:
	gitcury trace --view

• Edit the trace file:
	gitcury trace --edit

• Purge all traces:
	gitcury trace --purge

[NOTICE]: Ensure traces are generated before attempting to view or edit.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if purgeFlag {
			output.Clear()
			utils.Success("[TRACE.PURGE]: 🌌 All commit traces and commands have been obliterated.")
			return
		}
		if viewFlag {
			commitTraces := output.GetAll()
			utils.Success("[TRACE.VIEW]: 🖥️ Displaying generated commit traces:")
			utils.Print(utils.ToJSON(commitTraces))
		} else if editFlag {
			editor := resolveEditor()
			traceFile := resolveTraceFile()

			if _, err := os.Stat(traceFile); os.IsNotExist(err) {
				utils.Error("[TRACE.EDIT]: 🚨 Trace file not found. Ensure traces have been generated before editing.")
				return
			}

			cmd := exec.Command(editor, traceFile)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				utils.Error("[TRACE.EDIT]: ⚠️ Failed to open the editor: " + err.Error())
				return
			}

			utils.Success("[TRACE.EDIT]: ✨ Successfully edited the trace file.")
		} else {
			cmd.Help()
		}
	},
}

func resolveEditor() string {
	editor, ok := config.Get("editor").(string)
	if !ok || editor == "" {
		editor = os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano" // Default to nano if not set
		}
	}
	return editor
}

func resolveTraceFile() string {
	traceFile, ok := config.Get("trace_file_path").(string)
	if !ok || traceFile == "" {
		traceFile = os.Getenv("HOME") + "/traces.json" // Default to the user's home directory
	} else {
		traceFile = strings.TrimSpace(traceFile)
	}
	return traceFile
}

func init() {
	traceCmd.Flags().BoolVarP(&purgeFlag, "purge", "p", false, "🌌 Purge all commit traces and related commands")
	traceCmd.Flags().BoolVarP(&viewFlag, "view", "v", false, "🖥️ View all generated commit traces and their related commands")
	traceCmd.Flags().BoolVarP(&editFlag, "edit", "e", false, "✏️ Edit the trace file using the configured editor")
	rootCmd.AddCommand(traceCmd)
}
