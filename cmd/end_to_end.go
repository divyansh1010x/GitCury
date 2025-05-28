// package cmd

// import (
// 	"GitCury/config"
// 	"GitCury/core"
// 	"GitCury/output"
// 	"GitCury/utils"
// 	"bufio"
// 	"fmt"
// 	"os"
// 	"strings"

// 	"github.com/spf13/cobra"
// )

// var (
// 	cascadeAll      bool
// 	cascadeRoot     string
// 	cascadeNumFiles int
// )

// var cascadeCmd = &cobra.Command{
// 	Use:   "cascade",
// 	Short: "Trigger a complete neural git transformation sequence",
// 	Long: `
// ╔══════════════════════════════════════════════════════════╗
// ║           "+ config.Aliases.Boom +": QUANTUM TRANSFORMATION CHAIN          ║
// ╚══════════════════════════════════════════════════════════╝

// [INITIATING]: The Cascade Protocol—complete neural git transformation sequence.

// This neural sequence executes an autonomous chain reaction:
// • 🧠 Neural pattern analysis of quantum state differentials
// • 🔄 Interactive confirmation of pattern recognition results
// • 🔒 Sealing of approved quantum state changes
// • 🌐 Neural transmission to remote nodes

// The cascade creates an optimal path through the entire git transformation cycle
// with minimal human intervention required.

// Examples:
// • Full system cascade:
//     gitcury cascade --all

// • Localized cascade:
//     gitcury cascade --root /path/to/folder

// [NOTICE]: Prepare for sequential protocol execution with confirmation checkpoints.
// `,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		// Validation checks
// 		if !cascadeAll && cascadeRoot == "" {
// 			utils.Error("[" + config.Aliases.Boom + ".FAIL]: ❗ You must specify either --all or --root flag.")
// 			return
// 		}

// 		// PHASE 1: Generate commit messages (similar to genesis)
// 		utils.Info("[" + config.Aliases.Boom + "]: 🧠 PHASE 1 - Initiating neural pattern analysis...")

// 		var err error
// 		if cascadeAll {
// 			err = core.GetAllMsgs(cascadeNumFiles)
// 		} else {
// 			err = core.GetMsgsForRootFolder(cascadeRoot, cascadeNumFiles)
// 		}

// 		if err != nil {
// 			utils.Error("[" + config.Aliases.Boom + ".FAIL]: ❌ Neural pattern analysis failed: " + err.Error())
// 			return
// 		}

// 		// PHASE 2: Display results and get confirmation
// 		utils.Success("[" + config.Aliases.Boom + "]: ✨ Neural patterns generated. Displaying quantum state analysis:")

// 		allOutput := output.GetAll()
// 		utils.Print(utils.ToJSON(allOutput))

// 		if len(allOutput.Folders) == 0 {
// 			utils.Error("[" + config.Aliases.Boom + ".FAIL]: ⚠️ No changes detected in quantum state.")
// 			return
// 		}

// 		// Ask for confirmation to proceed with commit
// 		utils.Info("[" + config.Aliases.Boom + "]: 🔄 PHASE 2 - Awaiting approval for quantum state sealing...")

// 		reader := bufio.NewReader(os.Stdin)
// 		fmt.Print("\n" + utils.BlackBg + utils.Green + "[" + config.Aliases.Boom + ".PROMPT]: Proceed with sealing these quantum states? (y/n): " + utils.Reset + " ")
// 		response, _ := reader.ReadString('\n')
// 		response = strings.TrimSpace(strings.ToLower(response))

// 		if response != "y" && response != "yes" {
// 			utils.Warning("[" + config.Aliases.Boom + ".ABORT]: 🛑 Quantum state sealing aborted by user.")
// 			return
// 		}

// 		// PHASE 3: Commit changes (similar to seal)
// 		utils.Info("[" + config.Aliases.Boom + "]: 🔒 PHASE 3 - Initiating quantum state sealing...")

