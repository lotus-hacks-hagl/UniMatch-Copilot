package service

import (
	"context"
	"io"

	"unimatch-be/internal/dto"
	"unimatch-be/internal/model"
	"unimatch-be/pkg/apperror"

	"github.com/google/uuid"
)

type CaseService interface {
	Create(ctx context.Context, req dto.CreateCaseRequest) (*dto.CaseCreatedResponse, *apperror.AppError)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Case, *apperror.AppError)
	List(ctx context.Context, status string, assignedToID *uuid.UUID, filterNone bool, search string, page, limit int) ([]model.Case, int64, *apperror.AppError)
	Claim(ctx context.Context, id uuid.UUID, userID uuid.UUID) *apperror.AppError
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) *apperror.AppError
	Count(ctx context.Context, status string) (int64, *apperror.AppError)
	RequestReport(ctx context.Context, caseID uuid.UUID) (*dto.ReportStatusResponse, *apperror.AppError)
	HandleJobDone(ctx context.Context, payload dto.JobDonePayload) *apperror.AppError
	AddNote(ctx context.Context, id uuid.UUID, userID *uuid.UUID, text string) *apperror.AppError
	ReAnalyze(ctx context.Context, id uuid.UUID) *apperror.AppError
}

type UniversityService interface {
	Create(ctx context.Context, req dto.CreateUniversityRequest) (*model.University, *apperror.AppError)
	List(ctx context.Context, country, search string, page, limit int) ([]model.University, int64, *apperror.AppError)
	Crawl(ctx context.Context, id uuid.UUID) *apperror.AppError
	CrawlAll(ctx context.Context) (int, *apperror.AppError)
	CountActiveCrawls(ctx context.Context) (int64, *apperror.AppError)
	HandleCrawlDone(ctx context.Context, payload dto.JobDonePayload) *apperror.AppError
	Delete(ctx context.Context, id uuid.UUID) *apperror.AppError
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
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateStudentRequest) *apperror.AppError
	Delete(ctx context.Context, id uuid.UUID) *apperror.AppError
}

type CaseDocumentService interface {
	Upload(ctx context.Context, caseID uuid.UUID, userID *uuid.UUID, fileName string, fileType string, fileSize int64, reader io.Reader) (*model.CaseDocument, *apperror.AppError)
	List(ctx context.Context, caseID uuid.UUID) ([]model.CaseDocument, *apperror.AppError)
	GetByID(ctx context.Context, id uuid.UUID) (*model.CaseDocument, *apperror.AppError)
	GetPhysicalPath(doc *model.CaseDocument) string
	Delete(ctx context.Context, id uuid.UUID) *apperror.AppError
}

type StudentRepository interface {
	FindAll(ctx context.Context, page, limit int) ([]model.Student, int64, error)
	Count(ctx context.Context) (int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Student, error)
	Update(ctx context.Context, id uuid.UUID, s *model.Student) error
	Delete(ctx context.Context, id uuid.UUID) error
}
