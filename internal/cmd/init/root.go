package init

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/replicate/replicate-go"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "init <prediction-ID-or-URL> [<directory>] [--template=<template>]",
	Short: "Initialize a new local development environment from a prediction",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {

		// Check whether REPLICATE_API_TOKEN env var is set, if not exit with an error message
		if os.Getenv("REPLICATE_API_TOKEN") == "" {
			fmt.Println("REPLICATE_API_TOKEN environment variable not set. Please set it to your Replicate API token.")
			os.Exit(1)
		}

		predictionId, err := parsePredictionId(args[0])
		if err != nil {
			fmt.Println(fmt.Errorf("failed to parse prediction id: %w", err))
			os.Exit(1)
		}

		var directory string
		if len(args) == 2 {
			directory = args[1]
		} else {
			directory = predictionId
		}

		template, _ := cmd.Flags().GetString("template")

		client, err := replicate.NewClient(replicate.WithTokenFromEnv())
		if err != nil {
			fmt.Println(fmt.Errorf("failed to create client: %w", err))
		}

		ctx := cmd.Context()

		prediction, err := client.GetPrediction(ctx, predictionId)
		if prediction == nil || err != nil {
			fmt.Println(fmt.Errorf("failed to get prediction: %w", err))
		}

		switch template {
		case "node", "":
			return handleNodeTemplate(ctx, prediction, directory)
		case "python":
			return handlePythonTemplate(ctx, prediction, directory)
		default:
			return fmt.Errorf("unsupported template: %s, expected one of: node, python", template)
		}
	},
}

func init() {
	RootCmd.Flags().StringP("template", "t", "", "Starter git repo template to use. Currently supported: node, python")
}

// Parse the prediction id from a url, or return the prediction id if it's not a url
func parsePredictionId(value string) (string, error) {
	// Case 1: A prediction ID
	if !strings.Contains(value, "/") {
		return value, nil
	}

	// Case 2: A URL in the form https://replicate.com/p/{id}
	if strings.HasPrefix(value, "replicate.com/p/") || strings.HasPrefix(value, "https://replicate.com/p/") {
		splitUrl := strings.Split(value, "/")
		if len(splitUrl) == 0 {
			return "", fmt.Errorf("invalid URL format")
		}
		return splitUrl[len(splitUrl)-1], nil
	}

	// Case 3: A URL in the form https://api.replicate.com/v1/predictions/{id}
	if strings.HasPrefix(value, "api.replicate.com/v1/predictions/") || strings.HasPrefix(value, "https://api.replicate.com/v1/predictions/") {
		splitUrl := strings.Split(value, "/")
		if len(splitUrl) == 0 {
			return "", fmt.Errorf("invalid URL format")
		}
		return splitUrl[len(splitUrl)-1], nil
	}

	// Case 4: A URL in the form "https://replicate.com/*?prediction={id}"
	if strings.Contains(value, "replicate.com") || strings.Contains(value, "https://replicate.com") {
		parsedUrl, err := url.Parse(value)
		if err != nil {
			return "", fmt.Errorf("failed to parse URL: %w", err)
		}
		queryParams, err := url.ParseQuery(parsedUrl.RawQuery)
		if err != nil {
			return "", fmt.Errorf("failed to parse query parameters: %w", err)
		}
		predictionId := queryParams.Get("prediction")
		if predictionId == "" {
			return "", fmt.Errorf("no prediction ID found in URL")
		}
		return predictionId, nil
	}

	// If none of the above cases match, return an error
	return "", fmt.Errorf("invalid prediction ID or URL format")
}

func execCommand(ctx context.Context, command string) error {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}
	return nil
}

func handleNodeTemplate(ctx context.Context, prediction *replicate.Prediction, directory string) error {
	fmt.Println("Cloning starter repo and installing dependencies...")

	// 1. Clone the starter repo
	if err := execCommand(ctx, fmt.Sprintf("git clone https://github.com/replicate/node-starter.git %s", directory)); err != nil {
		return fmt.Errorf("failed to clone starter repo: %w", err)
	}

	// 2. Set chdir to the output path
	if err := os.Chdir(directory); err != nil {
		fmt.Println(fmt.Errorf("failed to change directory: %w", err))
		os.Exit(1)
	}

	// 3. Install dependencies
	if err := execCommand(ctx, "npm install"); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// 4. Set the REPLICATE_API_TOKEN env var
	if err := execCommand(ctx, fmt.Sprintf(`echo 'REPLICATE_API_TOKEN="%s"' >> .env`, os.Getenv("REPLICATE_API_TOKEN"))); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	// Open the template file
	templateData, err := os.ReadFile("index.js.template")
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	fullModelString := fmt.Sprintf("%s:%s", prediction.Model, prediction.Version)

	// Perform string replacement on the template file.
	replacedData := strings.ReplaceAll(string(templateData), "{{MODEL_STRING}}", fullModelString)
	inputs, _ := json.Marshal(prediction.Input)
	replacedData = strings.ReplaceAll(replacedData, "{{INPUTS}}", string(inputs))

	// 5. Write the populated template to {directory}/index.js
	fmt.Println("Writing new index.js...")
	err = os.WriteFile("index.js", []byte(replacedData), 0o644)
	if err != nil {
		return err
	}

	// 6. Run the example prediction
	fmt.Println("Running example prediction...")
	if err := execCommand(ctx, "node index.js"); err != nil {
		return fmt.Errorf("failed to run example prediction: %w", err)
	}

	return nil
}

func handlePythonTemplate(ctx context.Context, prediction *replicate.Prediction, directory string) error {
	fmt.Println("Cloning starter repo and installing dependencies...")

	// 1. Clone the starter repo
	if err := execCommand(ctx, fmt.Sprintf("git clone git@github.com:replicate/python-starter.git %s", directory)); err != nil {
		return fmt.Errorf("failed to clone starter repo: %w", err)
	}

	// 2. Set chdir to the output path
	err := os.Chdir(directory)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to change directory: %w", err))
		os.Exit(1)
	}

	// 3. Create virtualenv
	if err := execCommand(ctx, "virtualenv .venv"); err != nil {
		return fmt.Errorf("failed to create virtualenv: %w", err)
	}

	// 4. Install dependencies
	if err := execCommand(ctx, ".venv/bin/pip install -r requirements.txt"); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// 5. Set the REPLICATE_API_TOKEN env var
	if err := execCommand(ctx, fmt.Sprintf(`echo 'REPLICATE_API_TOKEN="%s"' >> .env`, os.Getenv("REPLICATE_API_TOKEN"))); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	fmt.Println("Cloning starter repo, and installing dependencies...")

	// Open the template file
	templateData, err := os.ReadFile("prediction.py.template")
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}
	fullModelString := fmt.Sprintf("%s:%s", prediction.Model, prediction.Version)

	replacedData := strings.ReplaceAll(string(templateData), "{{MODEL_STRING}}", fullModelString)
	inputs, _ := json.Marshal(prediction.Input)
	replacedData = strings.ReplaceAll(replacedData, "{{INPUTS}}", string(inputs))

	// 6. Write the populated template to {directory}/prediction.py
	fmt.Println("Writing new prediction.py...")
	err = os.WriteFile("prediction.py", []byte(replacedData), 0o644)
	if err != nil {
		return err
	}

	// 7. Run the example prediction
	fmt.Println("Running example prediction...")
	if err := execCommand(ctx, ".venv/bin/python prediction.py"); err != nil {
		return fmt.Errorf("failed to run example prediction: %w", err)
	}

	return nil
}
