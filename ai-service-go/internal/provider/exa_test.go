package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"ai-service-go/config"
)

// ── Unit Tests (mock server) ──────────────────────────────────────────────────

func TestExaSearch_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request shape
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("x-api-key") == "" {
			t.Error("missing x-api-key header")
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{
					"title": "Test Result",
					"url":   "https://example.com",
					"score": 0.95,
				},
			},
		})
	}))
	defer srv.Close()

	client := NewExaClient(&config.Config{
		ExaAPIKey:   "test-key",
		ExaBaseURL:  srv.URL,
		HTTPTimeout: time.Second,
	})

	results, err := client.Search(context.Background(), "golang jobs", 1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
}

func TestExaSearch_SkipsWhenNoAPIKey(t *testing.T) {
	t.Parallel()

	client := NewExaClient(&config.Config{
		ExaAPIKey:   "", // empty key
		ExaBaseURL:  "https://api.exa.ai",
		HTTPTimeout: time.Second,
	})

	results, err := client.Search(context.Background(), "test", 3)
	if err != nil {
		t.Fatalf("expected nil error when key is empty, got %v", err)
	}
	if results != nil {
		t.Fatal("expected nil results when key is empty")
	}
}

func TestExaSearch_HandlesNon2xxStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid api key"}`))
	}))
	defer srv.Close()

	client := NewExaClient(&config.Config{
		ExaAPIKey:   "bad-key",
		ExaBaseURL:  srv.URL,
		HTTPTimeout: time.Second,
	})

	_, err := client.Search(context.Background(), "test", 1)
	if err == nil {
		t.Fatal("expected error on 401, got nil")
	}
}

func TestExaSearch_HandlesInvalidJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not-valid-json`))
	}))
	defer srv.Close()

	client := NewExaClient(&config.Config{
		ExaAPIKey:   "test-key",
		ExaBaseURL:  srv.URL,
		HTTPTimeout: time.Second,
	})

	_, err := client.Search(context.Background(), "test", 1)
	if err == nil {
		t.Fatal("expected JSON parse error, got nil")
	}
}

func TestExaSearch_ContextCancellation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than context deadline; return when either request ctx or server shuts down
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	defer srv.Close()

	client := NewExaClient(&config.Config{
		ExaAPIKey:   "test-key",
		ExaBaseURL:  srv.URL,
		HTTPTimeout: 5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.Search(ctx, "test", 1)
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

// ── Integration Test (calls real Exa API) ────────────────────────────────────
// Run with: go test ./internal/provider/... -run Integration -v -timeout 30s

func TestExaSearch_Integration(t *testing.T) {
	apiKey := os.Getenv("EXA_API_KEY")
	if apiKey == "" {
		t.Skip("EXA_API_KEY not set, skipping integration test")
	}

	baseURL := os.Getenv("EXA_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.exa.ai"
	}

	client := NewExaClient(&config.Config{
		ExaAPIKey:   apiKey,
		ExaBaseURL:  baseURL,
		HTTPTimeout: 15 * time.Second,
	})

	results, err := client.Search(context.Background(), "software engineer jobs vietnam 2024", 3)
	if err != nil {
		t.Fatalf("❌ Exa API error: %v", err)
	}

	t.Logf("✅ Exa returned %d results", len(results))
	for i, r := range results {
		t.Logf("  [%d] title=%q url=%s", i+1, r.Title, r.URL)
	}

	if len(results) == 0 {
		t.Log("⚠️  No results returned (valid but unexpected)")
	}
}
