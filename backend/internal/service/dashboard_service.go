package service

import (
	"context"

	"unimatch-be/internal/dto"
	"unimatch-be/internal/model"
	"unimatch-be/internal/repository"
	"unimatch-be/pkg/apperror"
)

type dashboardService struct {
	dashRepo repository.DashboardRepository
	actRepo  repository.ActivityRepository
}

func NewDashboardService(
	dashRepo repository.DashboardRepository,
	actRepo repository.ActivityRepository,
) DashboardService {
	return &dashboardService{
		dashRepo: dashRepo,
		actRepo:  actRepo,
	}
}

func (s *dashboardService) GetStats(ctx context.Context) (*dto.DashboardStats, *apperror.AppError) {
	stats, err := s.dashRepo.GetStats(ctx)
	if err != nil {
		return nil, apperror.Internal(err, "failed to get stats")
	}
	return stats, nil
}

func (s *dashboardService) GetCasesByDay(ctx context.Context) ([]dto.CasesByDay, *apperror.AppError) {
	data, err := s.dashRepo.GetCasesByDay(ctx)
	if err != nil {
		return nil, apperror.Internal(err, "failed to get cases by day")
	}
	return data, nil
}

func (s *dashboardService) GetEscalationTrend(ctx context.Context) ([]dto.EscalationTrend, *apperror.AppError) {
	data, err := s.dashRepo.GetEscalationTrend(ctx)
	if err != nil {
		return nil, apperror.Internal(err, "failed to get escalation trend")
	}
	return data, nil
}

func (s *dashboardService) GetAnalytics(ctx context.Context) (*dto.Analytics, *apperror.AppError) {
	analytics, err := s.dashRepo.GetAnalytics(ctx)
	if err != nil {
		return nil, apperror.Internal(err, "failed to get analytics")
	}
	return analytics, nil
}

func (s *dashboardService) GetActivityLog(ctx context.Context, limit int) ([]model.ActivityLog, *apperror.AppError) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	logs, err := s.actRepo.FindRecent(ctx, limit)
	if err != nil {
		return nil, apperror.Internal(err, "failed to get activity log")
	}
	return logs, nil
}
