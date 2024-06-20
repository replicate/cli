package deployment

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:     "deployments [subcommand]",
	Short:   "Interact with deployments",
	Aliases: []string{"deployments", "d"},
}

func init() {
	RootCmd.AddGroup(&cobra.Group{
		ID:    "subcommand",
		Title: "Subcommands:",
	})
	for _, cmd := range []*cobra.Command{
		listCmd,
		showCmd,
		schemaCmd,
		createCmd,
		updateCmd,
	} {
		RootCmd.AddCommand(cmd)
		cmd.GroupID = "subcommand"
	}

	// RootCmd.AddGroup(&cobra.Group{
	// 	ID:    "alias",
	// 	Title: "Alias commands:",
	// })
	// for _, cmd := range []*cobra.Command{
	// 	runCmd,
	// } {
	// 	RootCmd.AddCommand(cmd)
	// 	cmd.GroupID = "alias"
	// }
}
