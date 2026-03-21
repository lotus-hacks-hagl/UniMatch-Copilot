package service

import (
	"context"

	"unimatch-be/internal/dto"
	"unimatch-be/internal/model"
	"unimatch-be/internal/repository"
	"unimatch-be/pkg/apperror"
	"unimatch-be/pkg/response"

	"github.com/google/uuid"
)

type studentService struct {
	repo repository.StudentRepository
}

func NewStudentService(repo repository.StudentRepository) StudentService {
	return &studentService{repo: repo}
}

func (s *studentService) List(ctx context.Context, page, limit int) (*dto.ListStudentsResponse, *apperror.AppError) {
	students, total, err := s.repo.FindAll(ctx, page, limit)
	if err != nil {
		return nil, apperror.Internal(err, "failed to list students")
	}

	return &dto.ListStudentsResponse{
		Data: students,
		Meta: response.Meta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: int((total + int64(limit) - 1) / int64(limit)),
			HasNext:    int64(page*limit) < total,
		},
	}, nil
}

func (s *studentService) GetByID(ctx context.Context, id uuid.UUID) (*model.Student, *apperror.AppError) {
	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.NotFound("student not found")
	}
	return student, nil
}
