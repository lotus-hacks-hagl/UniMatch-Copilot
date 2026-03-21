package model

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	EventCaseCreated       = "case_created"
	EventProcessingStarted = "processing_started"
	EventAutoApproved      = "auto_approved"
	EventEscalated         = "escalated"
	EventCrawlStarted      = "crawl_started"
	EventCrawlChange       = "crawl_change"
	EventCrawlDone         = "crawl_done"
	EventReportGenerated   = "report_generated"
	EventCaseNote          = "case_note"
)

type ActivityLog struct {
	Base
	CaseID       *uuid.UUID     `json:"case_id,omitempty" gorm:"type:uuid"`
	UniversityID *uuid.UUID     `json:"university_id,omitempty" gorm:"type:uuid"`
	UserID       *uuid.UUID     `json:"user_id,omitempty" gorm:"type:uuid"`
	EventType    string         `json:"event_type"`
	Description  string         `json:"description"`
	Metadata     datatypes.JSON `json:"metadata" gorm:"type:jsonb"`
}

func (ActivityLog) TableName() string { return "activity_logs" }
