package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ai-service-go/config"
	"ai-service-go/internal/dto"
)

type ExaClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewExaClient(cfg *config.Config) *ExaClient {
	return &ExaClient{
		baseURL:    strings.TrimRight(cfg.ExaBaseURL, "/"),
		apiKey:     cfg.ExaAPIKey,
		httpClient: &http.Client{Timeout: cfg.HTTPTimeout},
	}
}

func (c *ExaClient) Search(ctx context.Context, query string, numResults int) ([]dto.ExaSearchResult, error) {
	if c.apiKey == "" {
		return nil, nil
	}

	reqBody := dto.ExaSearchRequest{
		Query:      query,
		NumResults: numResults,
		Type:       "auto",
		Contents: dto.ExaContents{
			Highlights: dto.ExaHighlights{MaxCharacters: 2000},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/search", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("exa returned status %d", resp.StatusCode)
	}

	var parsed dto.ExaSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	return parsed.Results, nil
}
