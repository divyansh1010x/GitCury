package cmd

import (
	"GitCury/config"
	"GitCury/utils"

	"github.com/spf13/cobra"
)

var Aliases = func() map[string]string {
	rawAliases := config.Get("aliases").(map[string]interface{})
	convertedAliases := make(map[string]string)
	for key, value := range rawAliases {
		convertedAliases[key] = value.(string)
	}
	return convertedAliases
}()

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
		if cmd.Flag("add").Changed {
			if len(args) != 2 {
				utils.Error("[ALIAS]: Invalid arguments. Usage: --add <command> <alias>")
				cmd.Help()
				return
			}
			utils.Info("[ALIAS]: Adding alias '" + args[1] + "' for command '" + args[0] + "'.")
			Aliases[args[0]] = args[1]
			config.Set("aliases", Aliases)
			utils.Success("[ALIAS]: Alias added successfully.")
		} else if cmd.Flag("remove").Changed {
			if len(args) != 1 {
				utils.Error("[ALIAS]: Invalid arguments. Usage: --remove <alias>")
				cmd.Help()
				return
			}
			utils.Info("[ALIAS]: Removing alias '" + args[0] + "'.")
			delete(Aliases, args[0])
			config.Set("aliases", Aliases)
			utils.Success("[ALIAS]: Alias removed successfully.")
		} else if cmd.Flag("list").Changed {
			utils.Info("[ALIAS]: Listing all aliases.")
			for cmdName, alias := range Aliases {
				cmd.Printf("%s -> %s\n", cmdName, alias)
			}
			utils.Success("[ALIAS]: Alias listing completed.")
		} else {
			utils.Error("[ALIAS]: No valid flag provided. Use --add, --remove, or --list.")
			cmd.Help()
		}
	},
}

func ReampAlias(root *cobra.Command) {
	utils.Info("[ALIAS]: Re-mapping aliases to commands.")
	for cmdName, alias := range Aliases {
		cmd, _, err := root.Find([]string{cmdName})
		if err != nil {
			utils.Error("[ALIAS]: Error finding command '" + cmdName + "' - " + err.Error())
			continue
		}
		if cmd == nil {
			utils.Error("[ALIAS]: Command '" + cmdName + "' not found.")
			continue
		}
		cmd.Aliases = append(cmd.Aliases, alias)
		utils.Success("[ALIAS]: Alias '" + alias + "' mapped to command '" + cmdName + "'.")
	}
}

func init() {
	rootCmd.AddCommand(aliasCmd)
	aliasCmd.Flags().StringArrayP("add", "a", []string{}, "Add a new alias for a command")
	aliasCmd.Flags().StringP("remove", "r", "", "Remove an existing alias")
	aliasCmd.Flags().BoolP("list", "l", false, "List all existing aliases")
}
