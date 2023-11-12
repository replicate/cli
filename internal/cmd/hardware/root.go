package hardware

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:     "hardware [subcommand]",
	Short:   "Interact with hardware",
	Aliases: []string{"hw"},
}

func init() {
	RootCmd.AddGroup(&cobra.Group{
		ID:    "subcommand",
		Title: "Subcommands:",
	})
	for _, cmd := range []*cobra.Command{
		listCmd,
	} {
		RootCmd.AddCommand(cmd)
		cmd.GroupID = "subcommand"
	}
}
