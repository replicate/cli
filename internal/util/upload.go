package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// uploadFile uploads a Zip file to Replicate's experimental DreamBooth API and returns the URL
func UploadFile(ctx context.Context, path string) (string, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Check that file is a zip file
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	fileType := http.DetectContentType(buff)
	if fileType != "application/zip" {
		return "", fmt.Errorf("file is not a zip file")
	}

	// Reset the file pointer
	file.Seek(0, io.SeekStart)

	// Get the upload URL
	request, err := http.NewRequestWithContext(ctx, "POST", "https://dreambooth-api-experimental.replicate.com/v1/upload/data.zip", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	request.Header.Set("Authorization", "Token "+os.Getenv("REPLICATE_API_TOKEN"))
	resp, err := http.DefaultClient.Do(request)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get upload URL: %w", err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return "", fmt.Errorf("failed to read upload response: %w", err)
	}

	uploadResponse := &struct {
		UploadURL  string `json:"upload_url"`
		ServingURL string `json:"serving_url"`
	}{}
	err = json.Unmarshal(bodyBytes, uploadResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse upload response: %w", err)
	}

	// Upload the file
	uploadURL, _ := url.Parse(uploadResponse.UploadURL)
	servingURL, _ := url.Parse(uploadResponse.ServingURL)
	if uploadURL == nil || servingURL == nil {
		return "", fmt.Errorf("failed to parse upload URL: %w", err)
	}

	request, err = http.NewRequestWithContext(ctx, "PUT", uploadURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create file upload request: %w", err)
	}
	request.Header.Set("Content-Type", "application/zip")

	request.Body = file

	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to upload data: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to upload data: %s", resp.Status)
	}

	return servingURL.String(), nil
}
