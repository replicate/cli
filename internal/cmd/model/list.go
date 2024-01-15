package model

import (
	"encoding/json"
	"fmt"

	"github.com/cli/browser"
	"github.com/spf13/cobra"

	"github.com/replicate/cli/internal/client"
	"github.com/replicate/cli/internal/util"
)

// listCmd represents the list models command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List models",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if cmd.Flags().Changed("web") {
			if util.IsTTY() {
				fmt.Println("Opening in browser...")
			}

			url := "https://replicate.com/explore"
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

		models, err := r8.ListModels(ctx)
		if err != nil {
			return fmt.Errorf("failed to list models: %w", err)
		}

		if cmd.Flags().Changed("json") || !util.IsTTY() {
			bytes, err := json.MarshalIndent(models, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal models: %w", err)
			}
			fmt.Println(string(bytes))
			return nil
		}

		for _, model := range models.Results {
			fmt.Printf("%s/%s\n", model.Owner, model.Name)
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
