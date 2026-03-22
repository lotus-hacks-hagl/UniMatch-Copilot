package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"unimatch-be/config"
	"unimatch-be/internal/dto"
	"unimatch-be/internal/model"
	"unimatch-be/internal/repository"
	"unimatch-be/pkg/apperror"
	"unimatch-be/pkg/client"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type universityService struct {
	db       *gorm.DB
	uniRepo  repository.UniversityRepository
	actRepo  repository.ActivityRepository
	aiClient *client.AIClient
	cfg      *config.Config
}

func NewUniversityService(
	db *gorm.DB,
	uniRepo repository.UniversityRepository,
	actRepo repository.ActivityRepository,
	aiClient *client.AIClient,
	cfg *config.Config,
) UniversityService {
	return &universityService{
		db:       db,
		uniRepo:  uniRepo,
		actRepo:  actRepo,
		aiClient: aiClient,
		cfg:      cfg,
	}
}

func (s *universityService) Create(ctx context.Context, req dto.CreateUniversityRequest) (*model.University, *apperror.AppError) {
	u := &model.University{
		Name:                     req.Name,
		Country:                  req.Country,
		QsRank:                   req.QsRank,
		GroupTag:                 req.GroupTag,
		IeltsMin:                 req.IeltsMin,
		SatRequired:              req.SatRequired,
		GpaExpectationNormalized: req.GpaExpectationNormalized,
		TuitionUsdPerYear:        req.TuitionUsdPerYear,
		ScholarshipAvailable:     req.ScholarshipAvailable,
		ScholarshipNotes:         req.ScholarshipNotes,
		AvailableMajors:          pq.StringArray(req.AvailableMajors),
		AcceptanceRate:           req.AcceptanceRate,
		CounselorNotes:           req.CounselorNotes,
		CrawlStatus:              model.CrawlStatusNeverCrawled,
	}

	if err := s.uniRepo.Create(ctx, u); err != nil {
		return nil, apperror.Internal(err, "failed to create university")
	}
	return u, nil
}

func (s *universityService) List(ctx context.Context, country, search string, page, limit int) ([]model.University, int64, *apperror.AppError) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	unis, total, err := s.uniRepo.FindAll(ctx, country, search, page, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err, "failed to list universities")
	}
	return unis, total, nil
}

func (s *universityService) Crawl(ctx context.Context, id uuid.UUID) *apperror.AppError {
	u, err := s.uniRepo.FindByID(ctx, id)
	if err != nil {
		return apperror.NotFound("university not found")
	}

	jid, appErr := s.submitCrawlJob(ctx, u)
	if appErr != nil {
		return appErr
	}

	s.uniRepo.UpdateCrawlResult(ctx, u.ID, map[string]interface{}{
		"crawl_status": model.CrawlStatusPending,
		"crawl_job_id": &jid,
	})
	s.actRepo.Create(ctx, &model.ActivityLog{
		UniversityID: &u.ID,
		EventType:    model.EventCrawlStarted,
		Description:  fmt.Sprintf("Crawl job submitted for %s", u.Name),
	})

	return nil
}

func (s *universityService) CrawlAll(ctx context.Context) (int, *apperror.AppError) {
	unis, err := s.uniRepo.FindCrawlable(ctx)
	if err != nil {
		return 0, apperror.Internal(err, "failed to find universities")
	}

	fmt.Printf("[UniversityService] Found %d crawlable universities\n", len(unis))
	triggered := 0
	for _, u := range unis {
		jid, appErr := s.submitCrawlJob(ctx, &u)
		if appErr != nil {
			fmt.Printf("[UniversityService] Failed to trigger sync for %s: %v\n", u.Name, appErr.Message)
			continue // skip failed, log and move on
		}

		s.uniRepo.UpdateCrawlResult(ctx, u.ID, map[string]interface{}{
			"crawl_status": model.CrawlStatusPending,
			"crawl_job_id": &jid,
		})
		s.actRepo.Create(ctx, &model.ActivityLog{
			UniversityID: &u.ID,
			EventType:    model.EventCrawlStarted,
			Description:  fmt.Sprintf("Crawl job submitted for %s", u.Name),
		})
		triggered++
	}

	return triggered, nil
}

func (s *universityService) submitCrawlJob(ctx context.Context, u *model.University) (uuid.UUID, *apperror.AppError) {
	jobID := uuid.New()
	metadata := map[string]interface{}{
		"name":                       u.Name,
		"country":                    u.Country,
		"qs_rank":                    u.QsRank,
		"ielts_min":                  u.IeltsMin,
		"sat_required":               u.SatRequired,
		"gpa_expectation_normalized": u.GpaExpectationNormalized,
		"tuition_usd_per_year":       u.TuitionUsdPerYear,
		"scholarship_available":      u.ScholarshipAvailable,
		"scholarship_notes":          u.ScholarshipNotes,
		"available_majors":           u.AvailableMajors,
		"acceptance_rate":            u.AcceptanceRate,
	}

	// Optimization: skip crawling if all key fields are already manually entered
	isMetadataOnly := u.QsRank != nil && u.IeltsMin != nil && u.GpaExpectationNormalized != nil && u.TuitionUsdPerYear != nil

	crawlReq := client.CrawlJobRequest{
		JobID:          jobID.String(),
		UniversityID:   u.ID.String(),
		CallbackURL:    s.cfg.InternalBaseURL + "/internal/jobs/done",
		IsMetadataOnly: isMetadataOnly,
		Metadata:       metadata,
	}

	if err := s.aiClient.SubmitCrawlJob(crawlReq); err != nil {
		return uuid.Nil, apperror.ServiceUnavailable("AI service unavailable")
	}

	return jobID, nil
}

