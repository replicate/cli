package model

import (
	"fmt"

	"github.com/cli/browser"
	"github.com/replicate/cli/internal/identifier"
	"github.com/replicate/cli/internal/util"
	"github.com/replicate/replicate-go"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <owner>/<name> [flags]",
	Short: "Create a new model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := identifier.ParseIdentifier(args[0])
		if err != nil || id.Version != "" {
			return fmt.Errorf("expected <owner>/<name> but got %s", args[0])
		}

		opts := &replicate.CreateModelOptions{}
		flags := cmd.Flags()

		if flags.Changed("public") {
			opts.Visibility = "public"
		} else if flags.Changed("private") {
			opts.Visibility = "private"
		}

		opts.Hardware, _ = flags.GetString("hardware")

		flagMap := map[string]**string{
			"description":     &opts.Description,
			"github-url":      &opts.GithubURL,
			"paper-url":       &opts.PaperURL,
			"license-url":     &opts.LicenseURL,
			"cover-image-url": &opts.CoverImageURL,
		}
		for flagName, optPtr := range flagMap {
			if flags.Changed(flagName) {
				value, _ := flags.GetString(flagName)
				*optPtr = &value
			}
		}

		client, err := replicate.NewClient(replicate.WithTokenFromEnv())
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		model, err := client.CreateModel(cmd.Context(), id.Owner, id.Name, *opts)
		if err != nil {
			return fmt.Errorf("failed to create model: %w", err)
		}

		if flags.Changed("json") || !util.IsTTY() {
			bytes, err := model.MarshalJSON()
			if err != nil {
				return fmt.Errorf("failed to serialize model: %w", err)
			}
			fmt.Println(string(bytes))
			return nil
		}

		url := fmt.Sprintf("https://replicate.com/%s/%s", id.Owner, id.Name)
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

		fmt.Printf("Model created: %s\n", url)

		return nil
	},
}

func init() {
	addCreateFlags(createCmd)
}

func addCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("public", false, "Make the new model public")
	cmd.Flags().Bool("private", false, "Make the new model private")
	cmd.MarkFlagsOneRequired("public", "private")
	cmd.MarkFlagsMutuallyExclusive("public", "private")

	cmd.Flags().String("hardware", "", "SKU of the hardware to run the model")
	_ = cmd.MarkFlagRequired("hardware")

	cmd.Flags().String("description", "", "Description of the model")
	cmd.Flags().String("github-url", "", "URL of the GitHub repository")
	cmd.Flags().String("paper-url", "", "URL of the paper")
	cmd.Flags().String("license-url", "", "URL of the license")
	cmd.Flags().String("cover-image-url", "", "URL of the cover image")

	cmd.Flags().Bool("json", false, "Emit JSON")
	cmd.Flags().Bool("web", false, "View on web")
	cmd.MarkFlagsMutuallyExclusive("json", "web")
}
