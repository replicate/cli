package cmd

import (
	"github.com/spf13/cobra"

	"github.com/replicate/cli/internal/cmd/training"
)

var TrainCmd = &cobra.Command{
	Use:   "train <owner/model[:version]> [input=value] ... [flags]",
	Short: `Alias for "training create"`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  training.CreateCmd.RunE,
}

func init() {
	training.AddCreateFlags(TrainCmd)
}
