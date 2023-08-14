package util

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

// ExplorePageListing represents a model listing in the JSON embedded in replicate.com/explore
//
// This is a temporary workaround until Replicate's API provides a `models.list` endpoint.
type ExplorePageListing struct {
	AbsoluteURL  string `json:"absolute_url"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Visibility   string `json:"visibility"`
	IsRunOnly    bool   `json:"is_run_only"`
	GitHubURL    string `json:"github_url"`
	PaperURL     string `json:"paper_url"`
	ArxivPaperID string `json:"arxiv_paper_id"`
	CoverImage   struct {
		URL             string `json:"url"`
		FileType        string `json:"file_type"`
		FileMimetype    string `json:"file_mimetype"`
		ModelIsPlayable bool   `json:"model_is_playable"`
	} `json:"cover_image"`
	LatestVersionCreatedAt string `json:"latest_version_created_at"`
	DefaultExampleUUID     string `json:"default_example_uuid"`
	DisplayOutputAsJSON    bool   `json:"display_output_as_json"`
}

func (e ExplorePageListing) String() string {
	return fmt.Sprintf("%s/%s", e.Username, e.Name)
}

func (e ExplorePageListing) URL() string {
	return fmt.Sprintf("https://replicate.com/%s", e.String())
}

// ListModelsOnExplorePage returns the list of models on replicate.com/explore
func ListModelsOnExplorePage(ctx context.Context) (*[]ExplorePageListing, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://replicate.com/explore", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create explore request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get explore page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse explore page: %w", err)
	}

	script := doc.Find("script#models")
	if script == nil {
		return nil, fmt.Errorf("failed to find script with id=models")
	}

	jsonString := script.Text()
	if jsonString == "" {
		return nil, fmt.Errorf("failed to extract JSON from script")
	}

	var models []ExplorePageListing
	err = json.Unmarshal([]byte(jsonString), &models)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &models, nil
}
