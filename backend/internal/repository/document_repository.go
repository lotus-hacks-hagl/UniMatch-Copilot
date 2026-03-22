package repository

import (
	"context"

	"unimatch-be/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type documentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) CaseDocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) Create(ctx context.Context, d *model.CaseDocument) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *documentRepository) FindByCaseID(ctx context.Context, caseID uuid.UUID) ([]model.CaseDocument, error) {
	var docs []model.CaseDocument
	err := r.db.WithContext(ctx).Where("case_id = ?", caseID).Find(&docs).Error
	return docs, err
}

func (r *documentRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.CaseDocument, error) {
	var doc model.CaseDocument
	err := r.db.WithContext(ctx).First(&doc, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *documentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.CaseDocument{}, "id = ?", id).Error
}
