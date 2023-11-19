package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// uploadFile uploads a file to Replicate's experimental DreamBooth API and returns the URL
func UploadFile(ctx context.Context, path string) (string, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get the upload URL
	filename := filepath.Base(path)
	request, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://dreambooth-api-experimental.replicate.com/v1/upload/%s", filename), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	request.Header.Set("Authorization", "Token "+os.Getenv("REPLICATE_API_TOKEN"))
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to get upload URL: %w", err)
	}
	if resp == nil {
		return "", fmt.Errorf("failed to get upload URL")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get upload URL: %s", resp.Status)
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

	// Detect the content type
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	contentType := http.DetectContentType(buff)
	request.Header.Set("Content-Type", contentType)

	// Reset the file pointer
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("failed to reset file pointer: %w", err)
	}

	request.Body = file

	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to upload data: %w", err)
	}
	if resp == nil {
		return "", fmt.Errorf("failed to upload data")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to upload data: %s", resp.Status)
	}

	return servingURL.String(), nil
}
