package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ai-service-go/config"
)

type TinyfishClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type TinyfishRunRequest struct {
	URL  string `json:"url"`
	Goal string `json:"goal"`
}

func NewTinyfishClient(cfg *config.Config) *TinyfishClient {
	return &TinyfishClient{
		baseURL:    strings.TrimRight(cfg.TinyfishBaseURL, "/"),
		apiKey:     cfg.TinyfishAPIKey,
		httpClient: &http.Client{Timeout: cfg.HTTPTimeout},
	}
}

func (c *TinyfishClient) Run(ctx context.Context, url, goal string) (json.RawMessage, error) {
	if c.apiKey == "" {
		return nil, nil
	}

	reqBody := TinyfishRunRequest{
		URL:  url,
		Goal: goal,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/automation/run", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tinyfish returned status %d", resp.StatusCode)
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	if data, ok := raw["data"]; ok {
		return data, nil
	}
	if result, ok := raw["result"]; ok {
		return result, nil
	}

	fallback, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	return fallback, nil
}
