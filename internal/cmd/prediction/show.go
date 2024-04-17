package prediction

import (
	"encoding/json"
	"fmt"

	"github.com/cli/browser"
	"github.com/spf13/cobra"

	"github.com/replicate/cli/internal/client"
	"github.com/replicate/cli/internal/util"
)

var showCmd = &cobra.Command{
	Use:     "show <id>",
	Short:   "Show a prediction",
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

		r8, err := client.NewClient()
		if err != nil {
			return err
		}

		prediction, err := r8.GetPrediction(ctx, id)
		if prediction == nil || err != nil {
			return fmt.Errorf("failed to get prediction: %w", err)
		}

		if cmd.Flags().Changed("json") || !util.IsTTY() {
			bytes, err := json.MarshalIndent(prediction, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal prediction: %w", err)
			}
			fmt.Println(string(bytes))
			return nil
		}

		// TODO: render prediction with TUI
		fmt.Println(prediction.ID)
		fmt.Println("Status: " + prediction.Status)

		if prediction.CompletedAt != nil {
			fmt.Println("Completed at: " + *prediction.CompletedAt)
			fmt.Println("Inputs:")
			for key, value := range prediction.Input {
				fmt.Printf("  %s: %s\n", key, value)
			}

			fmt.Println("Outputs:")
			bytes, err := json.MarshalIndent(prediction.Output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal prediction output: %w", err)
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
