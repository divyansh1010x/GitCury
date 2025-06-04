// package cmd

// import (
// 	"github.com/lakshyajain-0291/gitcury/utils"
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
	"github.com/lakshyajain-0291/gitcury/config"
	"github.com/lakshyajain-0291/gitcury/utils"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version information
var (
	versionInfo struct {
		Version string
		Commit  string
		Date    string
	}
)

// SetVersionInfo sets the version information for use in commands
func SetVersionInfo(version, commit, date string) {
	versionInfo.Version = version
	versionInfo.Commit = commit
	versionInfo.Date = date
	
	// Update the Long description with dynamic version
	updateLongDescription()
}

// updateLongDescription updates the root command's Long description with current version
func updateLongDescription() {
	version := versionInfo.Version
	if version == "" {
		version = "dev"
	}
	
	rootCmd.Long = fmt.Sprintf(`
██████╗ ██╗████████╗ ██████╗██╗   ██╗██████╗ ██╗   ██╗
██╔════╝ ██║╚══██╔══╝██╔════╝██║   ██║██╔══██╗╚██╗ ██╔╝
██║  ███╗██║   ██║   ██║     ██║   ██║██████╔╝ ╚████╔╝ 
██║   ██║██║   ██║   ██║     ██║   ██║██╔══██╗  ╚██╔╝  
╚██████╔╝██║   ██║   ╚██████╗╚██████╔╝██║  ██║   ██║   
 ╚═════╝ ╚═╝   ╚═╝    ╚═════╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   

🤖 AI-POWERED GIT ASSISTANT v%s

Transform your Git workflow with intelligent automation:
  ⚡ AI-generated commit messages using Google Gemini
  🚀 Multi-repository management and batch operations  
  🎯 Smart file clustering and semantic grouping
  🔧 Advanced configuration and workflow customization
  📊 Performance tracking and statistics

🚀 QUICK START:
  1. gitcury config set --key GEMINI_API_KEY --value YOUR_KEY
  2. gitcury config set --key root_folders --value /path/to/repo
  3. gitcury getmsgs --all
  4. gitcury commit --all

💡 TIP: Use 'gitcury [command] --help' for detailed command information
📖 Documentation: https://github.com/lakshyajain-0291/gitcury`, version)
}

var rootCmd = &cobra.Command{
	Use:   "gitcury",
	Short: "⚡ AI-powered Git assistant for automated commit messages",
	Long: `
██████╗ ██╗████████╗ ██████╗██╗   ██╗██████╗ ██╗   ██╗
██╔════╝ ██║╚══██╔══╝██╔════╝██║   ██║██╔══██╗╚██╗ ██╔╝
██║  ███╗██║   ██║   ██║     ██║   ██║██████╔╝ ╚████╔╝ 
██║   ██║██║   ██║   ██║     ██║   ██║██╔══██╗  ╚██╔╝  
╚██████╔╝██║   ██║   ╚██████╗╚██████╔╝██║  ██║   ██║   
 ╚═════╝ ╚═╝   ╚═╝    ╚═════╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   

🤖 AI-POWERED GIT ASSISTANT

Transform your Git workflow with intelligent automation:
  ⚡ AI-generated commit messages using Google Gemini
  🚀 Multi-repository management and batch operations  
  🎯 Smart file clustering and semantic grouping
  🔧 Advanced configuration and workflow customization
  📊 Performance tracking and statistics

🚀 QUICK START:
  1. gitcury config set --key GEMINI_API_KEY --value YOUR_KEY
  2. gitcury config set --key root_folders --value /path/to/repo
  3. gitcury getmsgs --all
  4. gitcury commit --all

💡 TIP: Use 'gitcury [command] --help' for detailed command information
📖 Documentation: https://github.com/lakshyajain-0291/gitcury`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is specified, show help
		if err := cmd.Help(); err != nil {
			utils.Error("Failed to show help: " + err.Error())
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

	// Add version flags (both -v and -V for convenience)
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print the version number of GitCury")

	// Add common flags
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Minimize output, only show errors")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug output")
	rootCmd.PersistentFlags().BoolP("stats", "s", false, "Enable statistics tracking and display performance metrics")

	// Add a hook to handle flags before executing the command
	cobra.OnInitialize(func() {
		// Check if version flag is set
		versionFlag, _ := rootCmd.PersistentFlags().GetBool("version")
		if versionFlag {
			// Display dynamic version information
			version := versionInfo.Version
			commit := versionInfo.Commit
			date := versionInfo.Date
			
			if version == "" {
				version = "dev"
			}
			if commit == "" {
				commit = "unknown"
			}
			if date == "" {
				date = "unknown"
			}
			
			fmt.Printf("GitCury %s (commit %s, built on %s)\n", version, commit, date)
			os.Exit(0)
		}

		// Handle quiet flag
		quietFlag, _ := rootCmd.PersistentFlags().GetBool("quiet")
		if quietFlag {
			utils.SetLogLevel("error")
		}

		// Handle debug flag
		debugFlag, _ := rootCmd.PersistentFlags().GetBool("debug")
		if debugFlag {
			utils.SetLogLevel("debug")
		}

		// Handle stats flag
		statsFlag, _ := rootCmd.PersistentFlags().GetBool("stats")
		if statsFlag {
			utils.EnableStats()
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
║                           📋 AVAILABLE COMMANDS                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
{{if .HasAvailableSubCommands}}{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name 12}} {{rpad (aliasList .) 15}} {{.Short}}{{end}}{{end}}{{end}}

╔══════════════════════════════════════════════════════════════════════════════╗
║                            🚩 GLOBAL FLAGS                                   ║
╚══════════════════════════════════════════════════════════════════════════════╝
{{if .HasAvailableInheritedFlags}}{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
{{if .HasAvailableLocalFlags}}{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}

💡 Use "{{.CommandPath}} [command] --help" for detailed information about any command
📚 Documentation: https://github.com/lakshyajain-0291/gitcury
`)

	// Use custom error handling with user-friendly messages
	if err := rootCmd.Execute(); err != nil {
		// Convert error to user-friendly message
		utils.Error(utils.ToUserFriendlyMessage(err))
		os.Exit(1)
	}
}
