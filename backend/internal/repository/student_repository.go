package repository

import (
	"context"

	"unimatch-be/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type studentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) StudentRepository {
	return &studentRepository{db: db}
}

func (r *studentRepository) FindAll(ctx context.Context, page, limit int) ([]model.Student, int64, error) {
	var students []model.Student
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Student{})
	
	q.Count(&total)
	err := q.Order("created_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&students).Error
	return students, total, err
}

func (r *studentRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Student{}).Count(&count).Error
	return count, err
}

func (r *studentRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Student, error) {
	var s model.Student
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *studentRepository) Update(ctx context.Context, id uuid.UUID, s *model.Student) error {
	return r.db.WithContext(ctx).Model(&model.Student{}).Where("id = ?", id).Updates(s).Error
}

func (r *studentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Student{}, "id = ?", id).Error
}
