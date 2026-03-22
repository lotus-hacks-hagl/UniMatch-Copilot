package dto

import (
	"encoding/json"
	"time"
)

type CandidateUniversity struct {
	UniversityID             string   `json:"university_id"`
	UniversityName           string   `json:"university_name"`
	Country                  string   `json:"country"`
	QsRank                   *int     `json:"qs_rank"`
	IeltsMin                 *float64 `json:"ielts_min"`
	SatRequired              bool     `json:"sat_required"`
	GpaExpectationNormalized *float64 `json:"gpa_expectation_normalized"`
	TuitionUsdPerYear        *int     `json:"tuition_usd_per_year"`
	ScholarshipAvailable     bool     `json:"scholarship_available"`
	AvailableMajors          []string `json:"available_majors"`
	AcceptanceRate           *float64 `json:"acceptance_rate"`
}

type CrawlJobRequest struct {
	JobID        string                 `json:"job_id" binding:"required"`
	UniversityID string                 `json:"university_id" binding:"required"`
	CallbackURL  string                 `json:"callback_url" binding:"required"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type AnalyzeJobRequest struct {
	JobID       string       `json:"job_id" binding:"required"`
	CaseID      string       `json:"case_id" binding:"required"`
	CallbackURL string       `json:"callback_url" binding:"required"`
	Input       AnalyzeInput `json:"input" binding:"required"`
}

type AnalyzeInput struct {
	FullName              string                `json:"full_name"`
	GpaNormalized         float64               `json:"gpa_normalized"`
	IeltsOverall          *float64              `json:"ielts_overall"`
	SatTotal              *int                  `json:"sat_total"`
	ToeflTotal            *int                  `json:"toefl_total"`
	IntendedMajor         string                `json:"intended_major"`
	BudgetUsdPerYear      int                   `json:"budget_usd_per_year"`
	PreferredCountries    []string              `json:"preferred_countries"`
	TargetIntake          string                `json:"target_intake"`
	ScholarshipRequired   bool                  `json:"scholarship_required"`
	Extracurriculars      string                `json:"extracurriculars"`
	Achievements          string                `json:"achievements"`
	BackgroundText        string                `json:"background_text"`
	CandidateUniversities []CandidateUniversity `json:"candidate_universities"`
}

type ReportJobRequest struct {
	JobID           string                 `json:"job_id" binding:"required"`
	CaseID          string                 `json:"case_id" binding:"required"`
	CallbackURL     string                 `json:"callback_url" binding:"required"`
	StudentName     string                 `json:"student_name"`
	Recommendations []RecommendationResult `json:"recommendations"`
}

type JobAcceptedResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type SearchAttempt struct {
	Query       string `json:"query"`
	ResultCount int    `json:"result_count"`
	Error       string `json:"error,omitempty"`
}

type JobDebugResponse struct {
	JobID            string          `json:"job_id"`
	JobType          string          `json:"job_type"`
	Status           string          `json:"status"`
	SearchAttempts   []SearchAttempt `json:"search_attempts"`
	TinyfishFetches  []string        `json:"tinyfish_fetches"`
	OpenAIFillUsed   bool            `json:"openai_fill_used"`
	OpenAIAttempts   int             `json:"openai_attempts"`
	CallbackStatus   string          `json:"callback_status"`
	CallbackAttempts int             `json:"callback_attempts"`
	LastError        string          `json:"last_error,omitempty"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type JobDonePayload struct {
	JobID        string          `json:"job_id"`
	JobType      string          `json:"job_type"`
	Status       string          `json:"status"`
	CaseID       string          `json:"case_id,omitempty"`
	UniversityID string          `json:"university_id,omitempty"`
	Error        *string         `json:"error,omitempty"`
	Result       json.RawMessage `json:"result,omitempty"`
}

type AnalyzeResult struct {
	ProfileSummary   interface{}            `json:"profile_summary"`
	Recommendations  []RecommendationResult `json:"recommendations"`
	ConfidenceScore  float64                `json:"confidence_score"`
	EscalationNeeded bool                   `json:"escalation_needed"`
	EscalationReason string                 `json:"escalation_reason"`
}

type RecommendationResult struct {
	UniversityID             string   `json:"university_id"`
	UniversityName           string   `json:"university_name"`
	Tier                     string   `json:"tier"`
	AdmissionLikelihoodScore int      `json:"admission_likelihood_score"`
	StudentFitScore          int      `json:"student_fit_score"`
	Reason                   string   `json:"reason"`
	Risks                    []string `json:"risks"`
	Improvements             []string `json:"improvements"`
	RankOrder                int      `json:"rank_order"`
}

type CrawlResult struct {
	Name                     string   `json:"name"`
	Country                  string   `json:"country"`
	QsRank                   *int     `json:"qs_rank"`
	IeltsMin                 *float64 `json:"ielts_min"`
	SatRequired              *bool    `json:"sat_required"`
	GpaExpectationNormalized *float64 `json:"gpa_expectation_normalized"`
	TuitionUsdPerYear        *int     `json:"tuition_usd_per_year"`
	ScholarshipAvailable     *bool    `json:"scholarship_available"`
	ScholarshipNotes         *string  `json:"scholarship_notes"`
	ApplicationDeadline      *string  `json:"application_deadline"`
	AvailableMajors          []string `json:"available_majors"`
	AcceptanceRate           *float64 `json:"acceptance_rate"`
	CrawlStatus              string   `json:"crawl_status"`
	ChangesDetected          []string `json:"changes_detected"`
	SourceURLs               []string `json:"source_urls"`
}

type ReportResult struct {
	PDFContent string `json:"pdf_content"`
	Summary    string `json:"summary"`
}

type ExaSearchRequest struct {
	Query      string      `json:"query"`
	NumResults int         `json:"numResults"`
	Type       string      `json:"type"`
	Contents   ExaContents `json:"contents"`
}

type ExaContents struct {
	Highlights ExaHighlights `json:"highlights"`
}

type ExaHighlights struct {
	MaxCharacters int `json:"maxCharacters"`
}

type ExaSearchResponse struct {
	Results []ExaSearchResult `json:"results"`
}

type ExaSearchResult struct {
	Title      string   `json:"title"`
	URL        string   `json:"url"`
	Highlights []string `json:"highlights"`
	Text       string   `json:"text"`
}
