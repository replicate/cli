package training

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/cli/browser"
	"github.com/replicate/replicate-go"
	"github.com/spf13/cobra"

	"github.com/replicate/cli/internal/client"
	"github.com/replicate/cli/internal/identifier"
	"github.com/replicate/cli/internal/util"
)

// CreateCmd represents the create command
var CreateCmd = &cobra.Command{
	Use:     "create <owner/model[:version]> --destination <owner/model> [input=value] ... [flags]",
	Short:   "Create a training",
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"new", "train"},
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO support running interactively

		destination := cmd.Flag("destination").Value.String()
		if _, err := identifier.ParseIdentifier(destination); err != nil {
			return fmt.Errorf("invalid destination specified: %s", destination)
		}

		// parse arg into model.Identifier
		id, err := identifier.ParseIdentifier(args[0])
		if err != nil {
			return fmt.Errorf("invalid model specified: %s", args[0])
		}

		s := spinner.New(spinner.CharSets[21], 100*time.Millisecond)
		s.FinalMSG = ""

		ctx := cmd.Context()

		r8, err := client.NewClient()
		if err != nil {
			return err
		}

		var version *replicate.ModelVersion
		if id.Version == "" {
			model, err := r8.GetModel(ctx, id.Owner, id.Name)
			if err != nil {
				return fmt.Errorf("failed to get model: %w", err)
			}

			if model.LatestVersion == nil {
				return fmt.Errorf("no versions found for model %s", args[0])
			}

			version = model.LatestVersion
		} else {
			version, err = r8.GetModelVersion(ctx, id.Owner, id.Name, id.Version)
			if err != nil {
				return fmt.Errorf("failed to get model version: %w", err)
			}
		}

		stdin, err := util.GetPipedArgs()
		if err != nil {
			return fmt.Errorf("failed to get stdin info: %w", err)
		}

		separator := cmd.Flag("separator").Value.String()
		inputs, err := util.ParseInputs(ctx, r8, args[1:], stdin, separator)
		if err != nil {
			return fmt.Errorf("failed to parse inputs: %w", err)
		}

		coercedInputs, err := util.CoerceTypes(inputs, nil)
		if err != nil {
			return fmt.Errorf("failed to coerce inputs: %w", err)
		}

		s.Start()
		training, err := r8.CreateTraining(ctx, id.Owner, id.Name, version.ID, destination, coercedInputs, nil)
		if err != nil {
			return fmt.Errorf("failed to create training: %w", err)
		}
		s.Stop()

		url := fmt.Sprintf("https://replicate.com/p/%s", training.ID)
		fmt.Printf("Training created: %s\n", url)

		if cmd.Flags().Changed("web") {
			if util.IsTTY() {
				fmt.Println("Opening in browser...")
			}

			err = browser.OpenURL(url)
			if err != nil {
				return fmt.Errorf("failed to open browser: %w", err)
			}

			return nil
		}

		if cmd.Flags().Changed("json") || !util.IsTTY() {
			b, err := json.Marshal(training)
			if err != nil {
				return fmt.Errorf("failed to marshal training: %w", err)
			}

			fmt.Println(string(b))
			return nil
		}

		return nil
	},
}

func init() {
	AddCreateFlags(CreateCmd)
}

func AddCreateFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("destination", "d", "", "Destination model for training")

	cmd.Flags().Bool("json", false, "Emit JSON")
	cmd.Flags().Bool("web", false, "View on web")
	cmd.Flags().String("separator", "=", "Separator between input key and value")

	cmd.MarkFlagsMutuallyExclusive("json", "web")
}
