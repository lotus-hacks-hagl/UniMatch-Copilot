package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"ai-service-go/config"
	"ai-service-go/internal/dto"
	"ai-service-go/internal/provider"
)

func TestComputeEvidenceCoverage(t *testing.T) {
	t.Parallel()

	evidence := searchEvidence{
		Attempts: []dto.SearchAttempt{
			{Query: "a", ResultCount: 0},
			{Query: "b", ResultCount: 2},
		},
		Results: []dto.ExaSearchResult{
			{URL: "https://one.test"},
			{URL: "https://two.test"},
			{URL: "https://three.test"},
		},
		URLs: []string{"https://one.test", "https://two.test"},
		Tinyfish: []map[string]interface{}{
			{"majors": []string{"Computer Science"}},
		},
	}

	got := computeEvidenceCoverage(evidence)
	if got <= 0 {
		t.Fatalf("expected positive coverage, got %v", got)
	}
	if got > 1 {
		t.Fatalf("expected coverage <= 1, got %v", got)
	}
}

func TestMergeCrawlResult(t *testing.T) {
	t.Parallel()

	base := &dto.CrawlResult{
		Name:            "Base University",
		Country:         "USA",
		CrawlStatus:     "ok",
		SourceURLs:      []string{"https://base.test"},
		AvailableMajors: []string{"Engineering"},
	}
	filled := &dto.CrawlResult{
		Name:            "Filled University",
		AvailableMajors: []string{"Computer Science"},
		SourceURLs:      []string{"https://filled.test"},
	}

	got := mergeCrawlResult(base, filled)
	if got.Name != "Filled University" {
		t.Fatalf("expected merged name, got %q", got.Name)
	}
	if len(got.AvailableMajors) != 1 || got.AvailableMajors[0] != "Computer Science" {
		t.Fatalf("expected merged majors, got %#v", got.AvailableMajors)
	}
	if len(got.SourceURLs) != 1 || got.SourceURLs[0] != "https://filled.test" {
		t.Fatalf("expected merged source URLs, got %#v", got.SourceURLs)
	}
}

