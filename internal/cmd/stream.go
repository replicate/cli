package cmd

import (
	"github.com/spf13/cobra"

	"github.com/replicate/cli/internal/cmd/prediction"
)

var StreamCmd = &cobra.Command{
	Use:   "stream <owner/model[:version]> [input=value] ... [flags]",
	Short: `Alias for "prediction create --stream"`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cmd.Flags().Set("stream", "true")
		if err != nil {
			return err
		}

		return prediction.CreateCmd.RunE(cmd, args)
	},
}

func init() {
	prediction.AddCreateFlags(StreamCmd)
}
