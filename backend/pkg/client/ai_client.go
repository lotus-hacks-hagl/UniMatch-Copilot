package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAIClient(baseURL string) *AIClient {
	return &AIClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// CrawlJobRequest matches shared contract with AI Service
type CrawlJobRequest struct {
	JobID        string                 `json:"job_id"`
	UniversityID string                 `json:"university_id"`
	CallbackURL    string                 `json:"callback_url"`
	IsMetadataOnly bool                   `json:"is_metadata_only"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// AnalyzeJobRequest matches shared contract with AI Service
type AnalyzeJobRequest struct {
	JobID       string       `json:"job_id"`
	CaseID      string       `json:"case_id"`
	CallbackURL string       `json:"callback_url"`
	Input       AnalyzeInput `json:"input"`
}

type AnalyzeInput struct {
	FullName              string                `json:"full_name"`
	GpaNormalized         float64               `json:"gpa_normalized"`
	IeltsOverall          *float64              `json:"ielts_overall"`
	SatTotal              *int                  `json:"sat_total"`
	ToeflTotal            *int                  `json:"toefl_total,omitempty"`
	IntendedMajor         string                `json:"intended_major"`
	BudgetUsdPerYear      int                   `json:"budget_usd_per_year"`
	PreferredCountries    []string              `json:"preferred_countries"`
	TargetIntake          string                `json:"target_intake"`
	ScholarshipRequired   bool                  `json:"scholarship_required"`
	Extracurriculars      string                `json:"extracurriculars"`
	Achievements          string                `json:"achievements"`
}

// ReportJobRequest matches shared contract with AI Service
type ReportJobRequest struct {
	JobID           string        `json:"job_id"`
	CaseID          string        `json:"case_id"`
	CallbackURL     string        `json:"callback_url"`
	StudentName     string        `json:"student_name"`
	Recommendations []interface{} `json:"recommendations"`
}

func (c *AIClient) SubmitCrawlJob(req CrawlJobRequest) error {
	return c.post("/jobs/crawl", req)
}

func (c *AIClient) SubmitAnalyzeJob(req AnalyzeJobRequest) error {
	return c.post("/jobs/analyze", req)
}

func (c *AIClient) SubmitReportJob(req ReportJobRequest) error {
	return c.post("/jobs/report", req)
}

func (c *AIClient) DeleteUniversity(universityID string) error {
	req, err := http.NewRequest(http.MethodDelete, c.baseURL+"/jobs/university/"+universityID, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ai service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ai service returned status %d", resp.StatusCode)
	}
	return nil
}

func (c *AIClient) post(path string, body interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+path, "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Printf("[AIClient] POST %s failed: %v\n", c.baseURL+path, err)
		return fmt.Errorf("ai service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("[AIClient] POST %s returned status %d\n", c.baseURL+path, resp.StatusCode)
		return fmt.Errorf("ai service returned status %d", resp.StatusCode)
	}
	return nil
}
