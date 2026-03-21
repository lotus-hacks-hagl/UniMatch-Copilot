package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ai-service-go/config"
	"ai-service-go/internal/dto"
	"ai-service-go/internal/handler"
	"ai-service-go/internal/provider"
	"ai-service-go/internal/router"
	"ai-service-go/internal/service"
)

func TestJobsRouterExposesDebugEndpointInTestEnv(t *testing.T) {
	exaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"results":[]}`)
	}))
	defer exaSrv.Close()

	tinyfishSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"data":{"majors":["Computer Science"]}}`)
	}))
	defer tinyfishSrv.Close()

	openAISrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": `{"profile_summary":{"mode":"endpoint"},"recommendations":[{"university_id":"u-1","university_name":"Alpha University","tier":"match","admission_likelihood_score":70,"student_fit_score":72,"reason":"Filled through endpoint integration test.","risks":["Limited evidence"],"improvements":["Improve essays"],"rank_order":1}],"confidence_score":0.6,"escalation_needed":false,"escalation_reason":""}`,
					},
				},
			},
		})
	}))
	defer openAISrv.Close()

	callbacks := make(chan struct{}, 1)
	callbackSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callbacks <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	defer callbackSrv.Close()

	cfg := &config.Config{
		Port:                "9000",
		Env:                 "test",
		ExaAPIKey:           "exa-test",
		TinyfishAPIKey:      "tinyfish-test",
		OpenAIAPIKey:        "openai-test",
		ExaBaseURL:          exaSrv.URL,
		TinyfishBaseURL:     tinyfishSrv.URL,
		OpenAIBaseURL:       openAISrv.URL,
		OpenAIModel:         "gpt-4.1-mini",
		HTTPTimeout:         time.Second,
		CallbackTimeout:     time.Second,
		MaxCandidates:       24,
		MaxRecommendations:  6,
		MaxSearchAttempts:   5,
		MaxDetailFetches:    3,
		OpenAIRetryAttempts: 5,
		CallbackRetryCount:  3,
		CallbackRetryDelay:  10 * time.Millisecond,
		FallbackEnabled:     true,
		AllowOpenAIFill:     true,
	}

	svc := service.NewJobService(cfg, provider.NewExaClient(cfg), provider.NewTinyfishClient(cfg), provider.NewOpenAIClient(cfg))
	jobsH := handler.NewJobsHandler(svc)
	r := router.SetupRouter(cfg, jobsH)

	body := map[string]interface{}{
		"job_id":       "job-endpoint-1",
		"case_id":      "case-endpoint-1",
		"callback_url": callbackSrv.URL,
		"input": map[string]interface{}{
			"full_name":           "Endpoint Student",
			"gpa_normalized":      3.6,
			"intended_major":      "Computer Science",
			"budget_usd_per_year": 30000,
			"preferred_countries": []string{"USA"},
			"target_intake":       "Fall 2026",
			"candidate_universities": []map[string]interface{}{
				{
					"university_id":   "u-1",
					"university_name": "Alpha University",
					"country":         "USA",
					"available_majors": []string{
						"Computer Science",
					},
				},
			},
		},
	}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/jobs/analyze", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.Code)
	}

	select {
	case <-callbacks:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for callback")
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		req = httptest.NewRequest(http.MethodGet, "/jobs/job-endpoint-1", nil)
		resp = httptest.NewRecorder()
		r.ServeHTTP(resp, req)
		if resp.Code == http.StatusOK {
			var parsed struct {
				Data dto.JobDebugResponse `json:"data"`
			}
			if err := json.Unmarshal(resp.Body.Bytes(), &parsed); err == nil {
				if parsed.Data.CallbackStatus == "delivered" {
					if len(parsed.Data.SearchAttempts) != 5 {
						t.Fatalf("expected 5 attempts, got %d", len(parsed.Data.SearchAttempts))
					}
					return
				}
			}
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("timed out waiting for debug endpoint state")
}
