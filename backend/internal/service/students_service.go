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

func (s *studentService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateStudentRequest) *apperror.AppError {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return apperror.NotFound("student not found")
	}

	// Map request to model
	existing.FullName = req.FullName
	existing.GpaRaw = req.GpaRaw
	existing.GpaScale = req.GpaScale
	existing.GpaNormalized = req.GpaNormalized
	existing.IeltsOverall = req.IeltsOverall
	existing.SatTotal = req.SatTotal
	existing.ToeflTotal = req.ToeflTotal
	existing.IntendedMajor = req.IntendedMajor
	existing.BudgetUsdPerYear = req.BudgetUsdPerYear
	existing.PreferredCountries = req.PreferredCountries
	existing.TargetIntake = req.TargetIntake
	existing.ScholarshipRequired = req.ScholarshipRequired
	existing.Extracurriculars = req.Extracurriculars
	existing.Achievements = req.Achievements
	existing.PersonalStatementNotes = req.PersonalStatementNotes
	existing.BackgroundText = req.BackgroundText

	err = s.repo.Update(ctx, id, existing)
	if err != nil {
		return apperror.Internal(err, "failed to update student")
	}
	return nil
}

func (s *studentService) Delete(ctx context.Context, id uuid.UUID) *apperror.AppError {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return apperror.Internal(err, "failed to delete student")
	}
	return nil
}
