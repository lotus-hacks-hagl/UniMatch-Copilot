package dto

type CreateUniversityRequest struct {
	Name                     string   `json:"name" validate:"required"`
	Country                  string   `json:"country"`
	QsRank                   *int     `json:"qs_rank"`
	GroupTag                 string   `json:"group_tag"`
	IeltsMin                 *float64 `json:"ielts_min"`
	SatRequired              bool     `json:"sat_required"`
	GpaExpectationNormalized *float64 `json:"gpa_expectation_normalized"`
	TuitionUsdPerYear        *int     `json:"tuition_usd_per_year"`
	ScholarshipAvailable     bool     `json:"scholarship_available"`
	ScholarshipNotes         string   `json:"scholarship_notes"`
	AvailableMajors          []string `json:"available_majors"`
	AcceptanceRate           *float64 `json:"acceptance_rate"`
	CounselorNotes           string   `json:"counselor_notes"`
}

type CrawlAllResponse struct {
	Triggered int    `json:"triggered"`
	Message   string `json:"message"`
}

type UpdateUniversityRequest struct {
	Name                     string   `json:"name" validate:"required"`
	Country                  string   `json:"country"`
	QsRank                   *int     `json:"qs_rank"`
	GroupTag                 string   `json:"group_tag"`
	IeltsMin                 *float64 `json:"ielts_min"`
	SatRequired              bool     `json:"sat_required"`
	GpaExpectationNormalized *float64 `json:"gpa_expectation_normalized"`
	TuitionUsdPerYear        *int     `json:"tuition_usd_per_year"`
	ScholarshipAvailable     bool     `json:"scholarship_available"`
	ScholarshipNotes         string   `json:"scholarship_notes"`
	AvailableMajors          []string `json:"available_majors"`
	AcceptanceRate           *float64 `json:"acceptance_rate"`
	CounselorNotes           string   `json:"counselor_notes"`
}
