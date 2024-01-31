package account

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:     "account [subcommand]",
	Short:   "Interact with accounts",
	Aliases: []string{"accounts", "a"},
}

func init() {
	RootCmd.AddGroup(&cobra.Group{
		ID:    "subcommand",
		Title: "Subcommands:",
	})
	for _, cmd := range []*cobra.Command{
		CurrentCmd,
	} {
		RootCmd.AddCommand(cmd)
		cmd.GroupID = "subcommand"
	}
}
