package swagger

import (
	"unimatch-be/internal/dto"
	"unimatch-be/pkg/response"
)

// SwaggerError is a wrapper for documentation
type SwaggerError struct {
	Code    string `json:"code" example:"NOT_FOUND"`
	Message string `json:"message" example:"Resource not found"`
	Details string `json:"details,omitempty" example:"User with ID 123 not found"`
}

// Ensure docs use full shapes for examples

// Request examples
type _ dto.CreateCaseRequest
type _ dto.CreateUniversityRequest

// Response examples
type _ response.Response
type _ response.Meta
type _ dto.CaseCreatedResponse
type _ dto.ReportStatusResponse
type _ dto.DashboardStats
type _ dto.CasesByDay
type _ dto.EscalationTrend
type _ dto.Analytics
type _ dto.CrawlAllResponse
