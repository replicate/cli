package client

import (
	"fmt"

	"github.com/replicate/cli/internal"
	"github.com/replicate/cli/internal/config"
	"github.com/replicate/replicate-go"
)

func NewClient() (*replicate.Client, error) {
	baseURL := config.GetAPIBaseURL()

	token, err := config.GetAPIToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get API token: %w", err)
	}
	if token == "" {
		return nil, fmt.Errorf("please authenticate with `replicate login`")
	}

	userAgent := fmt.Sprintf("replicate/%s", internal.Version())

	opts := []replicate.ClientOption{
		replicate.WithBaseURL(baseURL),
		replicate.WithToken(token),
		replicate.WithUserAgent(userAgent),
	}

	r8, err := replicate.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return r8, nil
}
