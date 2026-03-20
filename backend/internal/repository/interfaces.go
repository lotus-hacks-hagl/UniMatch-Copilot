package repository

import (
	"context"

	"unimatch-be/internal/dto"
	"unimatch-be/internal/model"

	"github.com/google/uuid"
)

type CaseRepository interface {
	Create(ctx context.Context, c *model.Case) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Case, error)
	FindAll(ctx context.Context, status string, page, limit int) ([]model.Case, int64, error)
	Update(ctx context.Context, c *model.Case) error
	UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error
	Count(ctx context.Context, status string) (int64, error)
}

type UniversityRepository interface {
	Create(ctx context.Context, u *model.University) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.University, error)
	FindAll(ctx context.Context, country, search string, page, limit int) ([]model.University, int64, error)
	FindCrawlable(ctx context.Context) ([]model.University, error)
	CountByCrawlStatus(ctx context.Context, status string) (int64, error)
	UpdateCrawlResult(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error
}

type ActivityRepository interface {
	Create(ctx context.Context, log *model.ActivityLog) error
	FindRecent(ctx context.Context, limit int) ([]model.ActivityLog, error)
}

type DashboardRepository interface {
	GetStats(ctx context.Context) (*dto.DashboardStats, error)
	GetCasesByDay(ctx context.Context) ([]dto.CasesByDay, error)
	GetEscalationTrend(ctx context.Context) ([]dto.EscalationTrend, error)
	GetAnalytics(ctx context.Context) (*dto.Analytics, error)
}
