package model

import (
	"encoding/json"
	"fmt"

	"github.com/cli/browser"
	"github.com/replicate/cli/internal/identifier"
	"github.com/replicate/cli/internal/util"
	"github.com/replicate/replicate-go"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show a model",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"view"},
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := identifier.ParseIdentifier(args[0])
		if err != nil {
			return fmt.Errorf("invalid model specified: %s", args[0])
		}

		if cmd.Flags().Changed("web") {
			if util.IsTTY() {
				fmt.Println("Opening in browser...")
			}

			var url string
			if id.Version != "" {
				url = fmt.Sprintf("https://replicate.com/%s/%s/versions/%s", id.Owner, id.Name, id.Version)
			} else {
				url = fmt.Sprintf("https://replicate.com/%s/%s", id.Owner, id.Name)
			}

			err := browser.OpenURL(url)
			if err != nil {
				return fmt.Errorf("failed to open browser: %w", err)
			}

			return nil
		}

		ctx := cmd.Context()

		var model *replicate.Model
		// var version *replicate.ModelVersion

		client, err := replicate.NewClient(replicate.WithTokenFromEnv())
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		model, err = client.GetModel(ctx, id.Owner, id.Name)
		if err != nil {
			return fmt.Errorf("failed to get model: %w", err)
		}

		if cmd.Flags().Changed("json") || !util.IsTTY() {
			bytes, err := json.MarshalIndent(model, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal model: %w", err)
			}
			fmt.Println(string(bytes))
			return nil
		}

		if id.Version != "" {
			fmt.Println("Ignoring specified version", id.Version)
		}

		fmt.Println(model.Name)
		fmt.Println(model.Description)
		fmt.Println()
		fmt.Println("Latest version:", model.LatestVersion.ID)

		return nil
	},
}

func init() {
	showCmd.Flags().Bool("json", false, "Emit JSON")
	showCmd.Flags().Bool("web", false, "Open in web browser")

	showCmd.MarkFlagsMutuallyExclusive("json", "web")
}
