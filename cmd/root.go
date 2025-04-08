package cmd

import (
	"GitCury/utils"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitcury",
	Short: "⚡ GitCury - The neural interface for Git",
	Long: `
██████╗ ██╗████████╗ ██████╗██╗   ██╗██████╗ ██╗   ██╗
██╔════╝ ██║╚══██╔══╝██╔════╝██║   ██║██╔══██╗╚██╗ ██╔╝
██║  ███╗██║   ██║   ██║     ██║   ██║██████╔╝ ╚████╔╝ 
██║   ██║██║   ██║   ██║     ██║   ██║██╔══██╗  ╚██╔╝  
╚██████╔╝██║   ██║   ╚██████╗╚██████╔╝██║  ██║   ██║   
 ╚═════╝ ╚═╝   ╚═╝    ╚═════╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   
                                                       
>> NEURAL GIT INTERFACE v1.0.0 <<

Automated Neural Network-Based Git Operations:
• Neural commit message generation through Gemini API
• Multi-repository simulation architecture
• Advanced operational parameters via config protocol
• Quantum state manipulation of Git repositories

[SYSTEM]: Connection established. All subsystems online.
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		utils.Error("Error executing command: " + err.Error())
		os.Exit(1)
	}
}
