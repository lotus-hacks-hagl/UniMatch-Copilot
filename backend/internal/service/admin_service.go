package service

import (
	"context"
	"unimatch-be/internal/model"
	"unimatch-be/internal/repository"
	"unimatch-be/pkg/apperror"
)

type AdminService interface {
	ListTeachers(ctx context.Context) ([]*model.User, *apperror.AppError)
	VerifyTeacher(ctx context.Context, teacherID string, verify bool) *apperror.AppError
}

type adminService struct {
	userRepo repository.UserRepository
}

func NewAdminService(userRepo repository.UserRepository) AdminService {
	return &adminService{userRepo: userRepo}
}

func (s *adminService) ListTeachers(ctx context.Context) ([]*model.User, *apperror.AppError) {
	teachers, err := s.userRepo.ListTeachers(ctx)
	if err != nil {
		return nil, apperror.Internal(err, "failed to list teachers")
	}
	return teachers, nil
}

func (s *adminService) VerifyTeacher(ctx context.Context, teacherID string, verify bool) *apperror.AppError {
	if err := s.userRepo.UpdateVerificationStatus(ctx, teacherID, verify); err != nil {
		return apperror.Internal(err, "failed to update teacher verification status")
	}
	return nil
}