// 		if cascadeAll {
// 			err = core.CommitAllRoots()
// 		} else {
// 			err = core.CommitOneRoot(cascadeRoot)
// 		}

// 		if err != nil {
// 			utils.Error("[" + config.Aliases.Boom + ".FAIL]: ❌ Quantum state sealing failed: " + err.Error())
// 			return
// 		}

// 		utils.Success("[" + config.Aliases.Boom + "]: ✅ Quantum states successfully sealed.")

// 		// PHASE 4: Ask about pushing changes
// 		utils.Info("[" + config.Aliases.Boom + "]: 🌐 PHASE 4 - Preparing for neural transmission...")

// 		fmt.Print("\n" + utils.BlackBg + utils.Cyan + "[" + config.Aliases.Boom + ".PROMPT]: Transmit sealed states to remote node? (y/n): " + utils.Reset + " ")
// 		response, _ = reader.ReadString('\n')
// 		response = strings.TrimSpace(strings.ToLower(response))

// 		if response != "y" && response != "yes" {
// 			utils.Success("[" + config.Aliases.Boom + "]: ✅ Cascade protocol completed. Neural transmission skipped.")
// 			return
// 		}

// 		// Get branch name
// 		fmt.Print("\n" + utils.BlackBg + utils.Cyan + "[" + config.Aliases.Boom + ".PROMPT]: Specify transmission vector (branch name) [default: main]: " + utils.Reset + " ")
// 		branchName, _ := reader.ReadString('\n')
// 		branchName = strings.TrimSpace(branchName)

// 		if branchName == "" {
// 			branchName = "main"
// 			utils.Info("[" + config.Aliases.Boom + "]: Using default transmission vector: main")
// 		}

// 		// PHASE 5: Push changes (similar to deploy)
// 		utils.Info("[" + config.Aliases.Boom + "]: 📡 PHASE 5 - Initiating neural transmission to vector: " + branchName)

// 		if cascadeAll {
// 			err = core.PushAllRoots(branchName)
// 		} else {
// 			err = core.PushOneRoot(cascadeRoot, branchName)
// 		}

// 		if err != nil {
// 			utils.Error("[" + config.Aliases.Boom + ".FAIL]: ❌ Neural transmission failed: " + err.Error())
// 			return
// 		}

// 		utils.Success("[" + config.Aliases.Boom + ".COMPLETE]: 🎉 Cascade protocol executed successfully. All phases completed.")
// 	},
// }

// func init() {
// 	cascadeCmd.Flags().BoolVarP(&cascadeAll, "all", "a", false, "🌐 Execute cascade across all root folders")
// 	cascadeCmd.Flags().StringVarP(&cascadeRoot, "root", "r", "", "📂 Target a specific root folder for cascade execution")
// 	cascadeCmd.Flags().IntVarP(&cascadeNumFiles, "num", "n", 0, "🔢 Maximum number of files to process per folder")

// 	rootCmd.AddCommand(cascadeCmd)
// }

package cmd

