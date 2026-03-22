package service

import (
	"context"
	"io"

	"unimatch-be/internal/model"
	"unimatch-be/internal/repository"
	"unimatch-be/pkg/apperror"
	"unimatch-be/pkg/filestore"

	"github.com/google/uuid"
)

type documentService struct {
	repo    repository.CaseDocumentRepository
	actRepo repository.ActivityRepository
	store   filestore.FileStore
}

func NewDocumentService(repo repository.CaseDocumentRepository, actRepo repository.ActivityRepository, store filestore.FileStore) CaseDocumentService {
	return &documentService{
		repo:    repo,
		actRepo: actRepo,
		store:   store,
	}
}

func (s *documentService) Upload(ctx context.Context, caseID uuid.UUID, userID *uuid.UUID, fileName string, fileType string, fileSize int64, reader io.Reader) (*model.CaseDocument, *apperror.AppError) {
	// 1. Save to physical store
	relPath, err := s.store.Save(caseID, fileName, reader)
	if err != nil {
		return nil, apperror.Internal(err, "failed to save file to store")
	}

	// 2. Save metadata to DB
	doc := &model.CaseDocument{
		CaseID:       caseID,
		FileName:     fileName,
		FileType:     fileType,
		FileSize:     fileSize,
		FilePath:     relPath,
		UploadedByID: userID,
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		// Cleanup physical file if DB fails
		_ = s.store.Delete(relPath)
		return nil, apperror.Internal(err, "failed to save document metadata")
	}

	// 3. Log activity
	_ = s.actRepo.Create(ctx, &model.ActivityLog{
		CaseID:      &caseID,
		UserID:      userID,
		EventType:   "document_uploaded",
		Description: "Uploaded document: " + fileName,
	})

	return doc, nil
}

func (s *documentService) List(ctx context.Context, caseID uuid.UUID) ([]model.CaseDocument, *apperror.AppError) {
	docs, err := s.repo.FindByCaseID(ctx, caseID)
	if err != nil {
		return nil, apperror.Internal(err, "failed to list documents")
	}
	return docs, nil
}

func (s *documentService) GetByID(ctx context.Context, id uuid.UUID) (*model.CaseDocument, *apperror.AppError) {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.NotFound("document not found")
	}
	return doc, nil
}

func (s *documentService) GetPhysicalPath(doc *model.CaseDocument) string {
	return s.store.GetPath(doc.FilePath)
}

func (s *documentService) Delete(ctx context.Context, id uuid.UUID) *apperror.AppError {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return apperror.NotFound("document not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return apperror.Internal(err, "failed to delete document metadata")
	}

	_ = s.store.Delete(doc.FilePath)
	return nil
}
