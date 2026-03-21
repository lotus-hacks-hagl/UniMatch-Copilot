package service

import (
	"context"
	"time"

	"unimatch-be/config"
	"unimatch-be/internal/dto"
	"unimatch-be/internal/model"
	"unimatch-be/internal/repository"
	"unimatch-be/pkg/apperror"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, *apperror.AppError)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, *apperror.AppError)
}

type authService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{userRepo: userRepo, cfg: cfg}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, *apperror.AppError) {
	// Rule: Maintain single internal account
	count, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, apperror.Internal(err, "failed to count users")
	}
	if count > 0 {
		return nil, apperror.Forbidden("only one internal account is allowed, registration blocked")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Internal(err, "failed to hash password")
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Role:         "admin",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperror.Internal(err, "failed to create user")
	}

	token, expiresAt, appErr := s.generateToken(user)
	if appErr != nil {
		return nil, appErr
	}

	return &dto.AuthResponse{
		Token:     token,
		ExpiresIn: expiresAt,
		Username:  user.Username,
		Role:      user.Role,
	}, nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, *apperror.AppError) {
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		// Do not differentiate between "user not found" and "wrong password" for security
		return nil, apperror.Unauthorized("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperror.Unauthorized("invalid username or password")
	}

	token, expiresAt, appErr := s.generateToken(user)
	if appErr != nil {
		return nil, appErr
	}

	return &dto.AuthResponse{
		Token:     token,
		ExpiresIn: expiresAt,
		Username:  user.Username,
		Role:      user.Role,
	}, nil
}

func (s *authService) generateToken(user *model.User) (string, int64, *apperror.AppError) {
	exp := time.Now().Add(24 * time.Hour).Unix()
	claims := jwt.MapClaims{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"role":     user.Role,
		"exp":      exp,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", 0, apperror.Internal(err, "failed to generate token")
	}

	return tokenString, exp, nil
}