import (
	"GitCury/config"
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

var boomCmd = &cobra.Command{
	Use:   "boom",
	Short: "Execute a complete git transformation sequence",
	Long: `
Execute a complete git transformation sequence.

Aliases:
• ` + config.Aliases.Boom + `

Options:
• --all : Execute boom across all root folders.
• --root <folder> : Target a specific root folder for boom execution.
• --num <number> : Maximum number of files to process per folder.

Examples:
• Full system boom:
	gitcury boom --all

• Localized boom:
	gitcury boom --root /path/to/folder
`,
	Run: func(cmd *cobra.Command, args []string) {
		if !cascadeAll && cascadeRoot == "" {
			utils.Error("Specify either --all or --root flag.")
			return
		}

		// Start overall boom operation tracking
		if utils.IsStatsEnabled() {
			utils.StartOperation("BoomWorkflow")
			// Set initial progress
			utils.UpdateOperationProgress("BoomWorkflow", 10.0)
		}

		utils.Info("Starting analysis...")

		// Track message generation
		if utils.IsStatsEnabled() {
			utils.StartOperation("BoomGenerateMessages")
			// Set initial progress
			utils.UpdateOperationProgress("BoomGenerateMessages", 10.0)
		}

		var err error
		if cascadeAll {
			err = core.GetAllMsgs(cascadeNumFiles)
		} else {
			err = core.GetMsgsForRootFolder(cascadeRoot, cascadeNumFiles)
		}

		if err != nil {
			if utils.IsStatsEnabled() {
				utils.FailOperation("BoomGenerateMessages", err.Error())
				utils.FailOperation("BoomWorkflow", "Message generation failed")
			}
			utils.Error("Analysis failed: " + err.Error())
			return
		}

		if utils.IsStatsEnabled() {
			utils.MarkOperationComplete("BoomGenerateMessages")
			utils.UpdateOperationProgress("BoomWorkflow", 30.0)
		}

		allOutput := output.GetAll()
		utils.Print(utils.ToJSON(allOutput))

		if len(allOutput.Folders) == 0 {
			if utils.IsStatsEnabled() {
				utils.FailOperation("BoomWorkflow", "No changes detected")
			}
			utils.Error("No changes detected.")
			return
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Proceed with committing changes? (y/n): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			if utils.IsStatsEnabled() {
				utils.FailOperation("BoomWorkflow", "User aborted operation")
			}
			utils.Warning("Operation aborted by user.")
			return
		}

		utils.Info("Committing changes...")

		// Track commit operation
		if utils.IsStatsEnabled() {
			utils.StartOperation("BoomCommit")
			// Set initial progress
			utils.UpdateOperationProgress("BoomCommit", 10.0)
		}

		if cascadeAll {
			err = core.CommitAllRoots()
		} else {
			err = core.CommitOneRoot(cascadeRoot)
		}

		if err != nil {
			if utils.IsStatsEnabled() {
				utils.FailOperation("BoomCommit", err.Error())
				utils.FailOperation("BoomWorkflow", "Commit failed")
			}
			utils.Error("Commit failed: " + err.Error())
			return
		}

		if utils.IsStatsEnabled() {
			utils.MarkOperationComplete("BoomCommit")
			utils.UpdateOperationProgress("BoomWorkflow", 60.0)
		}

		fmt.Print("Push changes to remote? (y/n): ")
		response, _ = reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			if utils.IsStatsEnabled() {
				utils.MarkOperationComplete("BoomWorkflow")
			}
			utils.Success("Operation completed. Push skipped.")
			return
		}

		fmt.Print("Specify branch [default: main]: ")
		branchName, _ := reader.ReadString('\n')
		branchName = strings.TrimSpace(branchName)
		if branchName == "" {
			branchName = "main"
		}

		utils.Info("Pushing to branch: " + branchName)

		// Track push operation
		if utils.IsStatsEnabled() {
			utils.StartOperation("BoomPush")
			// Set initial progress
			utils.UpdateOperationProgress("BoomPush", 10.0)
		}

		if cascadeAll {
			err = core.PushAllRoots(branchName)
		} else {
			err = core.PushOneRoot(cascadeRoot, branchName)
		}

		if err != nil {
			if utils.IsStatsEnabled() {
				utils.FailOperation("BoomPush", err.Error())
				utils.FailOperation("BoomWorkflow", "Push failed")
			}
			utils.Error("Push failed: " + err.Error())
			return
		}

		if utils.IsStatsEnabled() {
			utils.MarkOperationComplete("BoomPush")
			utils.MarkOperationComplete("BoomWorkflow")
		}

		utils.Success("Operation completed successfully.")
	},
}

func init() {
	boomCmd.Flags().BoolVarP(&cascadeAll, "all", "a", false, "Execute boom across all root folders")
	boomCmd.Flags().StringVarP(&cascadeRoot, "root", "r", "", "Target a specific root folder for boom execution")
	boomCmd.Flags().IntVarP(&cascadeNumFiles, "num", "n", 0, "Maximum number of files to process per folder")

	rootCmd.AddCommand(boomCmd)
}