func TestBuildAnalyzeResultEmptyCandidatesEscalates(t *testing.T) {
	t.Parallel()

	svc := newTestJobService(t, providerBehavior{}, providerBehavior{}, openAIBehavior{})
	result, err := svc.buildAnalyzeResult(context.Background(), dto.AnalyzeJobRequest{
		JobID:       "job-empty",
		CaseID:      "case-empty",
		CallbackURL: "http://callback.test",
		Input: dto.AnalyzeInput{
			FullName:              "Empty Candidate Student",
			IntendedMajor:         "Computer Science",
			CandidateUniversities: []dto.CandidateUniversity{},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.EscalationNeeded {
		t.Fatalf("expected escalation for empty candidates")
	}
	if len(result.Recommendations) != 0 {
		t.Fatalf("expected no recommendations, got %d", len(result.Recommendations))
	}
}

func TestBuildAnalyzeResultWithoutProviderEvidenceOrOpenAIReturnsNoRecommendations(t *testing.T) {
	t.Parallel()

	svc := newTestJobService(t, providerBehavior{}, providerBehavior{}, openAIBehavior{})
	svc.openai = provider.NewOpenAIClient(&config.Config{
		OpenAIAPIKey:  "",
		OpenAIBaseURL: "http://unused.test",
		OpenAIModel:   "gpt-4.1-mini",
		HTTPTimeout:   time.Second,
	})

	result, err := svc.buildAnalyzeResult(context.Background(), dto.AnalyzeJobRequest{
		JobID:       "job-no-evidence",
		CaseID:      "case-no-evidence",
		CallbackURL: "http://callback.test",
		Input: dto.AnalyzeInput{
			FullName:           "No Evidence Student",
			GpaNormalized:      3.6,
			IntendedMajor:      "Computer Science",
			BudgetUsdPerYear:   30000,
			PreferredCountries: []string{"USA"},
			TargetIntake:       "Fall 2026",
			CandidateUniversities: []dto.CandidateUniversity{
				{
					UniversityID:   "u-1",
					UniversityName: "Alpha University",
					Country:        "USA",
					AvailableMajors: []string{
						"Computer Science",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Recommendations) != 0 {
		t.Fatalf("expected no recommendations when no provider evidence and no OpenAI, got %d", len(result.Recommendations))
	}
	if !result.EscalationNeeded {
		t.Fatalf("expected escalation when no provider evidence and no OpenAI")
	}
	if result.ConfidenceScore >= 0.1 {
		t.Fatalf("expected very low confidence, got %v", result.ConfidenceScore)
	}
}

func TestAnalyzeEndpointTracksFiveAttemptsAndUsesOpenAI(t *testing.T) {
	exaBehavior := providerBehavior{
		handler: func(w http.ResponseWriter, r *http.Request, count int) {
			io.WriteString(w, `{"results":[]}`)
		},
	}
	openAIBehavior := openAIBehavior{
		responses: []string{
			`{"profile_summary":{"mode":"openai"},"recommendations":[{"university_id":"u-1","university_name":"Alpha University","tier":"match","admission_likelihood_score":72,"student_fit_score":75,"reason":"OpenAI fill after empty search.","risks":["Limited evidence"],"improvements":["Improve essays"],"rank_order":1}],"confidence_score":0.62,"escalation_needed":false,"escalation_reason":""}`,
		},
	}

	callbacks := make(chan dto.JobDonePayload, 1)
	callbackSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload dto.JobDonePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode callback payload: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		callbacks <- payload
		w.WriteHeader(http.StatusOK)
	}))
	defer callbackSrv.Close()

	svc := newTestJobService(t, exaBehavior, providerBehavior{}, openAIBehavior)
	if appErr := svc.EnqueueAnalyze(context.Background(), dto.AnalyzeJobRequest{
		JobID:       "job-analyze-1",
		CaseID:      "case-analyze-1",
		CallbackURL: callbackSrv.URL,
		Input: dto.AnalyzeInput{
			FullName:           "Analyze Student",
			GpaNormalized:      3.5,
			IntendedMajor:      "Computer Science",
			BudgetUsdPerYear:   30000,
			PreferredCountries: []string{"USA"},
			TargetIntake:       "Fall 2026",
			CandidateUniversities: []dto.CandidateUniversity{
				{
					UniversityID:   "u-1",
					UniversityName: "Alpha University",
					Country:        "USA",
					AvailableMajors: []string{
						"Computer Science",
					},
				},
			},
		},
	}); appErr != nil {
		t.Fatalf("enqueue analyze: %v", appErr)
	}

	select {
	case payload := <-callbacks:
		if payload.JobID != "job-analyze-1" {
			t.Fatalf("unexpected callback job id: %s", payload.JobID)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for analyze callback")
	}

	job := waitForJobState(t, svc, "job-analyze-1")
	if len(job.SearchAttempts) != 5 {
		t.Fatalf("expected 5 search attempts, got %d", len(job.SearchAttempts))
	}
	if !job.OpenAIFillUsed {
		t.Fatalf("expected openai fill to be used")
	}
	if job.CallbackStatus != "delivered" {
		t.Fatalf("expected delivered callback status, got %q", job.CallbackStatus)
	}
}

func TestCrawlEndpointRetriesOpenAIUntilValidSchema(t *testing.T) {
	exaBehavior := providerBehavior{
		handler: func(w http.ResponseWriter, r *http.Request, count int) {
			io.WriteString(w, `{"results":[]}`)
		},
	}
	openAIBehavior := openAIBehavior{
		responses: []string{
			`not-json`,
			`{}`,
			`{"name":"Retry University","country":"Canada","crawl_status":"ok","source_urls":["https://retry.test"],"available_majors":["Computer Science"]}`,
		},
	}

	callbacks := make(chan dto.JobDonePayload, 1)
	callbackSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload dto.JobDonePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode callback payload: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		callbacks <- payload
		w.WriteHeader(http.StatusOK)
	}))
	defer callbackSrv.Close()

	svc := newTestJobService(t, exaBehavior, providerBehavior{}, openAIBehavior)
	if appErr := svc.EnqueueCrawl(context.Background(), dto.CrawlJobRequest{
		JobID:        "job-crawl-1",
		UniversityID: "uni-1",
		CallbackURL:  callbackSrv.URL,
		Metadata: map[string]interface{}{
			"name":    "Retry University",
			"country": "Canada",
		},
	}); appErr != nil {
		t.Fatalf("enqueue crawl: %v", appErr)
	}

	select {
	case payload := <-callbacks:
		if payload.JobID != "job-crawl-1" {
			t.Fatalf("unexpected callback job id: %s", payload.JobID)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for crawl callback")
	}

	job := waitForJobState(t, svc, "job-crawl-1")
	if job.OpenAIAttempts != 3 {
		t.Fatalf("expected 3 openai attempts, got %d", job.OpenAIAttempts)
	}
	if job.CallbackStatus != "delivered" {
		t.Fatalf("expected delivered callback status, got %q", job.CallbackStatus)
	}
}

func TestReportEndpointRetriesCallback(t *testing.T) {
	var mu sync.Mutex
	attempts := 0
	callbackSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		attempts++
		current := attempts
		mu.Unlock()
		if current < 3 {
			http.Error(w, "retry me", http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer callbackSrv.Close()

	svc := newTestJobService(t, providerBehavior{}, providerBehavior{}, openAIBehavior{})
	if appErr := svc.EnqueueReport(context.Background(), dto.ReportJobRequest{
		JobID:       "job-report-1",
		CaseID:      "case-report-1",
		CallbackURL: callbackSrv.URL,
		StudentName: "Retry Student",
		Recommendations: []dto.RecommendationResult{
			{
				UniversityID:             "u-1",
				UniversityName:           "Alpha University",
				Tier:                     "safe",
				AdmissionLikelihoodScore: 80,
				StudentFitScore:          82,
				Reason:                   "Strong fit.",
				Risks:                    []string{"Competition"},
				Improvements:             []string{"Polish essay"},
				RankOrder:                1,
			},
		},
	}); appErr != nil {
		t.Fatalf("enqueue report: %v", appErr)
	}

	job := waitForJobState(t, svc, "job-report-1")
	if job.CallbackAttempts != 3 {
		t.Fatalf("expected 3 callback attempts, got %d", job.CallbackAttempts)
	}
	if job.CallbackStatus != "delivered" {
		t.Fatalf("expected delivered callback status, got %q", job.CallbackStatus)
	}
}

type providerBehavior struct {
	handler func(w http.ResponseWriter, r *http.Request, count int)
}

type openAIBehavior struct {
	responses []string
}

func newTestJobService(t *testing.T, exa providerBehavior, tinyfish providerBehavior, openai openAIBehavior) *jobService {
	t.Helper()

	exaCount := 0
	exaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exaCount++
		if exa.handler != nil {
			exa.handler(w, r, exaCount)
			return
		}
		io.WriteString(w, `{"results":[]}`)
	}))
	t.Cleanup(exaSrv.Close)

	tinyfishCount := 0
	tinyfishSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tinyfishCount++
		if tinyfish.handler != nil {
			tinyfish.handler(w, r, tinyfishCount)
			return
		}
		io.WriteString(w, `{"data":{"majors":["Computer Science"]}}`)
	}))
	t.Cleanup(tinyfishSrv.Close)

	openAICount := 0
	openAISrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		openAICount++
		content := `{"profile_summary":{"source":"default"},"recommendations":[],"confidence_score":0.5,"escalation_needed":true,"escalation_reason":"default"}`
		if openAICount <= len(openai.responses) {
			content = openai.responses[openAICount-1]
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": content,
					},
				},
			},
		})
	}))
	t.Cleanup(openAISrv.Close)

	cfg := &config.Config{
		Port:                "8895",
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

	svc := NewJobService(cfg, provider.NewExaClient(cfg), provider.NewTinyfishClient(cfg), provider.NewOpenAIClient(cfg))
	impl, ok := svc.(*jobService)
	if !ok {
		t.Fatalf("expected *jobService")
	}
	return impl
}

func waitForJobState(t *testing.T, svc *jobService, jobID string) dto.JobDebugResponse {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		job, appErr := svc.GetJob(context.Background(), jobID)
		if appErr == nil {
			if job.CallbackStatus == "delivered" || job.CallbackStatus == "failed" {
				return *job
			}
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for job state %s", jobID)
	return dto.JobDebugResponse{}
}
