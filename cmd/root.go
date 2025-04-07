package cmd

import (
	"GitCury/utils"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitcury",
	Short: "🚀 GitCury CLI tool for automating git commands and generating commit messages",
	Long: `
🌟 GitCury CLI 🌟

GitCury automates Git commit message generation using the Gemini API.
It supports operations like:

  - Generating commit messages for changed files
  - Committing and pushing changes
  - Scoping operations to configured root folders

Simplify your Git workflow and boost productivity! 💻✨
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		utils.Error("Error executing command: " + err.Error())
		os.Exit(1)
	}
}
