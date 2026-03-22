package repository

import (
	"context"
	"fmt"

	"unimatch-be/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type universityRepository struct {
	db *gorm.DB
}

func NewUniversityRepository(db *gorm.DB) UniversityRepository {
	return &universityRepository{db: db}
}

func (r *universityRepository) Create(ctx context.Context, u *model.University) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *universityRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.University, error) {
	var u model.University
	err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *universityRepository) FindAll(ctx context.Context, country, search string, page, limit int) ([]model.University, int64, error) {
	var unis []model.University
	var total int64

	q := r.db.WithContext(ctx).Model(&model.University{})
	if country != "" {
		q = q.Where("country = ?", country)
	}
	if search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}

	q.Count(&total)
	err := q.Order("qs_rank ASC NULLS LAST").
		Offset((page - 1) * limit).Limit(limit).
		Find(&unis).Error
	return unis, total, err
}

func (r *universityRepository) FindAnalyzeCandidates(ctx context.Context, countries []string, major string, budget int, limit int) ([]model.University, error) {
	var unis []model.University

	q := r.db.WithContext(ctx).Model(&model.University{})
	if len(countries) > 0 {
		q = q.Where("country IN ?", countries)
	}
	if major != "" {
		q = q.Where("array_to_string(available_majors, ',') ILIKE ?", "%"+major+"%")
	}
	if budget > 0 {
		q = q.Order("CASE WHEN tuition_usd_per_year IS NULL THEN 1 WHEN tuition_usd_per_year <= " +
			fmt.Sprint(budget) + " THEN 0 ELSE 1 END ASC")
	}

	err := q.
		Order("qs_rank ASC NULLS LAST").
		Limit(limit).
		Find(&unis).Error
	if err != nil {
		return nil, err
	}

	if len(unis) > 0 || (len(countries) == 0 && major == "") {
		return unis, nil
	}

	err = r.db.WithContext(ctx).
		Model(&model.University{}).
		Order("qs_rank ASC NULLS LAST").
		Limit(limit).
		Find(&unis).Error
	return unis, err
}

func (r *universityRepository) FindCrawlable(ctx context.Context) ([]model.University, error) {
	var unis []model.University
	err := r.db.WithContext(ctx).
		Where("crawl_status != ? AND (last_crawled_at IS NULL OR last_crawled_at < NOW() - INTERVAL '1 day')",
			model.CrawlStatusPending).
		Find(&unis).Error
	return unis, err
}

func (r *universityRepository) CountByCrawlStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.University{}).
		Where("crawl_status = ?", status).
		Count(&count).Error
	return count, err
}

func (r *universityRepository) UpdateCrawlResult(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&model.University{}).
		Where("id = ?", id).
		Updates(fields).Error
}
func (r *universityRepository) Update(ctx context.Context, id uuid.UUID, u *model.University) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Save(u).Error
}

func (r *universityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.University{}, "id = ?", id).Error
}
