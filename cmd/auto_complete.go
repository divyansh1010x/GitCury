package cmd

// import (
// 	"github.com/lakshyajain-0291/gitcury/utils"
// 	"os"

// 	"github.com/spf13/cobra"
// )

// var spectrumCmd = &cobra.Command{
// 	Use:   "completion",
// 	Short: "🌌 Generate shell auto-completion scripts",
// 	Long: `
// ╔══════════════════════════════════════════════════════════╗
// ║                  SPECTRUM: COMPLETION MODULE             ║
// ╚══════════════════════════════════════════════════════════╝

// [INITIATING]: The Spectrum Protocol—enhancing your CLI experience.

// Supported Shells:
// • Bash:
// 	source <(gitcury spectrum bash)
// • Zsh:
// 	source <(gitcury spectrum zsh)
// • Fish:
// 	gitcury spectrum fish | source
// • PowerShell:
// 	gitcury spectrum powershell | Out-String | Invoke-Expression

// [NOTICE]: Ensure the generated script is sourced in your shell configuration file.
// `,
// 	DisableFlagsInUseLine: true,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		utils.Info("[SPECTRUM]: Displaying help for shell auto-completion.")
// 		cmd.Help()
// 	},
// }

// func init() {
// 	rootCmd.AddCommand(spectrumCmd)

// 	spectrumCmd.AddCommand(&cobra.Command{
// 		Use:   "bash",
// 		Short: "⚡ Generate Bash auto-completion script",
// 		Run: func(cmd *cobra.Command, args []string) {
// 			utils.Info("[SPECTRUM]: Generating Bash auto-completion script...")
// 			err := rootCmd.GenBashCompletion(os.Stdout)
// 			if err != nil {
// 				utils.Error("[SPECTRUM]: Failed to generate Bash auto-completion script - " + err.Error())
// 				return
// 			}
// 			utils.Success("[SPECTRUM]: Bash auto-completion script generated successfully.")
// 		},
// 	})

// 	spectrumCmd.AddCommand(&cobra.Command{
// 		Use:   "zsh",
// 		Short: "⚡ Generate Zsh auto-completion script",
// 		Run: func(cmd *cobra.Command, args []string) {
// 			utils.Info("[SPECTRUM]: Generating Zsh auto-completion script...")
// 			err := rootCmd.GenZshCompletion(os.Stdout)
// 			if err != nil {
// 				utils.Error("[SPECTRUM]: Failed to generate Zsh auto-completion script - " + err.Error())
// 				return
// 			}
// 			utils.Success("[SPECTRUM]: Zsh auto-completion script generated successfully.")
// 		},
// 	})

// 	spectrumCmd.AddCommand(&cobra.Command{
// 		Use:   "fish",
// 		Short: "⚡ Generate Fish auto-completion script",
// 		Run: func(cmd *cobra.Command, args []string) {
// 			utils.Info("[SPECTRUM]: Generating Fish auto-completion script...")
// 			err := rootCmd.GenFishCompletion(os.Stdout, true)
// 			if err != nil {
// 				utils.Error("[SPECTRUM]: Failed to generate Fish auto-completion script - " + err.Error())
// 				return
// 			}
// 			utils.Success("[SPECTRUM]: Fish auto-completion script generated successfully.")
// 		},
// 	})

// 	spectrumCmd.AddCommand(&cobra.Command{
// 		Use:   "powershell",
// 		Short: "⚡ Generate PowerShell auto-completion script",
// 		Run: func(cmd *cobra.Command, args []string) {
// 			utils.Info("[SPECTRUM]: Generating PowerShell auto-completion script...")
// 			err := rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
// 			if err != nil {
// 				utils.Error("[SPECTRUM]: Failed to generate PowerShell auto-completion script - " + err.Error())
// 				return
// 			}
// 			utils.Success("[SPECTRUM]: PowerShell auto-completion script generated successfully.")
// 		},
// 	})
// }
