package service

import (
	"context"

	"unimatch-be/internal/dto"
	"unimatch-be/internal/model"
	"unimatch-be/pkg/apperror"

	"github.com/google/uuid"
)

type CaseService interface {
	Create(ctx context.Context, req dto.CreateCaseRequest) (*dto.CaseCreatedResponse, *apperror.AppError)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Case, *apperror.AppError)
	List(ctx context.Context, status string, assignedToID *uuid.UUID, filterNone bool, page, limit int) ([]model.Case, int64, *apperror.AppError)
	Claim(ctx context.Context, id uuid.UUID, userID uuid.UUID) *apperror.AppError
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) *apperror.AppError
	Count(ctx context.Context, status string) (int64, *apperror.AppError)
	RequestReport(ctx context.Context, caseID uuid.UUID) (*dto.ReportStatusResponse, *apperror.AppError)
	HandleJobDone(ctx context.Context, payload dto.JobDonePayload) *apperror.AppError
}

type UniversityService interface {
	Create(ctx context.Context, req dto.CreateUniversityRequest) (*model.University, *apperror.AppError)
	List(ctx context.Context, country, search string, page, limit int) ([]model.University, int64, *apperror.AppError)
	CrawlAll(ctx context.Context) (int, *apperror.AppError)
	CountActiveCrawls(ctx context.Context) (int64, *apperror.AppError)
	HandleCrawlDone(ctx context.Context, payload dto.JobDonePayload) *apperror.AppError
}

type DashboardService interface {
	GetStats(ctx context.Context) (*dto.DashboardStats, *apperror.AppError)
	GetCasesByDay(ctx context.Context) ([]dto.CasesByDay, *apperror.AppError)
	GetEscalationTrend(ctx context.Context) ([]dto.EscalationTrend, *apperror.AppError)
	GetAnalytics(ctx context.Context) (*dto.Analytics, *apperror.AppError)
	GetActivityLog(ctx context.Context, limit int) ([]model.ActivityLog, *apperror.AppError)
}

type StudentService interface {
	List(ctx context.Context, page, limit int) (*dto.ListStudentsResponse, *apperror.AppError)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Student, *apperror.AppError)
}
