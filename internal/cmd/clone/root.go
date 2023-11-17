package clone

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/replicate/replicate-go"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "clone <prediction-ID-or-URL> [<directory>] [--template=<template>]",
	Short: "Setup a new local development environment from a prediction",
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

		var outputClonePath string
		if len(args) == 2 {
			outputClonePath = args[1]
		} else {
			outputClonePath = predictionId
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

		var fullModelString string
		if prediction != nil {
			fullModelString = fmt.Sprintf("%s:%s", prediction.Model, prediction.Version)
		}

		switch template {
		case "node", "":
			return handleNodeTemplate(cmd, prediction, fullModelString, outputClonePath)
		case "python":
			return handlePythonTemplate(cmd, prediction, fullModelString, outputClonePath)
		default:
			return fmt.Errorf("unsupported template: %s, expected one of: node, python", template)
		}
	},
}

func init() {
	RootCmd.Flags().StringP("template", "t", "", "Starter git repo template to use. Currently supported: node, python")
}

// Parse the prediction id from a url, or return the prediction id if it's not a url
func parsePredictionId(predictionish string) (string, error) {
	// Case 1: A prediction ID, which is a base32-encoded string
	if !strings.Contains(predictionish, "/") {
		return predictionish, nil
	}

	// Case 2: A URL in the form https://replicate.com/p/{id}
	if strings.HasPrefix(predictionish, "replicate.com/p/") || strings.HasPrefix(predictionish, "https://replicate.com/p/") {
		splitUrl := strings.Split(predictionish, "/")
		return splitUrl[len(splitUrl)-1], nil
	}

	// Case 3: A URL in the form https://api.replicate.com/v1/predictions/{id}
	if strings.HasPrefix(predictionish, "api.replicate.com/v1/predictions/") || strings.HasPrefix(predictionish, "https://api.replicate.com/v1/predictions/") {
		splitUrl := strings.Split(predictionish, "/")
		return splitUrl[len(splitUrl)-1], nil
	}

	// Case 4: A URL in the form "https://replicate.com/*?prediction={id}"
	if strings.Contains(predictionish, "replicate.com") || strings.Contains(predictionish, "https://replicate.com") {
		parsedUrl, err := url.Parse(predictionish)
		if err != nil {
			return "", fmt.Errorf("failed to parse URL: %w", err)
		}
		predictionId := parsedUrl.Query().Get("prediction")
		if predictionId == "" {
			return "", fmt.Errorf("no prediction ID found in URL")
		}
		return predictionId, nil
	}

	// If none of the above cases match, return an error
	return "", fmt.Errorf("invalid prediction ID or URL format")
}

func handleNodeTemplate(cmd *cobra.Command, prediction *replicate.Prediction, model string, outputClonePath string) error {
	commands := []string{
		fmt.Sprintf("git clone https://github.com/replicate/node-starter.git %s", outputClonePath),
		fmt.Sprintf("cd %s && npm install", outputClonePath),
		fmt.Sprintf(`cd %s && echo 'REPLICATE_API_TOKEN="%s"' >> .env`, outputClonePath, os.Getenv("REPLICATE_API_TOKEN")),
	}

	fmt.Println("Cloning starter repo, and installing dependencies...")

	for _, command := range commands {
		cmd := exec.CommandContext(cmd.Context(), "/bin/sh", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to run command: %w", err)
		}
	}

	// Open the template file
	templateData, err := os.ReadFile(filepath.Join(outputClonePath, "prediction.py.template"))
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Perform string replacement on the template file.
	replacedData := strings.ReplaceAll(string(templateData), "{{MODEL_STRING}}", model)
	inputs, _ := json.Marshal(prediction.Input)
	replacedData = strings.ReplaceAll(replacedData, "{{INPUTS}}", string(inputs))

	// Write the populated template to to outputClonePath/prediction.py
	fmt.Println("Writing new prediction.py...")
	err = os.WriteFile(filepath.Join(outputClonePath, "prediction.py"), []byte(replacedData), 0o644)
	if err != nil {
		return err
	}

	// Run the example prediction
	fmt.Println("Running example prediction...")
	commands = []string{
		fmt.Sprintf("cd %s && node prediction.py", outputClonePath),
	}
	for _, command := range commands {
		cmd := exec.Command("bash", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func handlePythonTemplate(cmd *cobra.Command, prediction *replicate.Prediction, model string, outputClonePath string) error {
	commands := []string{
		fmt.Sprintf("git clone git@github.com:replicate/python-starter.git %s", outputClonePath),
		fmt.Sprintf("cd %s && virtualenv .venv", outputClonePath),
		fmt.Sprintf("cd %s && .venv/bin/pip install -r requirements.txt", outputClonePath),
		fmt.Sprintf(`cd %s && echo 'REPLICATE_API_TOKEN="%s"' >> .env`, outputClonePath, os.Getenv("REPLICATE_API_TOKEN")),
	}

	fmt.Println("Cloning starter repo, and installing dependencies...")

	for _, command := range commands {
		cmd := exec.CommandContext(cmd.Context(), "bash", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to run command: %w", err)
		}
	}

	// Open the template file
	templateData, err := os.ReadFile(filepath.Join(outputClonePath, "prediction.py.template"))
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Perform string replacement on the template file.
	replacedData := strings.ReplaceAll(string(templateData), "{{MODEL_STRING}}", model)
	inputs, _ := json.Marshal(prediction.Input)
	replacedData = strings.ReplaceAll(replacedData, "{{INPUTS}}", string(inputs))

	// Write the populated template to to outputClonePath/prediction.py
	fmt.Println("Writing new prediction.py...")
	err = os.WriteFile(filepath.Join(outputClonePath, "prediction.py"), []byte(replacedData), 0o644)
	if err != nil {
		return err
	}

	// Run the example prediction
	fmt.Println("Running example prediction...")
	commands = []string{
		fmt.Sprintf("cd %s && .venv/bin/python prediction.py", outputClonePath),
	}
	for _, command := range commands {
		cmd := exec.Command("bash", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
