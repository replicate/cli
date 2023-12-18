package client

import (
	"context"
	"fmt"

	"github.com/replicate/cli/internal"
	"github.com/replicate/cli/internal/config"
	"github.com/replicate/replicate-go"
)

func NewClient(opts ...replicate.ClientOption) (*replicate.Client, error) {
	baseURL := config.GetAPIBaseURL()

	token, err := config.GetAPIToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get API token: %w", err)
	}
	if token == "" {
		return nil, fmt.Errorf("please authenticate with `replicate login`")
	}

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
	r8, err := NewClient(replicate.WithToken(token))
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
