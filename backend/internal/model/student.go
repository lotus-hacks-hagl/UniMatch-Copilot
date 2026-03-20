package model

import (
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type Student struct {
	Base
	FullName               string         `json:"full_name" gorm:"not null"`
	GpaRaw                 float64        `json:"gpa_raw"`
	GpaScale               float64        `json:"gpa_scale" gorm:"default:10"`
	GpaNormalized          float64        `json:"gpa_normalized"`
	IeltsOverall           *float64       `json:"ielts_overall"`
	IeltsBreakdown         datatypes.JSON `json:"ielts_breakdown" gorm:"type:jsonb"`
	SatTotal               *int           `json:"sat_total"`
	ToeflTotal             *int           `json:"toefl_total"`
	IntendedMajor          string         `json:"intended_major"`
	BudgetUsdPerYear       int            `json:"budget_usd_per_year"`
	PreferredCountries     pq.StringArray `json:"preferred_countries" gorm:"type:text[]"`
	TargetIntake           string         `json:"target_intake"`
	ScholarshipRequired    bool           `json:"scholarship_required"`
	Extracurriculars       string         `json:"extracurriculars"`
	Achievements           string         `json:"achievements"`
	PersonalStatementNotes string         `json:"personal_statement_notes"`
}

func (Student) TableName() string { return "students" }
