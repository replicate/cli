package cmd

import (
	"github.com/spf13/cobra"

	"github.com/replicate/cli/internal/cmd/prediction"
)

var RunCmd = &cobra.Command{
	Use:   "run <owner/model[:version]> [input=value] ... [flags]",
	Short: `Alias for "prediction create"`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  prediction.CreateCmd.RunE,
}

func init() {
	prediction.AddCreateFlags(RunCmd)
}
