package dto

import (
	"encoding/json"
	"errors"
)

// ─── Requests ────────────────────────────────────────────────────────────────

type CreateCaseRequest struct {
	FullName               string   `json:"full_name" validate:"required"`
	GpaNormalized          *float64 `json:"gpa_normalized" validate:"required,min=0,max=4"`
	GpaRaw                 float64  `json:"gpa_raw"`
	GpaScale               float64  `json:"gpa_scale"`
	IeltsOverall           *float64 `json:"ielts_overall"`
	IeltsBreakdown         *IeltsBreakdown `json:"ielts_breakdown"`
	SatTotal               *int     `json:"sat_total"`
	ToeflTotal             *int     `json:"toefl_total"`
	IntendedMajor          string   `json:"intended_major"`
	BudgetUsdPerYear       *int     `json:"budget_usd_per_year" validate:"min=0"`
	PreferredCountries     []string `json:"preferred_countries"`
	TargetIntake           string   `json:"target_intake"`
	ScholarshipRequired    bool     `json:"scholarship_required"`
	Extracurriculars       string   `json:"extracurriculars"`
	Achievements           string   `json:"achievements"`
	PersonalStatementNotes string   `json:"personal_statement_notes"`
}

type IeltsBreakdown struct {
	Listening float64 `json:"listening"`
	Reading   float64 `json:"reading"`
	Writing   float64 `json:"writing"`
	Speaking  float64 `json:"speaking"`
}

// Validate ensures at least one of ielts_overall or sat_total is provided
func (r *CreateCaseRequest) Validate() error {
	if r.IeltsOverall == nil && r.SatTotal == nil {
		return errors.New("ielts_overall or sat_total is required")
	}
	return nil
}

// ─── Responses ───────────────────────────────────────────────────────────────

type CaseCreatedResponse struct {
	CaseID string `json:"case_id"`
	Status string `json:"status"`
}

type ReportStatusResponse struct {
	CaseID    string `json:"case_id"`
	Status    string `json:"status"`
	ReportURL string `json:"report_url,omitempty"`
}

// ─── AI Callback Payloads ────────────────────────────────────────────────────

type JobDonePayload struct {
	JobID        string          `json:"job_id"`
	JobType      string          `json:"job_type"`
	Status       string          `json:"status"`
	CaseID       string          `json:"case_id"`
	UniversityID string          `json:"university_id"`
	Error        *string         `json:"error"`
	Result       json.RawMessage `json:"result"`
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
	PDFContent string `json:"pdf_content"` // base64 encoded or URL
	Summary    string `json:"summary"`
}

type CrawlStatusItem struct {
	UniversityID   string `json:"university_id"`
	UniversityName string `json:"university_name"`
	Status         string `json:"status"`
}
