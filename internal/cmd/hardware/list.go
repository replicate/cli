package hardware

import (
	"encoding/json"
	"fmt"

	"github.com/cli/browser"
	"github.com/spf13/cobra"

	"github.com/replicate/cli/internal/client"
	"github.com/replicate/cli/internal/util"
)

// listCmd represents the list hardware command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List hardware",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()

		if cmd.Flags().Changed("web") {
			if util.IsTTY() {
				fmt.Println("Opening in browser...")
			}

			url := "https://replicate.com/pricing#hardware"
			err := browser.OpenURL(url)
			if err != nil {
				return fmt.Errorf("failed to open browser: %w", err)
			}

			return nil
		}

		r8, err := client.NewClient()
		if err != nil {
			return err
		}

		hardware, err := r8.ListHardware(ctx)
		if err != nil {
			return fmt.Errorf("failed to list hardware: %w", err)
		}

		if cmd.Flags().Changed("json") || !util.IsTTY() {
			bytes, err := json.MarshalIndent(hardware, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal hardware: %w", err)
			}
			fmt.Println(string(bytes))
			return nil
		}

		for _, hw := range *hardware {
			fmt.Printf("- %s: %s\n", hw.SKU, hw.Name)
		}

		return nil
	},
}

func init() {
	addListFlags(listCmd)
}

func addListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Emit JSON")
	cmd.Flags().Bool("web", false, "View on web")
	cmd.MarkFlagsMutuallyExclusive("json", "web")
}
