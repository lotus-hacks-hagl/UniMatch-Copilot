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
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type caseService struct {
	db       *gorm.DB
	caseRepo repository.CaseRepository
	actRepo  repository.ActivityRepository
	aiClient *client.AIClient
	cfg      *config.Config
}

func NewCaseService(
	db *gorm.DB,
	caseRepo repository.CaseRepository,
	actRepo repository.ActivityRepository,
	aiClient *client.AIClient,
	cfg *config.Config,
) CaseService {
	return &caseService{
		db:       db,
		caseRepo: caseRepo,
		actRepo:  actRepo,
		aiClient: aiClient,
		cfg:      cfg,
	}
}

// Create — creates student + case in transaction, then fires AI analyze job
func (s *caseService) Create(ctx context.Context, req dto.CreateCaseRequest) (*dto.CaseCreatedResponse, *apperror.AppError) {
	student := &model.Student{
		FullName:               req.FullName,
		GpaNormalized:          req.GpaNormalized,
		GpaRaw:                 req.GpaRaw,
		GpaScale:               req.GpaScale,
		IeltsOverall:           req.IeltsOverall,
		SatTotal:               req.SatTotal,
		ToeflTotal:             req.ToeflTotal,
		IntendedMajor:          req.IntendedMajor,
		BudgetUsdPerYear:       req.BudgetUsdPerYear,
		PreferredCountries:     req.PreferredCountries,
		TargetIntake:           req.TargetIntake,
		ScholarshipRequired:    req.ScholarshipRequired,
		Extracurriculars:       req.Extracurriculars,
		Achievements:           req.Achievements,
		PersonalStatementNotes: req.PersonalStatementNotes,
	}

	if req.IeltsBreakdown != nil {
		b, _ := json.Marshal(req.IeltsBreakdown)
		student.IeltsBreakdown = datatypes.JSON(b)
	}

	jobID := uuid.New()
	var caseRecord model.Case

	// Transaction: student + case + activity log
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(student).Error; err != nil {
			return err
		}
		caseRecord = model.Case{
			StudentID: student.ID,
			Status:    model.CaseStatusPending,
			AiJobID:   &jobID,
		}
		if err := tx.Create(&caseRecord).Error; err != nil {
			return err
		}
		return tx.Create(&model.ActivityLog{
			CaseID:      &caseRecord.ID,
			EventType:   model.EventCaseCreated,
			Description: fmt.Sprintf("Case created for %s", req.FullName),
		}).Error
	})
	if err != nil {
		return nil, apperror.Internal(err, "failed to create case")
	}

	// Fire & forget: submit analyze job to AI Service
	analyzeReq := client.AnalyzeJobRequest{
		JobID:       jobID.String(),
		CaseID:      caseRecord.ID.String(),
		CallbackURL: s.cfg.PublicBaseURL + "/internal/jobs/done",
		Input: client.AnalyzeInput{
			FullName:            student.FullName,
			GpaNormalized:       student.GpaNormalized,
			IeltsOverall:        student.IeltsOverall,
			SatTotal:            student.SatTotal,
			ToeflTotal:          student.ToeflTotal,
			IntendedMajor:       student.IntendedMajor,
			BudgetUsdPerYear:    student.BudgetUsdPerYear,
			PreferredCountries:  []string(student.PreferredCountries),
			TargetIntake:        student.TargetIntake,
			ScholarshipRequired: student.ScholarshipRequired,
			Extracurriculars:    student.Extracurriculars,
			Achievements:        student.Achievements,
		},
	}

	if err := s.aiClient.SubmitAnalyzeJob(analyzeReq); err != nil {
		s.db.Model(&caseRecord).Updates(map[string]interface{}{
			"status": model.CaseStatusFailed,
		})
		return &dto.CaseCreatedResponse{CaseID: caseRecord.ID.String(), Status: model.CaseStatusFailed}, nil
	}

	now := time.Now()
	s.db.Model(&caseRecord).Updates(map[string]interface{}{
		"status":                model.CaseStatusProcessing,
		"processing_started_at": &now,
	})
	s.db.Create(&model.ActivityLog{
		CaseID:    &caseRecord.ID,
		EventType: model.EventProcessingStarted,
	})

	return &dto.CaseCreatedResponse{CaseID: caseRecord.ID.String(), Status: model.CaseStatusProcessing}, nil
}

func (s *caseService) GetByID(ctx context.Context, id uuid.UUID) (*model.Case, *apperror.AppError) {
	c, err := s.caseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err, "failed to get case")
	}
	if c == nil {
		return nil, apperror.NotFound("case not found")
	}
	return c, nil
}

func (s *caseService) List(ctx context.Context, status string, page, limit int) ([]model.Case, int64, *apperror.AppError) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	cases, total, err := s.caseRepo.FindAll(ctx, status, page, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err, "failed to list cases")
	}
	return cases, total, nil
}

func (s *caseService) Count(ctx context.Context, status string) (int64, *apperror.AppError) {
	count, err := s.caseRepo.Count(ctx, status)
	if err != nil {
		return 0, apperror.Internal(err, "failed to count cases")
	}
	return count, nil
}

