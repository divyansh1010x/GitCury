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
}

func Execute() {
	config.LoadConfig()
	ReampAlias(rootCmd)

	// Override the default help template to include aliases
	cobra.AddTemplateFunc("aliasList", func(cmd *cobra.Command) string {
		if len(cmd.Aliases) > 0 {
			return cmd.NameAndAliases()
		}
		return ""
	})

	rootCmd.SetHelpTemplate(`{{.UseLine}}
{{.Long}}

{{if .HasAvailableSubCommands}}Available Commands:
  {{printf "\n  %-15s %-20s %s" "Name" "Aliases" "Description"}}
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{printf "%-15s %-20s %s" .Name (aliasList .) .Short}}{{end}}{{end}}{{end}}

{{if .HasAvailableLocalFlags}}Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}

{{if .HasAvailableInheritedFlags}}Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}

{{if .HasHelpSubCommands}}Additional help topics:
{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}

Use "{{.CommandPath}} [command] --help" for more information about a command.
`)

	if err := rootCmd.Execute(); err != nil {
		utils.Error("Error executing command: " + err.Error())
		os.Exit(1)
	}
}
