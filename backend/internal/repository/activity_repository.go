package repository

import (
	"context"

	"unimatch-be/internal/model"

	"gorm.io/gorm"
)

type activityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) ActivityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) Create(ctx context.Context, log *model.ActivityLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *activityRepository) FindRecent(ctx context.Context, limit int) ([]model.ActivityLog, error) {
	var logs []model.ActivityLog
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
