package model

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:     "model [subcommand]",
	Short:   "Interact with models",
	Aliases: []string{"models", "m"},
}

func init() {
	RootCmd.AddGroup(&cobra.Group{
		ID:    "subcommand",
		Title: "Subcommands:",
	})
	for _, cmd := range []*cobra.Command{
		showCmd,
		schemaCmd,
		createCmd,
		listCmd,
	} {
		RootCmd.AddCommand(cmd)
		cmd.GroupID = "subcommand"
	}

	RootCmd.AddGroup(&cobra.Group{
		ID:    "alias",
		Title: "Alias commands:",
	})
	for _, cmd := range []*cobra.Command{
		runCmd,
	} {
		RootCmd.AddCommand(cmd)
		cmd.GroupID = "alias"
	}
}
