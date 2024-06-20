package deployment

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cli/browser"
	"github.com/spf13/cobra"

	"github.com/replicate/cli/internal/client"
	"github.com/replicate/cli/internal/identifier"
	"github.com/replicate/cli/internal/util"
)

var showCmd = &cobra.Command{
	Use:     "show <[owner/]name> [flags]",
	Short:   "Show a deployment",
	Example: "replicate deployment show acme/text-to-image",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"view"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		r8, err := client.NewClient()
		if err != nil {
			return err
		}

		name := args[0]
		if !strings.Contains(name, "/") {
			account, err := r8.GetCurrentAccount(ctx)
			if err != nil {
				return fmt.Errorf("failed to get current account: %w", err)
			}
			name = fmt.Sprintf("%s/%s", account.Username, name)
		}
		id, err := identifier.ParseIdentifier(name)
		if err != nil {
			return fmt.Errorf("invalid deployment specified: %s", name)
		}

		if cmd.Flags().Changed("web") {
			if util.IsTTY() {
				fmt.Println("Opening in browser...")
			}

			url := fmt.Sprintf("https://replicate.com/deployments/%s/%s", id.Owner, id.Name)
			err := browser.OpenURL(url)
			if err != nil {
				return fmt.Errorf("failed to open browser: %w", err)
			}

			return nil
		}

		deployment, err := r8.GetDeployment(ctx, id.Owner, id.Name)
		if err != nil {
			return fmt.Errorf("failed to get deployment: %w", err)
		}

		if cmd.Flags().Changed("json") || !util.IsTTY() {
			bytes, err := json.MarshalIndent(deployment, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal model: %w", err)
			}
			fmt.Println(string(bytes))
			return nil
		}

		if id.Version != "" {
			fmt.Println("Ignoring specified version", id.Version)
		}

		fmt.Printf("%s/%s\n", deployment.Owner, deployment.Name)
		fmt.Println()
		fmt.Printf("Release #%d\n", deployment.CurrentRelease.Number)
		fmt.Println("Model:", deployment.CurrentRelease.Model)
		fmt.Println("Version:", deployment.CurrentRelease.Version)
		fmt.Println("Hardware:", deployment.CurrentRelease.Configuration.Hardware)
		fmt.Println("Min instances:", deployment.CurrentRelease.Configuration.MinInstances)
		fmt.Println("Max instances:", deployment.CurrentRelease.Configuration.MaxInstances)

		return nil
	},
}

func init() {
	showCmd.Flags().Bool("json", false, "Emit JSON")
	showCmd.Flags().Bool("web", false, "Open in web browser")

	showCmd.MarkFlagsMutuallyExclusive("json", "web")
}
