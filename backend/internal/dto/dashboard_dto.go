package dto

import "time"

// ─── Dashboard Stats ──────────────────────────────────────────────────────────

type DashboardStats struct {
	CasesToday           int64   `json:"casesToday"`
	AwaitingReview       int64   `json:"awaitingReview"`
	AvgProcessingMinutes float64 `json:"avgProcessingTime"`
	AiConfidenceAvg      float64 `json:"aiConfidenceAvg"`
	ActiveCrawls         int64   `json:"activeCrawls"`
}

type CasesByDay struct {
	Date  time.Time `json:"date"`
	Count int64     `json:"count"`
}

type EscalationTrend struct {
	Date  time.Time `json:"date"`
	Count int64     `json:"count"`
}

type Analytics struct {
	AutoApprovalRate    float64            `json:"auto_approval_rate"`
	TopUniversities     []UniversityCount  `json:"top_universities"`
	CountryDistribution []CountryCount     `json:"country_distribution"`
}

type UniversityCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type CountryCount struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}
