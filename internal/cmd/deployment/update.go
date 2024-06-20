package deployment

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cli/browser"
	"github.com/replicate/replicate-go"
	"github.com/spf13/cobra"

	"github.com/replicate/cli/internal/client"
	"github.com/replicate/cli/internal/identifier"
	"github.com/replicate/cli/internal/util"
)

// updateCmd represents the create command
var updateCmd = &cobra.Command{
	Use:     "update <[owner/]name> [flags]",
	Short:   "Update an existing deployment",
	Example: `replicate deployment update acme/text-to-image --max-instances=2`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r8, err := client.NewClient()
		if err != nil {
			return err
		}

		name := args[0]
		if !strings.Contains(name, "/") {
			account, err := r8.GetCurrentAccount(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get current account: %w", err)
			}
			name = fmt.Sprintf("%s/%s", account.Username, name)
		}
		deploymentID, err := identifier.ParseIdentifier(name)
		if err != nil {
			return fmt.Errorf("invalid deployment specified: %s", name)
		}

		opts := &replicate.UpdateDeploymentOptions{}

		flags := cmd.Flags()

		if flags.Changed("version") {
			value, _ := flags.GetString("version")
			var version string
			if strings.Contains(value, ":") {
				modelID, err := identifier.ParseIdentifier(value)
				if err != nil {
					return fmt.Errorf("invalid model version specified: %s", value)
				}
				version = modelID.Version
			} else {
				version = value
			}
			opts.Version = &version
		}

		if flags.Changed("hardware") {
			value, _ := flags.GetString("hardware")
			opts.Hardware = &value
		}

		if flags.Changed("min-instances") {
			value, _ := flags.GetInt("min-instances")
			opts.MinInstances = &value
		}

		if flags.Changed("max-instances") {
			value, _ := flags.GetInt("max-instances")
			opts.MaxInstances = &value
		}

		deployment, err := r8.UpdateDeployment(cmd.Context(), deploymentID.Owner, deploymentID.Name, *opts)
		if err != nil {
			return fmt.Errorf("failed to update deployment: %w", err)
		}

		if flags.Changed("json") || !util.IsTTY() {
			bytes, err := json.MarshalIndent(deployment, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to serialize model: %w", err)
			}
			fmt.Println(string(bytes))
			return nil
		}

		url := fmt.Sprintf("https://replicate.com/deployments/%s/%s", deployment.Owner, deployment.Name)
		if flags.Changed("web") {
			if util.IsTTY() {
				fmt.Println("Opening in browser...")
			}

			err := browser.OpenURL(url)
			if err != nil {
				return fmt.Errorf("failed to open browser: %w", err)
			}

			return nil
		}

		fmt.Printf("Deployment updated: %s\n", url)

		return nil
	},
}

func init() {
	addUpdateFlags(updateCmd)
}

func addUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().String("version", "", "Version of the model to deploy")
	cmd.Flags().String("hardware", "", "SKU of the hardware to run the model")
	cmd.Flags().Int("min-instances", 0, "Minimum number of instances to run the model")
	cmd.Flags().Int("max-instances", 0, "Maximum number of instances to run the model")

	cmd.Flags().Bool("json", false, "Emit JSON")
	cmd.Flags().Bool("web", false, "View on web")
	cmd.MarkFlagsMutuallyExclusive("json", "web")
}
