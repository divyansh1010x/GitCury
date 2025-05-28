// package cmd

// import (
// 	"GitCury/utils"
// 	"os"

// 	"github.com/spf13/cobra"
// )

// var rootCmd = &cobra.Command{
// 	Use:   "gitcury",
// 	Short: "⚡ GitCury - The neural interface for Git",
// 	Long: `
// ██████╗ ██╗████████╗ ██████╗██╗   ██╗██████╗ ██╗   ██╗
// ██╔════╝ ██║╚══██╔══╝██╔════╝██║   ██║██╔══██╗╚██╗ ██╔╝
// ██║  ███╗██║   ██║   ██║     ██║   ██║██████╔╝ ╚████╔╝
// ██║   ██║██║   ██║   ██║     ██║   ██║██╔══██╗  ╚██╔╝
// ╚██████╔╝██║   ██║   ╚██████╗╚██████╔╝██║  ██║   ██║
//  ╚═════╝ ╚═╝   ╚═╝    ╚═════╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝

// >> NEURAL GIT INTERFACE v1.0.0 <<

// Automated Neural Network-Based Git Operations:
// • Neural commit message generation through Gemini API
// • Multi-repository simulation architecture
// • Advanced operational parameters via config protocol
// • Quantum state manipulation of Git repositories

// [SYSTEM]: Connection established. All subsystems online.
// `,
// }

// func Execute() {
// 	if err := rootCmd.Execute(); err != nil {
// 		utils.Error("Error executing command: " + err.Error())
// 		os.Exit(1)
// 	}
// }

package cmd

import (
	"GitCury/config"
	"GitCury/utils"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitcury",
	Short: "⚡ GitCury - Your AI-powered Git assistant",
	Long: `
██████╗ ██╗████████╗ ██████╗██╗   ██╗██████╗ ██╗   ██╗
██╔════╝ ██║╚══██╔══╝██╔════╝██║   ██║██╔══██╗╚██╗ ██╔╝
██║  ███╗██║   ██║   ██║     ██║   ██║██████╔╝ ╚████╔╝ 
██║   ██║██║   ██║   ██║     ██║   ██║██╔══██╗  ╚██╔╝  
╚██████╔╝██║   ██║   ╚██████╗╚██████╔╝██║  ██║   ██║   
 ╚═════╝ ╚═╝   ╚═╝    ╚═════╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   
													   
>> NEURAL GIT INTERFACE v1.0.0 <<

Simplify your Git workflow with AI:
• Smart commit message generation
• Manage multiple repositories effortlessly
• Advanced configuration options
• Streamline Git operations with ease

[SYSTEM]: Ready to assist. All systems operational.
`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is specified, show help
		cmd.Help()
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		// Print stats if enabled
		if utils.IsStatsEnabled() {
			utils.PrintStats()
		}
	},
}

func Execute() {
	// Ensure config is loaded
	err := utils.SafeExecute("LoadConfig", func() error {
		return config.LoadConfig()
	})

	if err != nil {
		utils.Error("Failed to load configuration: " + err.Error())
		utils.Info("Falling back to default configuration")
		// Continue with defaults
	}

	// Add a version flag to the root command
	rootCmd.PersistentFlags().BoolP("version", "V", false, "Print the version number of GitCury")

	// Add common flags
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Minimize output, only show errors")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug output")
	rootCmd.PersistentFlags().BoolP("stats", "s", false, "Show command execution statistics (completion time, progress, etc.)")

	// Add a hook to handle flags before executing the command
	cobra.OnInitialize(func() {
		// Check if version flag is set
		versionFlag, _ := rootCmd.Flags().GetBool("version")
		if versionFlag {
			fmt.Println("GitCury version 1.0.0")
			os.Exit(0)
		}

		// Handle quiet flag
		quietFlag, _ := rootCmd.Flags().GetBool("quiet")
		if quietFlag {
			utils.SetLogLevel("error")
		}

		// Handle debug flag
		debugFlag, _ := rootCmd.Flags().GetBool("debug")
		if debugFlag {
			utils.SetLogLevel("debug")
		}

		// Handle stats flag
		statsFlag, _ := rootCmd.Flags().GetBool("stats")
		if statsFlag {
			utils.EnableStats()
			// Record the start of the command in the stats
			commandName := rootCmd.Name()
			if len(os.Args) > 1 {
				commandName = os.Args[1] // Use the subcommand name for better tracking
			}
			utils.StartOperation("Command:" + commandName)
		}
	})

	// Remap aliases
	ReampAlias(rootCmd)

	// Override the default help template to include aliases and better formatting
	cobra.AddTemplateFunc("aliasList", func(cmd *cobra.Command) string {
		if len(cmd.Aliases) > 0 {
			return cmd.NameAndAliases()
		}
		return ""
	})

	rootCmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}{{end}}

╔══════════════════════════════════════════════════════════════════════════════╗
║                             AVAILABLE COMMANDS                               ║
╚══════════════════════════════════════════════════════════════════════════════╝

{{if .HasAvailableSubCommands}}{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}  {{rpad .Name 15}} {{rpad (aliasList .) 20}} {{.Short}}
{{end}}{{end}}{{end}}

╔══════════════════════════════════════════════════════════════════════════════╗
║                              COMMAND FLAGS                                   ║
╚══════════════════════════════════════════════════════════════════════════════╝

{{if .HasAvailableLocalFlags}}{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}

{{if .HasAvailableInheritedFlags}}╔══════════════════════════════════════════════════════════════════════════════╗
║                              GLOBAL FLAGS                                    ║
╚══════════════════════════════════════════════════════════════════════════════╝

{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}

Use "{{.CommandPath}} [command] --help" for more information about a command.
For complete documentation, visit: https://github.com/lakshyajain-0291/GitCury
`)

	// Add PostRun function to all commands to display stats
	addStatsPostRunToAllCommands(rootCmd)

	// Use custom error handling with user-friendly messages
	if err := rootCmd.Execute(); err != nil {
		// Convert error to user-friendly message
		utils.Error(utils.ToUserFriendlyMessage(err))
		os.Exit(1)
	}
}

// addStatsPostRunToAllCommands recursively adds a stats PostRun function to all commands
func addStatsPostRunToAllCommands(cmd *cobra.Command) {
	// Add stats PostRun to this command
	utils.AddStatsPostRunToCommand(cmd)
	
	// Recursively add to all subcommands
	for _, subCmd := range cmd.Commands() {
		addStatsPostRunToAllCommands(subCmd)
	}
}
