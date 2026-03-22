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
		Preload("Documents").
		Preload("ActivityLogs.User").
		Preload("Recommendations", func(db *gorm.DB) *gorm.DB {
			return db.Order("rank_order ASC")
		}).
		First(&c, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

func (r *caseRepository) FindAll(ctx context.Context, status string, assignedToID *uuid.UUID, filterNone bool, search string, page, limit int) ([]model.Case, int64, error) {
	var cases []model.Case
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Case{}).Joins("Student").Preload("Student")
	if status != "" && status != "all" {
		q = q.Where("cases.status = ?", status)
	}

	if assignedToID != nil {
		q = q.Where("cases.assigned_to_id = ?", assignedToID)
	} else if filterNone {
		q = q.Where("cases.assigned_to_id IS NULL")
	}

	if search != "" {
		q = q.Where("(\"Student\".full_name ILIKE ? OR \"Student\".intended_major ILIKE ?)", "%"+search+"%", "%"+search+"%")
	}

	q.Count(&total)
	err := q.Order("cases.created_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&cases).Error
	return cases, total, err
}

func (r *caseRepository) Claim(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var c model.Case
		if err := tx.First(&c, "id = ?", id).Error; err != nil {
			return err
		}
		if c.AssignedToID != nil {
			return errors.New("case already assigned")
		}
		return tx.Model(&c).Update("assigned_to_id", userID).Error
	})
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

func (r *caseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Case{}, "id = ?", id).Error
}
