package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ai-service-go/config"
)

func TestOpenAICompleteJSONSuccess(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": `{"ok":true}`,
					},
				},
			},
		})
	}))
	defer srv.Close()

	client := NewOpenAIClient(&config.Config{
		OpenAIAPIKey:  "test",
		OpenAIBaseURL: srv.URL,
		OpenAIModel:   "gpt-4.1-mini",
		HTTPTimeout:   time.Second,
	})

	raw, err := client.CompleteJSON(context.Background(), "system", map[string]string{"hello": "world"})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if string(raw) != `{"ok":true}` {
		t.Fatalf("unexpected raw response: %s", string(raw))
	}
}

func TestOpenAICompleteJSONHandlesEmptyChoices(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{},
		})
	}))
	defer srv.Close()

	client := NewOpenAIClient(&config.Config{
		OpenAIAPIKey:  "test",
		OpenAIBaseURL: srv.URL,
		OpenAIModel:   "gpt-4.1-mini",
		HTTPTimeout:   time.Second,
	})

	_, err := client.CompleteJSON(context.Background(), "system", map[string]string{"hello": "world"})
	if err == nil {
		t.Fatalf("expected error for empty choices")
	}
}
