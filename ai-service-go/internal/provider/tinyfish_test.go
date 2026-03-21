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

func TestTinyfishRun_Success_DataField(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("X-API-Key") == "" {
			t.Error("missing X-API-Key header")
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"name":    "MIT",
				"country": "USA",
			},
		})
	}))
	defer srv.Close()

	client := NewTinyfishClient(&config.Config{
		TinyfishAPIKey:  "test-key",
		TinyfishBaseURL: srv.URL,
		HTTPTimeout:     time.Second,
	})

	raw, err := client.Run(context.Background(), "https://example.com", "get university info")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if raw == nil {
		t.Fatal("expected non-nil result")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	if result["name"] != "MIT" {
		t.Errorf("expected name=MIT, got %v", result["name"])
	}
}

func TestTinyfishRun_Success_ResultField(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"qs_rank": 5,
			},
		})
	}))
	defer srv.Close()

	client := NewTinyfishClient(&config.Config{
		TinyfishAPIKey:  "test-key",
		TinyfishBaseURL: srv.URL,
		HTTPTimeout:     time.Second,
	})

	raw, err := client.Run(context.Background(), "https://example.com", "get rank")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	var result map[string]interface{}
	_ = json.Unmarshal(raw, &result)
	if result["qs_rank"] == nil {
		t.Error("expected qs_rank in result")
	}
}

func TestTinyfishRun_Success_FallbackField(t *testing.T) {
	t.Parallel()

	// Response với field không phải data/result → fallback toàn bộ map
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"custom_field": "custom_value",
		})
	}))
	defer srv.Close()

	client := NewTinyfishClient(&config.Config{
		TinyfishAPIKey:  "test-key",
		TinyfishBaseURL: srv.URL,
		HTTPTimeout:     time.Second,
	})

	raw, err := client.Run(context.Background(), "https://example.com", "test")
	if err != nil {
		t.Fatalf("expected fallback success, got %v", err)
	}
	if raw == nil {
		t.Fatal("expected non-nil fallback result")
	}
}

func TestTinyfishRun_SkipsWhenNoAPIKey(t *testing.T) {
	t.Parallel()

	client := NewTinyfishClient(&config.Config{
		TinyfishAPIKey:  "", // empty
		TinyfishBaseURL: "https://agent.tinyfish.ai",
		HTTPTimeout:     time.Second,
	})

	raw, err := client.Run(context.Background(), "https://example.com", "test")
	if err != nil {
		t.Fatalf("expected nil error when key is empty, got %v", err)
	}
	if raw != nil {
		t.Fatal("expected nil result when key is empty")
	}
}

func TestTinyfishRun_HandlesNon2xxStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"code":"FORBIDDEN","message":"Insufficient credits"}}`))
	}))
	defer srv.Close()

	client := NewTinyfishClient(&config.Config{
		TinyfishAPIKey:  "depleted-key",
		TinyfishBaseURL: srv.URL,
		HTTPTimeout:     time.Second,
	})

	_, err := client.Run(context.Background(), "https://example.com", "test")
	if err == nil {
		t.Fatal("expected error on 403, got nil")
	}
	t.Logf("got expected error: %v", err)
}

func TestTinyfishRun_HandlesTimeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	defer srv.Close()

	client := NewTinyfishClient(&config.Config{
		TinyfishAPIKey:  "test-key",
		TinyfishBaseURL: srv.URL,
		HTTPTimeout:     time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.Run(ctx, "https://example.com", "test")
	if err == nil {
		t.Fatal("expected timeout/context error, got nil")
	}
}

// ── Integration Test (calls real Tinyfish API) ────────────────────────────────
// Run with: go test ./internal/provider/... -run Integration -v -timeout 60s

func TestTinyfishRun_Integration(t *testing.T) {
	apiKey := os.Getenv("TINYFISH_API_KEY")
	if apiKey == "" {
		t.Skip("TINYFISH_API_KEY not set, skipping integration test")
	}

	baseURL := os.Getenv("TINYFISH_BASE_URL")
	if baseURL == "" {
		baseURL = "https://agent.tinyfish.ai"
	}

	client := NewTinyfishClient(&config.Config{
		TinyfishAPIKey:  apiKey,
		TinyfishBaseURL: baseURL,
		HTTPTimeout:     30 * time.Second,
	})

	raw, err := client.Run(
		context.Background(),
		"https://www.mit.edu/",
		"Get the university name and country",
	)
	if err != nil {
		t.Fatalf("❌ Tinyfish API error: %v", err)
	}

	t.Logf("✅ Tinyfish responded: %s", string(raw))
}
