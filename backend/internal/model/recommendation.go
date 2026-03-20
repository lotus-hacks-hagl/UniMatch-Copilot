package model

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Recommendation struct {
	Base
	CaseID                  uuid.UUID      `json:"case_id" gorm:"type:uuid;not null"`
	UniversityID            uuid.UUID      `json:"university_id" gorm:"type:uuid"`
	UniversityName          string         `json:"university_name"`
	Tier                    string         `json:"tier"` // safe | match | reach
	AdmissionLikelihoodScore int           `json:"admission_likelihood_score"`
	StudentFitScore         int            `json:"student_fit_score"`
	Reason                  string         `json:"reason"`
	Risks                   datatypes.JSON `json:"risks" gorm:"type:jsonb"`
	Improvements            datatypes.JSON `json:"improvements" gorm:"type:jsonb"`
	RankOrder               int            `json:"rank_order"`
}

func (Recommendation) TableName() string { return "recommendations" }
