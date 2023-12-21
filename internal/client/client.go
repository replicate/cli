package client

import (
	"context"
	"fmt"
	"os"

	"github.com/replicate/cli/internal"
	"github.com/replicate/cli/internal/config"
	"github.com/replicate/replicate-go"
)

func NewClient(opts ...replicate.ClientOption) (*replicate.Client, error) {
	token, err := getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get API token: %w", err)
	}

	baseURL := getBaseURL()

	// Validate token when connecting to api.replicate.com.
	// Alternate API hosts proxying Replicate may not require a token.
	if token == "" && baseURL == config.DefaultBaseURL {
		return nil, fmt.Errorf("please authenticate with `replicate auth login`")
	}

	return NewClientWithAPIToken(token, opts...)
}

func NewClientWithAPIToken(token string, opts ...replicate.ClientOption) (*replicate.Client, error) {
	baseURL := getBaseURL()
	userAgent := fmt.Sprintf("replicate-cli/%s", internal.Version())

	opts = append([]replicate.ClientOption{
		replicate.WithBaseURL(baseURL),
		replicate.WithToken(token),
		replicate.WithUserAgent(userAgent),
	}, opts...)

	r8, err := replicate.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return r8, nil
}

func VerifyToken(ctx context.Context, token string) (bool, error) {
	r8, err := NewClientWithAPIToken(token)
	if err != nil {
		return false, err
	}

	// FIXME: Add better endpoint for verifying token
	_, err = r8.ListHardware(ctx)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func getToken() (string, error) {
	token, exists := os.LookupEnv("REPLICATE_API_TOKEN")
	if !exists {
		return config.GetAPIToken()
	}
	return token, nil
}

func getBaseURL() string {
	baseURL, exists := os.LookupEnv("REPLICATE_BASE_URL")
	if !exists {
		baseURL = config.GetAPIBaseURL()
	}
	return baseURL
}
