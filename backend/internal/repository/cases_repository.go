package repository

import (
	"context"
	"errors"

	"unimatch-be/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type caseRepository struct {
	db *gorm.DB
}

func NewCaseRepository(db *gorm.DB) CaseRepository {
	return &caseRepository{db: db}
}

func (r *caseRepository) Create(ctx context.Context, c *model.Case) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *caseRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Case, error) {
	var c model.Case
	err := r.db.WithContext(ctx).
		Preload("Student").
		Preload("Recommendations", func(db *gorm.DB) *gorm.DB {
			return db.Order("rank_order ASC")
		}).
		First(&c, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

func (r *caseRepository) FindAll(ctx context.Context, status string, page, limit int) ([]model.Case, int64, error) {
	var cases []model.Case
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Case{}).Preload("Student")
	if status != "" && status != "all" {
		q = q.Where("status = ?", status)
	}

	q.Count(&total)
	err := q.Order("created_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&cases).Error
	return cases, total, err
}

func (r *caseRepository) Update(ctx context.Context, c *model.Case) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *caseRepository) UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&model.Case{}).
		Where("id = ?", id).
		Updates(fields).Error
}

func (r *caseRepository) Count(ctx context.Context, status string) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&model.Case{})
	if status != "" && status != "all" {
		q = q.Where("status = ?", status)
	}
	err := q.Count(&count).Error
	return count, err
}
