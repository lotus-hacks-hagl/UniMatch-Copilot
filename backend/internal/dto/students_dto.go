package dto

import (
	"unimatch-be/internal/model"
	"unimatch-be/pkg/response"
)

type ListStudentsResponse struct {
	Data []model.Student `json:"data"`
	Meta response.Meta   `json:"meta"`
}

type UpdateStudentRequest struct {
	FullName               string   `json:"full_name"`
	GpaRaw                 float64  `json:"gpa_raw"`
	GpaScale               float64  `json:"gpa_scale"`
	GpaNormalized          float64  `json:"gpa_normalized"`
	IeltsOverall           *float64 `json:"ielts_overall"`
	SatTotal               *int     `json:"sat_total"`
	ToeflTotal             *int     `json:"toefl_total"`
	IntendedMajor          string   `json:"intended_major"`
	BudgetUsdPerYear       int      `json:"budget_usd_per_year"`
	PreferredCountries     []string `json:"preferred_countries"`
	TargetIntake           string   `json:"target_intake"`
	ScholarshipRequired    bool     `json:"scholarship_required"`
	Extracurriculars       string   `json:"extracurriculars"`
	Achievements           string   `json:"achievements"`
	PersonalStatementNotes string   `json:"personal_statement_notes"`
	BackgroundText         string   `json:"background_text"`
}
