package cmd

import (
	"GitCury/config"
	"GitCury/utils"

	"github.com/spf13/cobra"
)

var Aliases = make(map[string]string)

func loadAliases() {
	if len(Aliases) > 0 {
		return // Already loaded
	}

	rawAliases := config.Get("aliases")
	if rawAliases == nil {
		// Use default aliases if none configured
		Aliases = map[string]string{
			"commit":  config.DefaultAliases.Commit,
			"push":    config.DefaultAliases.Push,
			"getmsgs": config.DefaultAliases.GetMsgs,
			"output":  config.DefaultAliases.Output,
			"config":  config.DefaultAliases.Config,
			"setup":   config.DefaultAliases.Setup,
			"boom":    config.DefaultAliases.Boom,
		}
		return
	}

	if aliasMap, ok := rawAliases.(map[string]interface{}); ok {
		convertedAliases := make(map[string]string)
		for key, value := range aliasMap {
			if strValue, ok := value.(string); ok {
				convertedAliases[key] = strValue
			}
		}
		Aliases = convertedAliases
	}
}

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage command aliases",
	Long: `
Manage command aliases for GitCury.

Options:
• --add <command> <alias> : Add a new alias for a command.
• --remove <alias> : Remove an existing alias.
• --list : List all existing aliases.

Examples:
• Add an alias for the 'commit' command:
	gitcury alias --add commit cm

• Remove an alias:
	gitcury alias --remove cm

• List all aliases:
	gitcury alias --list

[NOTICE]: Ensure aliases do not conflict with existing commands.
`,
	Run: func(cmd *cobra.Command, args []string) {
		loadAliases() // Ensure aliases are loaded
		if cmd.Flag("add").Changed {
			if len(args) != 2 {
				utils.Error("Invalid arguments. Usage: --add <command> <alias>")
				if err := cmd.Help(); err != nil {
					utils.Error("Failed to show help: " + err.Error())
				}
				return
			}
			utils.Info("Adding alias '" + args[1] + "' for command '" + args[0] + "'.")
			Aliases[args[0]] = args[1]
			config.Set("aliases", Aliases)
			utils.Success("Alias added successfully.")
		} else if cmd.Flag("remove").Changed {
			if len(args) != 1 {
				utils.Error("Invalid arguments. Usage: --remove <alias>")
				cmd.Help()
				return
			}
			utils.Info("Removing alias '" + args[0] + "'.")
			delete(Aliases, args[0])
			config.Set("aliases", Aliases)
			utils.Success("Alias removed successfully.")
		} else if cmd.Flag("list").Changed {
			utils.Info("Listing all aliases.")
			for cmdName, alias := range Aliases {
				cmd.Printf("%s -> %s\n", cmdName, alias)
			}
			utils.Success("Alias listing completed.")
		} else {
			utils.Error("No valid flag provided. Use --add, --remove, or --list.")
			cmd.Help()
		}
	},
}

func ReampAlias(root *cobra.Command) {
	loadAliases() // Ensure aliases are loaded before remapping
	utils.Debug("Re-mapping aliases to commands.")
	for cmdName, alias := range Aliases {
		cmd, _, err := root.Find([]string{cmdName})
		if err != nil {
			utils.Error("Error finding command '" + cmdName + "' - " + err.Error())
			continue
		}
		if cmd == nil {
			utils.Error("Command '" + cmdName + "' not found.")
			continue
		}
		cmd.Aliases = append(cmd.Aliases, alias)
		utils.Debug("Alias '" + alias + "' mapped to command '" + cmdName + "'.")
	}
}

func init() {
	utils.AddStatsPostRunToCommand(aliasCmd)
	rootCmd.AddCommand(aliasCmd)
	aliasCmd.Flags().StringSliceP("add", "a", []string{}, "Add a new alias for a command")
	aliasCmd.Flags().StringP("remove", "r", "", "Remove an existing alias")
	aliasCmd.Flags().BoolP("list", "l", false, "List all existing aliases")
}
