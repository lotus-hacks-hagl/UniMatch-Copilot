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

type OpenAIClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

type openAIChatCompletionRequest struct {
	Model          string               `json:"model"`
	Temperature    float64              `json:"temperature"`
	ResponseFormat openAIResponseFormat `json:"response_format"`
	Messages       []openAIChatMessage  `json:"messages"`
}

type openAIResponseFormat struct {
	Type string `json:"type"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewOpenAIClient(cfg *config.Config) *OpenAIClient {
	return &OpenAIClient{
		baseURL:    strings.TrimRight(cfg.OpenAIBaseURL, "/"),
		apiKey:     cfg.OpenAIAPIKey,
		model:      cfg.OpenAIModel,
		httpClient: &http.Client{Timeout: cfg.HTTPTimeout},
	}
}

func (c *OpenAIClient) Enabled() bool {
	return c != nil && c.apiKey != ""
}

func (c *OpenAIClient) CompleteJSON(ctx context.Context, systemPrompt string, payload interface{}) (json.RawMessage, error) {
	if !c.Enabled() {
		return nil, nil
	}

	userBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	reqBody := openAIChatCompletionRequest{
		Model:       c.model,
		Temperature: 0.7,
		ResponseFormat: openAIResponseFormat{
			Type: "json_object",
		},
		Messages: []openAIChatMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: string(userBody),
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openai returned status %d", resp.StatusCode)
	}

	var parsed openAIChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return nil, fmt.Errorf("openai returned empty content")
	}

	return json.RawMessage(parsed.Choices[0].Message.Content), nil
}
