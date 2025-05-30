// package cmd

// import (
// 	"github.com/lakshyajain-0291/gitcury/config"
// 	"github.com/lakshyajain-0291/gitcury/utils"
// 	"os"
// 	"strings"

// 	"github.com/spf13/cobra"
// )

// var bootstrapCmd = &cobra.Command{
// 	Use:   "bootstrap",
// 	Short: "Bootstrap GitCury with essential configurations and shell integrations",
// 	Long: `
// ╔══════════════════════════════════════════════════════════╗
// ║                  BOOTSTRAP: SYSTEM INITIALIZER           ║
// ╚══════════════════════════════════════════════════════════╝

// [INITIATING]: The Bootstrap Protocol—setting up GitCury for optimal performance.

// Includes:
// • Generating essential configuration files.
// • Installing shell completion scripts for enhanced CLI experience.
// • Ensuring necessary directories and files are created.

// [NOTICE]: Ensure your shell environment is properly configured for seamless integration.
// `,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		// Step 1: Generate basic configuration
// 		utils.Info("[BOOTSTRAP]: Generating essential configuration...")
// 		config.LoadConfig()
// 		utils.Success("[BOOTSTRAP]: Configuration generated successfully.")

// 		configDir := config.Get("config_dir").(string)

// 		// Step 2: Install shell completion scripts
// 		utils.Info("[BOOTSTRAP]: Installing shell completion scripts...")
// 		shell := os.Getenv("SHELL")
// 		switch {
// 		case strings.Contains(shell, "bash"):
// 			err := rootCmd.GenBashCompletionFile(configDir + "/gitcury-completion.bash")
// 			if err == nil {
// 				utils.Success("[BOOTSTRAP]: Bash completion script installed at ~/.gitcury/gitcury-completion.bash.")
// 				utils.Info("[BOOTSTRAP]: Add 'source ~/.gitcury/gitcury-completion.bash' to your ~/.bashrc.")
// 			}
// 		case strings.Contains(shell, "zsh"):
// 			err := rootCmd.GenZshCompletionFile(configDir + "/gitcury-completion.zsh")
// 			if err == nil {
// 				utils.Success("[BOOTSTRAP]: Zsh completion script installed at ~/.gitcury/gitcury-completion.zsh.")
// 				utils.Info("[BOOTSTRAP]: Add 'source ~/.gitcury/gitcury-completion.zsh' to your ~/.zshrc.")
// 			}
// 		case strings.Contains(shell, "fish"):
// 			err := rootCmd.GenFishCompletionFile(configDir+"/completions/gitcury.fish", true)
// 			if err == nil {
// 				utils.Success("[BOOTSTRAP]: Fish completion script installed at ~/.gitcury/completions/gitcury.fish.")
// 			}
// 		default:
// 			utils.Error("[BOOTSTRAP]: Shell not recognized. Please use 'gitcury completion' to manually set up.")
// 		}

// 		utils.Success("[BOOTSTRAP]: Setup completed successfully!")
// 	},
// }

// func init() {
// 	rootCmd.AddCommand(bootstrapCmd)
// }

package cmd

import (
	"github.com/lakshyajain-0291/gitcury/config"
	"github.com/lakshyajain-0291/gitcury/utils"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup GitCury with configurations and shell integrations",
	Long: `
Setup GitCury with essential configurations and shell integrations.

Alias:
• ` + config.Aliases.Setup + `

Includes:
• Generating configuration files.
• Installing shell completion scripts.

Examples:
• Run setup:
	gitcury setup

[NOTICE]: Ensure your shell environment is properly configured for integration.
`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.Info("Setting up GitCury...")

		// Generate configuration
		if err := config.LoadConfig(); err != nil {
			utils.Error("Failed to load configuration: " + err.Error())
			return
		}
		utils.Success("Configuration generated.")

		configDir := config.Get("config_dir").(string)

		// Install shell completion scripts
		utils.Info("Installing shell completion scripts...")
		shell := os.Getenv("SHELL")
		switch {
		case strings.Contains(shell, "bash"):
			err := rootCmd.GenBashCompletionFile(configDir + "/gitcury-completion.bash")
			if err == nil {
				utils.Success("Bash completion script installed.")
				utils.Info("Add 'source ~/.gitcury/gitcury-completion.bash' to your ~/.bashrc.")
			}
		case strings.Contains(shell, "zsh"):
			err := rootCmd.GenZshCompletionFile(configDir + "/gitcury-completion.zsh")
			if err == nil {
				utils.Success("Zsh completion script installed.")
				utils.Info("Add 'source ~/.gitcury/gitcury-completion.zsh' to your ~/.zshrc.")
			}
		case strings.Contains(shell, "fish"):
			err := rootCmd.GenFishCompletionFile(configDir+"/completions/gitcury.fish", true)
			if err == nil {
				utils.Success("Fish completion script installed.")
			}
		default:
			utils.Error("Shell not recognized. Use 'gitcury completion' for manual setup.")
		}

		utils.Success("Setup completed!")
	},
}

func init() {
	utils.AddStatsPostRunToCommand(setupCmd)
	rootCmd.AddCommand(setupCmd)
}
