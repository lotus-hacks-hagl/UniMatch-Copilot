package service

import (
	"context"

	"ai-service-go/internal/dto"
	"ai-service-go/pkg/apperror"
)

type JobService interface {
	EnqueueAnalyze(ctx context.Context, req dto.AnalyzeJobRequest) *apperror.AppError
	EnqueueCrawl(ctx context.Context, req dto.CrawlJobRequest) *apperror.AppError
	EnqueueReport(ctx context.Context, req dto.ReportJobRequest) *apperror.AppError
	GetJob(ctx context.Context, jobID string) (*dto.JobDebugResponse, *apperror.AppError)
}
