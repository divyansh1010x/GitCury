package cmd

import (
	"GitCury/core"
	"GitCury/output"
	"GitCury/utils"
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	cascadeAll      bool
	cascadeRoot     string
	cascadeNumFiles int
)

var cascadeCmd = &cobra.Command{
	Use:   "cascade",
	Short: "Trigger a complete neural git transformation sequence",
	Long: `
╔══════════════════════════════════════════════════════════╗
║           CASCADE: QUANTUM TRANSFORMATION CHAIN          ║
╚══════════════════════════════════════════════════════════╝

[INITIATING]: The Cascade Protocol—complete neural git transformation sequence.

This neural sequence executes an autonomous chain reaction:
• 🧠 Neural pattern analysis of quantum state differentials
• 🔄 Interactive confirmation of pattern recognition results
• 🔒 Sealing of approved quantum state changes
• 🌐 Neural transmission to remote nodes

The cascade creates an optimal path through the entire git transformation cycle
with minimal human intervention required.

Examples:
• Full system cascade:
    gitcury cascade --all

• Localized cascade:
    gitcury cascade --root /path/to/folder

[NOTICE]: Prepare for sequential protocol execution with confirmation checkpoints.
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validation checks
		if !cascadeAll && cascadeRoot == "" {
			utils.Error("[CASCADE.FAIL]: ❗ You must specify either --all or --root flag.")
			return
		}

		// PHASE 1: Generate commit messages (similar to genesis)
		utils.Info("[CASCADE]: 🧠 PHASE 1 - Initiating neural pattern analysis...")

		var err error
		if cascadeAll {
			err = core.GetAllMsgs(cascadeNumFiles)
		} else {
			err = core.GetMsgsForRootFolder(cascadeRoot, cascadeNumFiles)
		}

		if err != nil {
			utils.Error("[CASCADE.FAIL]: ❌ Neural pattern analysis failed: " + err.Error())
			return
		}

		// PHASE 2: Display results and get confirmation
		utils.Success("[CASCADE]: ✨ Neural patterns generated. Displaying quantum state analysis:")

		allOutput := output.GetAll()
		utils.Print(utils.ToJSON(allOutput))

		if len(allOutput.Folders) == 0 {
			utils.Error("[CASCADE.FAIL]: ⚠️ No changes detected in quantum state.")
			return
		}

		// Ask for confirmation to proceed with commit
		utils.Info("[CASCADE]: 🔄 PHASE 2 - Awaiting approval for quantum state sealing...")

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\n" + utils.BlackBg + utils.Green + "[CASCADE.PROMPT]: Proceed with sealing these quantum states? (y/n): " + utils.Reset + " ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			utils.Warning("[CASCADE.ABORT]: 🛑 Quantum state sealing aborted by user.")
			return
		}

		// PHASE 3: Commit changes (similar to seal)
		utils.Info("[CASCADE]: 🔒 PHASE 3 - Initiating quantum state sealing...")

		if cascadeAll {
			err = core.CommitAllRoots()
		} else {
			err = core.CommitOneRoot(cascadeRoot)
		}

		if err != nil {
			utils.Error("[CASCADE.FAIL]: ❌ Quantum state sealing failed: " + err.Error())
			return
		}

		utils.Success("[CASCADE]: ✅ Quantum states successfully sealed.")

		// PHASE 4: Ask about pushing changes
		utils.Info("[CASCADE]: 🌐 PHASE 4 - Preparing for neural transmission...")

		fmt.Print("\n" + utils.BlackBg + utils.Cyan + "[CASCADE.PROMPT]: Transmit sealed states to remote node? (y/n): " + utils.Reset + " ")
		response, _ = reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			utils.Success("[CASCADE]: ✅ Cascade protocol completed. Neural transmission skipped.")
			return
		}

		// Get branch name
		fmt.Print("\n" + utils.BlackBg + utils.Cyan + "[CASCADE.PROMPT]: Specify transmission vector (branch name) [default: main]: " + utils.Reset + " ")
		branchName, _ := reader.ReadString('\n')
		branchName = strings.TrimSpace(branchName)

		if branchName == "" {
			branchName = "main"
			utils.Info("[CASCADE]: Using default transmission vector: main")
		}

		// PHASE 5: Push changes (similar to deploy)
		utils.Info("[CASCADE]: 📡 PHASE 5 - Initiating neural transmission to vector: " + branchName)

		if cascadeAll {
			err = core.PushAllRoots(branchName)
		} else {
			err = core.PushOneRoot(cascadeRoot, branchName)
		}

		if err != nil {
			utils.Error("[CASCADE.FAIL]: ❌ Neural transmission failed: " + err.Error())
			return
		}

		utils.Success("[CASCADE.COMPLETE]: 🎉 Cascade protocol executed successfully. All phases completed.")
	},
}

func init() {
	cascadeCmd.Flags().BoolVarP(&cascadeAll, "all", "a", false, "🌐 Execute cascade across all root folders")
	cascadeCmd.Flags().StringVarP(&cascadeRoot, "root", "r", "", "📂 Target a specific root folder for cascade execution")
	cascadeCmd.Flags().IntVarP(&cascadeNumFiles, "num", "n", 0, "🔢 Maximum number of files to process per folder")

	rootCmd.AddCommand(cascadeCmd)
}
