package prediction

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:     "prediction [subcommand]",
	Short:   "Interact with predictions",
	Aliases: []string{"predictions", "p"},
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