func (s *universityService) CountActiveCrawls(ctx context.Context) (int64, *apperror.AppError) {
	count, err := s.uniRepo.CountByCrawlStatus(ctx, model.CrawlStatusPending)
	if err != nil {
		return 0, apperror.Internal(err, "failed to count active crawls")
	}
	return count, nil
}

// HandleCrawlDone processes crawl callback from AI Service
func (s *universityService) HandleCrawlDone(ctx context.Context, p dto.JobDonePayload) *apperror.AppError {
	uniID, err := uuid.Parse(p.UniversityID)
	if err != nil {
		return apperror.BadRequest("invalid university_id in callback")
	}

	if p.Status == "changed" {
		p.Status = "ok"
	}

	if p.Status == "failed" {
		s.uniRepo.UpdateCrawlResult(ctx, uniID, map[string]interface{}{
			"crawl_status": model.CrawlStatusFailed,
			"crawl_job_id": nil,
		})
		return nil
	}

	var result dto.CrawlResult
	if err := json.Unmarshal(p.Result, &result); err != nil {
		return apperror.BadRequest("invalid crawl result payload")
	}

	// Build partial update map — only update non-nil fields
	updates := map[string]interface{}{
		"crawl_status": result.CrawlStatus,
		"crawl_job_id": nil,
	}
	now := time.Now()
	updates["last_crawled_at"] = &now

	if result.QsRank != nil {
		updates["qs_rank"] = result.QsRank
	}
	if result.IeltsMin != nil {
		updates["ielts_min"] = result.IeltsMin
	}
	if result.SatRequired != nil {
		updates["sat_required"] = result.SatRequired
	}
	if result.GpaExpectationNormalized != nil {
		updates["gpa_expectation_normalized"] = result.GpaExpectationNormalized
	}
	if result.TuitionUsdPerYear != nil {
		updates["tuition_usd_per_year"] = result.TuitionUsdPerYear
	}
	if result.ScholarshipAvailable != nil {
		updates["scholarship_available"] = result.ScholarshipAvailable
	}
	if result.AcceptanceRate != nil {
		updates["acceptance_rate"] = result.AcceptanceRate
	}
	if len(result.AvailableMajors) > 0 {
		updates["available_majors"] = pq.StringArray(result.AvailableMajors)
	}
	if result.ScholarshipNotes != nil {
		updates["scholarship_notes"] = *result.ScholarshipNotes
	}
	if result.ApplicationDeadline != nil {
		if t, err := time.Parse("2006-01-02", *result.ApplicationDeadline); err == nil {
			updates["application_deadline"] = &t
		}
	}

	s.uniRepo.UpdateCrawlResult(ctx, uniID, updates)

	if len(result.ChangesDetected) > 0 {
		metaJSON, _ := json.Marshal(map[string]interface{}{
			"changes":     result.ChangesDetected,
			"source_urls": result.SourceURLs,
		})
		s.actRepo.Create(ctx, &model.ActivityLog{
			UniversityID: &uniID,
			EventType:    model.EventCrawlChange,
			Description:  fmt.Sprintf("%d changes detected", len(result.ChangesDetected)),
			Metadata:     datatypes.JSON(metaJSON),
		})
	}
	return nil
}

func (s *universityService) Delete(ctx context.Context, id uuid.UUID) *apperror.AppError {
	// First check if it exists
	uni, err := s.uniRepo.FindByID(ctx, id)
	if err != nil {
		return apperror.NotFound("university not found")
	}

	// Delete from Neo4j (Sync Graph)
	if err := s.aiClient.DeleteUniversity(id.String()); err != nil {
		fmt.Printf("Warning: failed to delete university from AI service graph: %v\n", err)
		// We proceed anyway to keep the SQL DB in sync, or we could fail here.
		// For 24h hackathon, let's just log and continue.
	}

	// Delete from PostgreSQL
	if err := s.uniRepo.Delete(ctx, id); err != nil {
		return apperror.Internal(err, "failed to delete university from database")
	}

	// Log activity
	s.actRepo.Create(ctx, &model.ActivityLog{
		EventType:   model.EventCrawlDone, // Or define a new event type if needed
		Description: fmt.Sprintf("University %s deleted and graph synced.", uni.Name),
	})

	return nil
}
