package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	CaseStatusPending     = "pending"
	CaseStatusProcessing  = "processing"
	CaseStatusDone        = "done"
	CaseStatusHumanReview = "human_review"
	CaseStatusFailed      = "failed"
)

type Case struct {
	Base
	StudentID            uuid.UUID        `json:"student_id" gorm:"type:uuid;not null"`
	Student              *Student         `json:"student,omitempty" gorm:"foreignKey:StudentID"`
	AssignedToID         *uuid.UUID       `json:"assigned_to_id" gorm:"type:uuid"`
	AssignedTo           *User            `json:"assigned_to,omitempty" gorm:"foreignKey:AssignedToID"`
	Status               string           `json:"status" gorm:"default:'pending'"`
	AiJobID              *uuid.UUID       `json:"ai_job_id" gorm:"type:uuid"`
	AiConfidence         *float64         `json:"ai_confidence"`
	EscalationReason     *string          `json:"escalation_reason"`
	ProfileSummary       datatypes.JSON   `json:"profile_summary" gorm:"type:jsonb"`
	ReportData           datatypes.JSON   `json:"report_data,omitempty" gorm:"type:jsonb"`
	ReportGeneratedAt    *time.Time       `json:"report_generated_at"`
	ProcessingStartedAt  *time.Time       `json:"processing_started_at"`
	ProcessingFinishedAt *time.Time       `json:"processing_finished_at"`
	Recommendations      []Recommendation `json:"recommendations,omitempty" gorm:"foreignKey:CaseID"`
}

func (Case) TableName() string { return "cases" }
