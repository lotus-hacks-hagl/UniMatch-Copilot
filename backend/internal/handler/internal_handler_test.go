package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"unimatch-be/internal/dto"
	"unimatch-be/internal/model"
	"unimatch-be/pkg/apperror"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubCaseService struct {
	handleJobDone func(ctx context.Context, payload dto.JobDonePayload) *apperror.AppError
}

func (s *stubCaseService) Create(ctx context.Context, req dto.CreateCaseRequest) (*dto.CaseCreatedResponse, *apperror.AppError) {
	return nil, nil
}

func (s *stubCaseService) GetByID(ctx context.Context, id uuid.UUID) (*model.Case, *apperror.AppError) {
	return nil, nil
}

func (s *stubCaseService) List(ctx context.Context, status string, assignedToID *uuid.UUID, filterNone bool, page, limit int) ([]model.Case, int64, *apperror.AppError) {
	return nil, 0, nil
}

func (s *stubCaseService) Claim(ctx context.Context, id uuid.UUID, userID uuid.UUID) *apperror.AppError {
	return nil
}

func (s *stubCaseService) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) *apperror.AppError {
	return nil
}

func (s *stubCaseService) Count(ctx context.Context, status string) (int64, *apperror.AppError) {
	return 0, nil
}

func (s *stubCaseService) RequestReport(ctx context.Context, caseID uuid.UUID) (*dto.ReportStatusResponse, *apperror.AppError) {
	return nil, nil
}

func (s *stubCaseService) HandleJobDone(ctx context.Context, payload dto.JobDonePayload) *apperror.AppError {
	if s.handleJobDone != nil {
		return s.handleJobDone(ctx, payload)
	}
	return nil
}

func (s *stubCaseService) AddNote(ctx context.Context, id uuid.UUID, userID *uuid.UUID, text string) *apperror.AppError {
	return nil
}

func (s *stubCaseService) ReAnalyze(ctx context.Context, id uuid.UUID) *apperror.AppError {
	return nil
}

type stubUniversityService struct {
	handleCrawlDone func(ctx context.Context, payload dto.JobDonePayload) *apperror.AppError
}

func (s *stubUniversityService) Create(ctx context.Context, req dto.CreateUniversityRequest) (*model.University, *apperror.AppError) {
	return nil, nil
}

func (s *stubUniversityService) List(ctx context.Context, country, search string, page, limit int) ([]model.University, int64, *apperror.AppError) {
	return nil, 0, nil
}

func (s *stubUniversityService) CrawlAll(ctx context.Context) (int, *apperror.AppError) {
	return 0, nil
}

func (s *stubUniversityService) CountActiveCrawls(ctx context.Context) (int64, *apperror.AppError) {
	return 0, nil
}

func (s *stubUniversityService) HandleCrawlDone(ctx context.Context, payload dto.JobDonePayload) *apperror.AppError {
	if s.handleCrawlDone != nil {
		return s.handleCrawlDone(ctx, payload)
	}
	return nil
}

func TestInternalHandlerJobDoneReturnsOKOnNilAppError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	handler := NewInternalHandler(
		&stubCaseService{
			handleJobDone: func(ctx context.Context, payload dto.JobDonePayload) *apperror.AppError {
				return nil
			},
		},
		&stubUniversityService{},
	)

	body, err := json.Marshal(dto.JobDonePayload{
		JobID:   "job-1",
		JobType: "analyze_profile",
		Status:  "done",
		CaseID:  uuid.NewString(),
		Result:  json.RawMessage(`{"recommendations":[]}`),
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/internal/jobs/done", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	handler.JobDone(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", rec.Code, rec.Body.String())
	}
}