func (s *caseService) RequestReport(ctx context.Context, caseID uuid.UUID) (*dto.ReportStatusResponse, *apperror.AppError) {
	c, err := s.caseRepo.FindByID(ctx, caseID)
	if err != nil {
		return nil, apperror.Internal(err, "failed to get case")
	}
	if c == nil {
		return nil, apperror.NotFound("case not found")
	}
	if c.Status != model.CaseStatusDone && c.Status != model.CaseStatusHumanReview {
		return nil, apperror.BadRequest("case is not ready for report generation")
	}

	jobID := uuid.New()
	recs := make([]interface{}, len(c.Recommendations))
	for i, r := range c.Recommendations {
		recs[i] = r
	}

	reportReq := client.ReportJobRequest{
		JobID:           jobID.String(),
		CaseID:          caseID.String(),
		CallbackURL:     s.cfg.PublicBaseURL + "/internal/jobs/done",
		StudentName:     c.Student.FullName,
		Recommendations: recs,
	}

	if err := s.aiClient.SubmitReportJob(reportReq); err != nil {
		return nil, apperror.ServiceUnavailable("report generation service unavailable")
	}

	return &dto.ReportStatusResponse{CaseID: caseID.String(), Status: "generating"}, nil
}

// HandleJobDone — routes AI callbacks to appropriate handlers
func (s *caseService) HandleJobDone(ctx context.Context, payload dto.JobDonePayload) *apperror.AppError {
	switch payload.JobType {
	case "analyze_profile":
		return s.handleAnalyzeDone(ctx, payload)
	case "crawl_university":
		return s.handleCrawlDoneForCase(ctx, payload)
	case "generate_report":
		return s.handleReportDone(ctx, payload)
	}
	return nil
}

func (s *caseService) handleAnalyzeDone(ctx context.Context, p dto.JobDonePayload) *apperror.AppError {
	caseID, err := uuid.Parse(p.CaseID)
	if err != nil {
		return apperror.BadRequest("invalid case_id in callback")
	}
	now := time.Now()

	if p.Status == "failed" {
		s.db.WithContext(ctx).Model(&model.Case{}).Where("id = ?", caseID).Updates(map[string]interface{}{
			"status":                  model.CaseStatusHumanReview,
			"escalation_reason":       "AI service failed",
			"processing_finished_at":  &now,
		})
		s.db.Create(&model.ActivityLog{CaseID: &caseID, EventType: model.EventEscalated, Description: "AI processing failed"})
		return nil
	}

	var result dto.AnalyzeResult
	if err := json.Unmarshal(p.Result, &result); err != nil {
		return apperror.BadRequest("invalid result payload")
	}

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Bulk insert recommendations
		if len(result.Recommendations) > 0 {
			recs := make([]model.Recommendation, len(result.Recommendations))
			for i, r := range result.Recommendations {
				uniID, _ := uuid.Parse(r.UniversityID)
				risksJSON, _ := json.Marshal(r.Risks)
				improvJSON, _ := json.Marshal(r.Improvements)
				recs[i] = model.Recommendation{
					CaseID:                   caseID,
					UniversityID:             uniID,
					UniversityName:           r.UniversityName,
					Tier:                     r.Tier,
					AdmissionLikelihoodScore: r.AdmissionLikelihoodScore,
					StudentFitScore:          r.StudentFitScore,
					Reason:                   r.Reason,
					Risks:                    datatypes.JSON(risksJSON),
					Improvements:             datatypes.JSON(improvJSON),
					RankOrder:                r.RankOrder,
				}
			}
			if err := tx.Create(&recs).Error; err != nil {
				return err
			}
		}

		finalStatus := model.CaseStatusDone
		if result.EscalationNeeded {
			finalStatus = model.CaseStatusHumanReview
		}

		profileJSON, _ := json.Marshal(result.ProfileSummary)
		updates := map[string]interface{}{
			"status":                  finalStatus,
			"ai_confidence":           result.ConfidenceScore,
			"profile_summary":         datatypes.JSON(profileJSON),
			"processing_finished_at":  &now,
		}
		if result.EscalationReason != "" {
			updates["escalation_reason"] = result.EscalationReason
		}
		if err := tx.Model(&model.Case{}).Where("id = ?", caseID).Updates(updates).Error; err != nil {
			return err
		}

		eventType := model.EventAutoApproved
		if result.EscalationNeeded {
			eventType = model.EventEscalated
		}
		return tx.Create(&model.ActivityLog{CaseID: &caseID, EventType: eventType}).Error
	})

	if txErr != nil {
		return apperror.Internal(txErr, "failed to handle analyze result")
	}
	return nil
}

func (s *caseService) handleCrawlDoneForCase(ctx context.Context, p dto.JobDonePayload) *apperror.AppError {
	// Delegate — crawl results are primarily handled by UniversityService
	// This is a no-op for cases; university side handles it
	return nil
}

func (s *caseService) handleReportDone(ctx context.Context, p dto.JobDonePayload) *apperror.AppError {
	caseID, err := uuid.Parse(p.CaseID)
	if err != nil {
		return apperror.BadRequest("invalid case_id in callback")
	}
	now := time.Now()

	if p.Status == "failed" {
		return nil // silently skip, counselor can retry
	}

	var result dto.ReportResult
	if err := json.Unmarshal(p.Result, &result); err != nil {
		return apperror.BadRequest("invalid report result payload")
	}

	reportJSON, _ := json.Marshal(result)
	s.db.WithContext(ctx).Model(&model.Case{}).Where("id = ?", caseID).Updates(map[string]interface{}{
		"report_data":         datatypes.JSON(reportJSON),
		"report_generated_at": &now,
	})
	s.db.Create(&model.ActivityLog{
		CaseID:      &caseID,
		EventType:   model.EventReportGenerated,
		Description: "PDF report generated successfully",
	})
	return nil
}
