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
	count, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, apperror.Internal(err, "failed to count users")
	}

	// Check if user already exists
	existingUser, _ := s.userRepo.FindByUsername(ctx, req.Username)
	if existingUser != nil {
		return nil, apperror.Conflict("Username already taken")
	}
	
	role := "teacher"
	isVerified := false
	if count == 0 {
		role = "admin"
		isVerified = true
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Internal(err, "failed to hash password")
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Role:         role,
		IsVerified:   isVerified,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperror.Internal(err, "failed to create user")
	}

	token, expiresAt, appErr := s.generateToken(user)
	if appErr != nil {
		return nil, appErr
	}

	return &dto.AuthResponse{
		Token:      token,
		ExpiresIn:  expiresAt,
		Username:   user.Username,
		Role:       user.Role,
		IsVerified: user.IsVerified,
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
		Token:      token,
		ExpiresIn:  expiresAt,
		Username:   user.Username,
		Role:       user.Role,
		IsVerified: user.IsVerified,
	}, nil
}

func (s *authService) generateToken(user *model.User) (string, int64, *apperror.AppError) {
	exp := time.Now().Add(24 * time.Hour).Unix()
	claims := jwt.MapClaims{
		"user_id":     user.ID.String(),
		"username":    user.Username,
		"role":        user.Role,
		"is_verified": user.IsVerified,
		"exp":         exp,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", 0, apperror.Internal(err, "failed to generate token")
	}

	return tokenString, exp, nil
}
