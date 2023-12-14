package auth

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "auth [subcommand]",
	Short: "Authenticate with Replicate",
}

func init() {
	RootCmd.AddGroup(&cobra.Group{
		ID:    "subcommand",
		Title: "Subcommands:",
	})
	for _, cmd := range []*cobra.Command{
		loginCmd,
	} {
		RootCmd.AddCommand(cmd)
		cmd.GroupID = "subcommand"
	}
}
