package model

import (
	"encoding/json"
	"fmt"

	"github.com/replicate/cli/internal/client"
	"github.com/replicate/cli/internal/identifier"
	"github.com/replicate/cli/internal/util"

	"github.com/replicate/replicate-go"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:     "schema <owner/model[:version]>",
	Short:   "Show the inputs and outputs of a model",
	Args:    cobra.ExactArgs(1),
	Example: `  replicate model schema stability-ai/sdxl`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := identifier.ParseIdentifier(args[0])
		if err != nil {
			return fmt.Errorf("invalid model specified: %s", args[0])
		}

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

		if cmd.Flags().Changed("json") || !util.IsTTY() {
			bytes, err := json.MarshalIndent(version.OpenAPISchema, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to serialize schema: %w", err)
			}
			fmt.Println(string(bytes))

			return nil
		}

		return printModelVersionSchema(version)
	},
}

func printModelVersionSchema(version *replicate.ModelVersion) error {
	inputSchema, outputSchema, err := util.GetSchemas(*version)
	if err != nil {
		return fmt.Errorf("failed to get schemas: %w", err)
	}

	if inputSchema != nil {
		fmt.Println("Inputs:")

		for _, propName := range util.SortedKeys(inputSchema.Properties) {
			prop, ok := inputSchema.Properties[propName]
			if !ok {
				continue
			}

			description := prop.Value.Description
			if prop.Value.Enum != nil {
				for _, enum := range prop.Value.Enum {
					description += fmt.Sprintf("\n- %s", enum)
				}
			}

			fmt.Printf("- %s: %s (type: %s)\n", propName, description, prop.Value.Type)
		}
		fmt.Println()
	}

	if outputSchema != nil {
		fmt.Println("Output:")
		fmt.Printf("- type: %s\n", outputSchema.Type)
		if outputSchema.Type.Is("array") {
			fmt.Printf("- items: %s %s\n", outputSchema.Items.Value.Type, outputSchema.Items.Value.Format)
		}
		fmt.Println()
	}

	return nil
}

func init() {
	schemaCmd.Flags().Bool("json", false, "Emit JSON")
}
