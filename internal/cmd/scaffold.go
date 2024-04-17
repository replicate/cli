package cmd

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

var ScaffoldCmd = &cobra.Command{
	Use:   "scaffold <prediction-ID-or-URL> [<directory>] [--template=<template>]",
	Short: "Create a new local development environment from a prediction",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		apiToken := os.Getenv("REPLICATE_API_TOKEN")
		if apiToken == "" {
			return fmt.Errorf("REPLICATE_API_TOKEN environment variable not set. Please set this to your Replicate API token")
		}

		client, err := replicate.NewClient(replicate.WithToken(apiToken))
		if err != nil {
			return err
		}

		predictionID, err := parsePredictionID(args[0])
		if err != nil {
			return fmt.Errorf("failed to parse prediction ID: %w", err)
		}
		prediction, err := client.GetPrediction(ctx, predictionID)
		if prediction == nil || err != nil {
			return fmt.Errorf("failed to get prediction: %w", err)
		}

		var directory string
		if len(args) == 2 {
			directory = args[1]
		} else {
			directory = predictionID
		}

		template, _ := cmd.Flags().GetString("template")

		switch template {
		case "node", "nodejs", "js", "":
			return handleNodeTemplate(ctx, prediction, directory, apiToken)
		case "python":
			return handlePythonTemplate(ctx, prediction, directory, apiToken)
		default:
			return fmt.Errorf("unsupported template: %s, expected one of: node, python", template)
		}
	},
}

func init() {
	ScaffoldCmd.Flags().StringP("template", "t", "", "Starter git repo template to use. Currently supported: node, python")
}

// Parse the prediction id from a url, or return the prediction id if it's not a url
func parsePredictionID(value string) (string, error) {
	// Case 1: A prediction ID
	if !strings.Contains(value, "/") {
		return value, nil
	}

	// Case 2: A URL in the form https://replicate.com/p/{id}
	if strings.HasPrefix(value, "replicate.com/p/") || strings.HasPrefix(value, "https://replicate.com/p/") {
		splitURL := strings.Split(value, "/")
		if len(splitURL) == 0 {
			return "", fmt.Errorf("invalid URL format")
		}
		return splitURL[len(splitURL)-1], nil
	}

	// Case 3: A URL in the form https://api.replicate.com/v1/predictions/{id}
	if strings.HasPrefix(value, "api.replicate.com/v1/predictions/") || strings.HasPrefix(value, "https://api.replicate.com/v1/predictions/") {
		splitURL := strings.Split(value, "/")
		if len(splitURL) == 0 {
			return "", fmt.Errorf("invalid URL format")
		}
		return splitURL[len(splitURL)-1], nil
	}

	// Case 4: A URL in the form "https://replicate.com/*?prediction={id}"
	if strings.Contains(value, "replicate.com") || strings.Contains(value, "https://replicate.com") {
		parsedURL, err := url.Parse(value)
		if err != nil {
			return "", fmt.Errorf("failed to parse URL: %w", err)
		}
		queryParams, err := url.ParseQuery(parsedURL.RawQuery)
		if err != nil {
			return "", fmt.Errorf("failed to parse query parameters: %w", err)
		}
		predictionID := queryParams.Get("prediction")
		if predictionID == "" {
			return "", fmt.Errorf("no prediction ID found in URL")
		}
		return predictionID, nil
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

func handleNodeTemplate(ctx context.Context, prediction *replicate.Prediction, directory string, apiToken string) error {
	fmt.Println("Cloning starter repo and installing dependencies...")

	// 1. Clone the starter repo
	if err := execCommand(ctx, fmt.Sprintf("git clone https://github.com/replicate/node-starter.git %s", directory)); err != nil {
		return fmt.Errorf("failed to clone starter repo: %w", err)
	}

	// 2. Set chdir to the output path
	if err := os.Chdir(directory); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	// 3. Install dependencies
	if err := execCommand(ctx, "npm install"); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// 4. Set the REPLICATE_API_TOKEN env var
	if err := execCommand(ctx, fmt.Sprintf(`echo 'REPLICATE_API_TOKEN="%s"' >> .env`, apiToken)); err != nil {
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

func handlePythonTemplate(ctx context.Context, prediction *replicate.Prediction, directory string, apiToken string) error {
	fmt.Println("Cloning starter repo and installing dependencies...")

	// 1. Clone the starter repo
	if err := execCommand(ctx, fmt.Sprintf("git clone git@github.com:replicate/python-starter.git %s", directory)); err != nil {
		return fmt.Errorf("failed to clone starter repo: %w", err)
	}

	// 2. Set chdir to the output path
	err := os.Chdir(directory)
	if err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
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
	if err := execCommand(ctx, fmt.Sprintf(`echo 'REPLICATE_API_TOKEN="%s"' >> .env`, apiToken)); err != nil {
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
