package deployment

import (
	"encoding/json"
	"fmt"

	"github.com/cli/browser"
	"github.com/replicate/replicate-go"
	"github.com/spf13/cobra"

	"github.com/replicate/cli/internal/client"
	"github.com/replicate/cli/internal/identifier"
	"github.com/replicate/cli/internal/util"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:     "create <[owner/]name> [flags]",
	Short:   "Create a new deployment",
	Example: `replicate deployment create text-to-image --model=stability-ai/sdxl --hardware=gpu-a100-large`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r8, err := client.NewClient()
		if err != nil {
			return err
		}

		opts := &replicate.CreateDeploymentOptions{}

		opts.Name = args[0]

		flags := cmd.Flags()

		modelFlag, _ := flags.GetString("model")
		id, err := identifier.ParseIdentifier(modelFlag)
		if err != nil {
			return fmt.Errorf("expected <owner>/<name>[:version] but got %s", args[0])
		}
		opts.Model = fmt.Sprintf("%s/%s", id.Owner, id.Name)
		if id.Version != "" {
			opts.Version = id.Version
		} else {
			model, err := r8.GetModel(cmd.Context(), id.Owner, id.Name)
			if err != nil {
				return fmt.Errorf("failed to get model: %w", err)
			}
			opts.Version = model.LatestVersion.ID
		}

		opts.Hardware, _ = flags.GetString("hardware")

		flagMap := map[string]*int{
			"min-instances": &opts.MinInstances,
			"max-instances": &opts.MaxInstances,
		}
		for flagName, optPtr := range flagMap {
			if flags.Changed(flagName) {
				value, _ := flags.GetInt(flagName)
				*optPtr = value
			}
		}

		deployment, err := r8.CreateDeployment(cmd.Context(), *opts)
		if err != nil {
			return fmt.Errorf("failed to create deployment: %w", err)
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

		fmt.Printf("Deployment created: %s\n", url)

		return nil
	},
}

func init() {
	addCreateFlags(createCmd)
}

func addCreateFlags(cmd *cobra.Command) {
	cmd.Flags().String("model", "", "Model to deploy")
	_ = cmd.MarkFlagRequired("model")

	cmd.Flags().String("hardware", "", "SKU of the hardware to run the model")
	_ = cmd.MarkFlagRequired("hardware")

	cmd.Flags().Int("min-instances", 0, "Minimum number of instances to run the model")
	cmd.Flags().Int("max-instances", 0, "Maximum number of instances to run the model")

	cmd.Flags().Bool("json", false, "Emit JSON")
	cmd.Flags().Bool("web", false, "View on web")
	cmd.MarkFlagsMutuallyExclusive("json", "web")
}
