package training

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:     "training [subcommand]",
	Short:   "Interact with trainings",
	Aliases: []string{"trainings", "t"},
}

func init() {
	RootCmd.AddGroup(&cobra.Group{
		ID:    "subcommand",
		Title: "Subcommands:",
	})
	for _, cmd := range []*cobra.Command{
		CreateCmd,
		listCmd,
		showCmd,
	} {
		RootCmd.AddCommand(cmd)
		cmd.GroupID = "subcommand"
	}
}
