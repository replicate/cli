package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/cli/browser"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/replicate/cli/internal/identifier"
	"github.com/replicate/cli/internal/util"
	"github.com/replicate/replicate-go"
	"github.com/spf13/cobra"
)

var StreamCmd = &cobra.Command{
	Use:   "stream <owner/model[:version]> [input=value] ... [flags]",
	Short: "Run a model and stream its output",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := identifier.ParseIdentifier(args[0])
		if err != nil {
			return fmt.Errorf("invalid model specified: %s", args[0])
		}

		s := spinner.New(spinner.CharSets[21], 100*time.Millisecond)
		s.FinalMSG = ""

		ctx := cmd.Context()

		client, err := replicate.NewClient(replicate.WithTokenFromEnv())
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		var version *replicate.ModelVersion
		if id.Version == "" {
			if model, err := client.GetModel(ctx, id.Owner, id.Name); err != nil {
				version = model.LatestVersion
			}
		} else {
			if v, err := client.GetModelVersion(ctx, id.Owner, id.Name, id.Version); err != nil {
				version = v
			}
		}

		stdin, err := util.GetPipedArgs()
		if err != nil {
			return fmt.Errorf("failed to get stdin info: %w", err)
		}

		separator := cmd.Flag("separator").Value.String()

		parsedInputs, err := util.ParseInputs(ctx, args[1:], stdin, separator)
		if err != nil {
			return fmt.Errorf("failed to parse inputs: %w", err)
		}

		var inputSchema *openapi3.Schema
		if version != nil {
			inputSchema, _, err = util.GetSchemas(*version)
			if err != nil {
				return fmt.Errorf("failed to get input schema for version: %w", err)
			}
		}

		coercedInputs, err := util.CoerceTypes(parsedInputs, inputSchema)
		if err != nil {
			return fmt.Errorf("failed to coerce inputs: %w", err)
		}

		s.Start()
		var prediction *replicate.Prediction
		if id.Version == "" {
			prediction, err = client.CreatePredictionWithModel(ctx, id.Owner, id.Name, coercedInputs, nil, true)
			// TODO check status code
			if err != nil && strings.Contains(err.Error(), "not found") {
				if version != nil {
					prediction, err = client.CreatePrediction(ctx, version.ID, coercedInputs, nil, true)
				}
			}
		} else {
			prediction, err = client.CreatePrediction(ctx, id.Version, coercedInputs, nil, true)
		}
		if err != nil {
			return fmt.Errorf("failed to create prediction: %w", err)
		}
		s.Stop()

		if cmd.Flags().Changed("web") {
			url := fmt.Sprintf("https://replicate.com/p/%s", prediction.ID)

			if util.IsTTY() {
				fmt.Printf("Prediction created: %s\n", url)
				fmt.Println("Opening in browser...")
			}

			err = browser.OpenURL(url)
			if err != nil {
				return fmt.Errorf("failed to open browser: %w", err)
			}

			return nil
		}

		sseChan, errChan := client.StreamPrediction(ctx, prediction)
		for {
			select {
			case event, ok := <-sseChan:
				if !ok {
					return nil
				}

				switch event.Type {
				case "output":
					fmt.Print(event.Data)
				case "logs":
					// TODO: print logs to stderr
				default:
					// ignore
				}
			case err, ok := <-errChan:
				if !ok {
					return nil
				}

				return fmt.Errorf("streaming error: %w", err)
			}
		}
	},
}

func init() {
	AddStreamFlags(StreamCmd)
}

func AddStreamFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("web", false, "View on web")
	cmd.Flags().String("separator", "=", "Separator between input key and value")
}
