package training

import (
	"encoding/json"
	"fmt"

	"github.com/cli/browser"
	"github.com/replicate/cli/internal/util"
	"github.com/replicate/replicate-go"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:     "show <id>",
	Short:   "Show a training",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"view"},
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		if cmd.Flags().Changed("web") {
			if util.IsTTY() {
				fmt.Println("Opening in browser...")
			}

			url := fmt.Sprintf("https://replicate.com/p/%s", id)
			err := browser.OpenURL(url)
			if err != nil {
				return fmt.Errorf("failed to open browser: %w", err)
			}

			return nil
		}

		ctx := cmd.Context()

		client, err := replicate.NewClient(replicate.WithTokenFromEnv())
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}
		training, err := client.GetTraining(ctx, id)
		if training == nil || err != nil {
			return fmt.Errorf("failed to get training: %w", err)
		}

		if cmd.Flags().Changed("json") || !util.IsTTY() {
			bytes, err := json.MarshalIndent(training, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal training: %w", err)
			}
			fmt.Println(string(bytes))
			return nil
		}

		// TODO: render training with TUI
		fmt.Println(training.ID)
		fmt.Println("Status: " + training.Status)

		if training.CompletedAt != nil {
			fmt.Println("Completed at: " + *training.CompletedAt)
			fmt.Println("Inputs:")
			for key, value := range training.Input {
				fmt.Printf("  %s: %s\n", key, value)
			}

			fmt.Println("Outputs:")
			bytes, err := json.MarshalIndent(training.Output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal training output: %w", err)
			}
			fmt.Println(string(bytes))
		}

		return nil
	},
}

func init() {
	showCmd.Flags().Bool("json", false, "Emit JSON")
	showCmd.Flags().Bool("web", false, "Open in web browser")

	showCmd.MarkFlagsMutuallyExclusive("json", "web")
}
