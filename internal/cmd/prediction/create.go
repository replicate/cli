package prediction

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/cli/browser"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/replicate/cli/internal/identifier"
	"github.com/replicate/cli/internal/util"
	"github.com/replicate/replicate-go"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var CreateCmd = &cobra.Command{
	Use:     "create <owner/model[:version]> [input=value] ... [flags]",
	Short:   "Create a prediction",
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"new", "run"},
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO support running interactively

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
			if model, err := client.GetModel(ctx, id.Owner, id.Name); err == nil {
				version = model.LatestVersion
			}
		} else {
			if v, err := client.GetModelVersion(ctx, id.Owner, id.Name, id.Version); err == nil {
				version = v
			}
		}

		stdin, err := util.GetPipedArgs()
		if err != nil {
			return fmt.Errorf("failed to get stdin info: %w", err)
		}

		separator := cmd.Flag("separator").Value.String()
		inputs, err := util.ParseInputs(ctx, args[1:], stdin, separator)
		if err != nil {
			return fmt.Errorf("failed to parse inputs: %w", err)
		}

		var inputSchema *openapi3.Schema
		var outputSchema *openapi3.Schema
		if version != nil {
			inputSchema, outputSchema, err = util.GetSchemas(*version)
			if err != nil {
				return fmt.Errorf("failed to get input schema for version: %w", err)
			}
		}

		coercedInputs, err := util.CoerceTypes(inputs, inputSchema)
		if err != nil {
			return fmt.Errorf("failed to coerce inputs: %w", err)
		}

		shouldWait := cmd.Flags().Changed("wait") || !cmd.Flags().Changed("no-wait")
		shouldStream := !cmd.Flags().Changed("wait") && cmd.Flags().Changed("stream") || (outputSchema != nil && outputSchema.Type == "array" && outputSchema.Items.Value.Type == "string" && outputSchema.Items.Value.Format != "uri")

		s.Start()
		var prediction *replicate.Prediction
		if id.Version == "" {
			prediction, err = client.CreatePredictionWithModel(ctx, id.Owner, id.Name, coercedInputs, nil, shouldStream)
			// TODO: check status code
			if err != nil {
				if version != nil {
					prediction, err = client.CreatePrediction(ctx, version.ID, coercedInputs, nil, shouldStream)
				}
			}
		} else {
			prediction, err = client.CreatePrediction(ctx, id.Version, coercedInputs, nil, shouldStream)
		}
		if err != nil {
			return fmt.Errorf("failed to create prediction: %w", err)
		}
		s.Stop()

		hasStream := prediction.URLs["stream"] != ""

		if !util.IsTTY() || cmd.Flags().Changed("json") {
			if hasStream {
				events, _ := client.StreamPrediction(ctx, prediction)

				if cmd.Flags().Changed("json") {
					fmt.Print("[")
					defer fmt.Print("]")
				}

				for event := range events {
					if cmd.Flags().Changed("json") {
						b, err := json.Marshal(event.Data)
						if err != nil {
							return fmt.Errorf("failed to marshal event: %w", err)
						}
						fmt.Printf("%s,", string(b))
					} else {
						fmt.Print(event.Data)
					}
				}
				fmt.Println("")

				return nil
			}

			if shouldWait {
				err = client.Wait(ctx, prediction)
				if err != nil {
					return fmt.Errorf("failed to wait for prediction: %w", err)
				}
			}

			b, err := json.Marshal(prediction)
			if err != nil {
				return fmt.Errorf("failed to marshal prediction: %w", err)
			}
			fmt.Println(string(b))

			return nil
		}

		url := fmt.Sprintf("https://replicate.com/p/%s", prediction.ID)
		if !hasStream {
			fmt.Printf("Prediction created: %s\n", url)
		}

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

		if hasStream {
			sseChan, errChan := client.StreamPrediction(ctx, prediction)

			tokens := []string{}
			for {
				select {
				case event, ok := <-sseChan:
					if !ok {
						return nil
					}

					switch event.Type {
					case replicate.SSETypeOutput:
						token := event.Data
						tokens = append(tokens, token)
						fmt.Print(token)
					case replicate.SSETypeLogs:
						// TODO: print logs to stderr
					case replicate.SSETypeDone:
						return nil
					default:
						// ignore
					}
				case err, ok := <-errChan:
					if !ok {
						return nil
					}

					return fmt.Errorf("streaming error: %w", err)
				}

				if cmd.Flags().Changed("save") {
					var dirname string
					if cmd.Flags().Changed("output-directory") {
						dirname = cmd.Flag("output-directory").Value.String()
					} else {
						dirname = fmt.Sprintf("./%s", prediction.ID)
					}

					dir, err := filepath.Abs(dirname)
					if err != nil {
						return fmt.Errorf("failed to create output directory: %w", err)
					}

					err = os.MkdirAll(dir, 0o755)
					if err != nil {
						return fmt.Errorf("failed to create directory: %w", err)
					}

					err = os.WriteFile(filepath.Join(dir, "output.txt"), []byte(strings.Join(tokens, "")), 0o644)
					if err != nil {
						return fmt.Errorf("failed to write output: %w", err)
					}
				}
			}
		} else if shouldWait {
			bar := progressbar.Default(100)
			bar.Describe("processing")

			predChan, errChan := client.WaitAsync(ctx, prediction)
			for pred := range predChan {
				progress := pred.Progress()
				if progress != nil {
					bar.ChangeMax(progress.Total)
					_ = bar.Set(progress.Current)
				}

				if pred.Status.Terminated() {
					_ = bar.Finish()
					break
				}
			}

			if err := <-errChan; err != nil {
				return fmt.Errorf("failed to wait for prediction: %w", err)
			}

			switch prediction.Status {
			case replicate.Succeeded:
				fmt.Println("✅ Succeeded")
				bytes, err := json.MarshalIndent(prediction.Output, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal output: %w", err)
				}
				fmt.Println(string(bytes))
			case replicate.Failed:
				fmt.Println("❌ Failed")
				fmt.Println(*prediction.Logs)
				bytes, err := json.MarshalIndent(prediction.Error, "", "  ")
				if err != nil {
					return fmt.Errorf("error: %v", prediction.Error)
				}
				fmt.Println(string(bytes))
			case replicate.Canceled:
				fmt.Println("🚫 Canceled")
				fmt.Println(prediction.Logs)
			}

			if cmd.Flags().Changed("save") && prediction.Status == replicate.Succeeded {
				var dirname string
				if cmd.Flags().Changed("output-directory") {
					dirname = cmd.Flag("output-directory").Value.String()
				} else {
					dirname = fmt.Sprintf("./%s", prediction.ID)
				}

				dir, err := filepath.Abs(dirname)
				if err != nil {
					return fmt.Errorf("failed to create output directory: %w", err)
				}

				err = util.DownloadPrediction(ctx, *prediction, dir)
				if err != nil {
					return fmt.Errorf("failed to save output: %w", err)
				}
			}
		}

		return nil

	},
}

func init() {
	AddCreateFlags(CreateCmd)
}

func AddCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Emit JSON")
	cmd.Flags().Bool("no-wait", false, "Don't wait for prediction to complete")
	cmd.Flags().BoolP("wait", "w", true, "Wait for prediction to complete")
	cmd.Flags().Bool("stream", false, "Stream prediction output")
	cmd.Flags().Bool("web", false, "View on web")
	cmd.Flags().String("separator", "=", "Separator between input key and value")
	cmd.Flags().Bool("save", false, "Save prediction outputs to directory")
	cmd.Flags().String("output-directory", "", "Output directory, defaults to ./{prediction-id}")

	cmd.MarkFlagsMutuallyExclusive("json", "web")
	cmd.MarkFlagsMutuallyExclusive("stream", "wait")
	cmd.MarkFlagsMutuallyExclusive("wait", "no-wait")
}
