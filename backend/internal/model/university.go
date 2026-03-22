package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const (
	CrawlStatusOK           = "ok"
	CrawlStatusPending      = "pending"
	CrawlStatusChanged      = "changed"
	CrawlStatusFailed       = "failed"
	CrawlStatusNeverCrawled = "never_crawled"
)

type University struct {
	Base
	Name                     string         `json:"name" gorm:"not null"`
	Country                  string         `json:"country"`
	QsRank                   *int           `json:"qs_rank"`
	GroupTag                 string         `json:"group_tag"`
	IeltsMin                 *float64       `json:"ielts_min"`
	SatRequired              bool           `json:"sat_required"`
	GpaExpectationNormalized *float64       `json:"gpa_expectation_normalized"`
	TuitionUsdPerYear        *int           `json:"tuition_usd_per_year"`
	ScholarshipAvailable     bool           `json:"scholarship_available"`
	ScholarshipNotes         string         `json:"scholarship_notes"`
	AvailableMajors          pq.StringArray `json:"available_majors" gorm:"type:text[]"`
	ApplicationDeadline      *time.Time     `json:"application_deadline"`
	AcceptanceRate           *float64       `json:"acceptance_rate"`
	CrawlStatus              string         `json:"crawl_status" gorm:"default:'never_crawled'"`
	LastCrawledAt            *time.Time     `json:"last_crawled_at"`
	CrawlJobID               *uuid.UUID     `json:"crawl_job_id" gorm:"type:uuid"`
	CounselorNotes           string         `json:"counselor_notes"`
}

func (University) TableName() string { return "universities" }
