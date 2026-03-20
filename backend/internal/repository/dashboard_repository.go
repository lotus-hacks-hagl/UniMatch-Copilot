package repository

import (
	"context"

	"unimatch-be/internal/dto"
	"unimatch-be/internal/model"

	"gorm.io/gorm"
)

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) GetStats(ctx context.Context) (*dto.DashboardStats, error) {
	var stats dto.DashboardStats

	r.db.WithContext(ctx).Model(&model.Case{}).
		Where("DATE(created_at) = CURRENT_DATE").Count(&stats.CasesToday)

	r.db.WithContext(ctx).Model(&model.Case{}).
		Where("status = ?", model.CaseStatusHumanReview).Count(&stats.AwaitingReview)

	r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (processing_finished_at - processing_started_at))/60), 0)
		FROM cases
		WHERE processing_finished_at IS NOT NULL
		  AND created_at > NOW() - INTERVAL '7 days'
	`).Scan(&stats.AvgProcessingMinutes)

	r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(AVG(ai_confidence) * 100, 0)
		FROM cases
		WHERE created_at > NOW() - INTERVAL '7 days'
		  AND ai_confidence IS NOT NULL
	`).Scan(&stats.AiConfidenceAvg)

	r.db.WithContext(ctx).Model(&model.University{}).
		Where("crawl_status = ?", model.CrawlStatusPending).Count(&stats.ActiveCrawls)

	return &stats, nil
}

func (r *dashboardRepository) GetCasesByDay(ctx context.Context) ([]dto.CasesByDay, error) {
	var results []dto.CasesByDay
	err := r.db.WithContext(ctx).Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM cases
		WHERE created_at > NOW() - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`).Scan(&results).Error
	return results, err
}

func (r *dashboardRepository) GetEscalationTrend(ctx context.Context) ([]dto.EscalationTrend, error) {
	var results []dto.EscalationTrend
	err := r.db.WithContext(ctx).Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM cases
		WHERE status = 'human_review'
		  AND created_at > NOW() - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`).Scan(&results).Error
	return results, err
}

func (r *dashboardRepository) GetAnalytics(ctx context.Context) (*dto.Analytics, error) {
	var analytics dto.Analytics

	// Auto approval rate
	var total, done int64
	r.db.WithContext(ctx).Model(&model.Case{}).Count(&total)
	r.db.WithContext(ctx).Model(&model.Case{}).Where("status = ?", model.CaseStatusDone).Count(&done)
	if total > 0 {
		analytics.AutoApprovalRate = float64(done) / float64(total) * 100
	}

	// Top universities from recommendations
	err := r.db.WithContext(ctx).Raw(`
		SELECT university_name as name, COUNT(*) as count
		FROM recommendations
		GROUP BY university_name
		ORDER BY count DESC
		LIMIT 10
	`).Scan(&analytics.TopUniversities).Error
	if err != nil {
		return nil, err
	}

	// Country distribution from students
	err = r.db.WithContext(ctx).Raw(`
		SELECT unnest(preferred_countries) as country, COUNT(*) as count
		FROM students
		GROUP BY country
		ORDER BY count DESC
		LIMIT 10
	`).Scan(&analytics.CountryDistribution).Error

	return &analytics, err
}
