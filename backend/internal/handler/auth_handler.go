package handler

import (
	"net/http"

	"unimatch-be/internal/dto"
	"unimatch-be/internal/service"
	"unimatch-be/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	svc       service.AuthService
	validator *validator.Validate
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc, validator: validator.New()}
}

// Register godoc
// @Summary Register internal user
// @Description Register the single internal admin account (fails if an account already exists)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Credentials"
// @Success 201 {object} response.Response{data=dto.AuthResponse}
// @Failure 400 {object} response.Response{error=swagger.SwaggerError}
// @Failure 403 {object} response.Response{error=swagger.SwaggerError}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if err := h.validator.Struct(&req); err != nil {
		response.FailWithDetails(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation failed", err.Error())
		return
	}

	result, appErr := h.svc.Register(c.Request.Context(), req)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.Created(c, result)
}

// Login godoc
// @Summary Login
// @Description Authenticate user and get JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Credentials"
// @Success 200 {object} response.Response{data=dto.AuthResponse}
// @Failure 400 {object} response.Response{error=swagger.SwaggerError}
// @Failure 401 {object} response.Response{error=swagger.SwaggerError}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if err := h.validator.Struct(&req); err != nil {
		response.FailWithDetails(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation failed", err.Error())
		return
	}

	result, appErr := h.svc.Login(c.Request.Context(), req)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.OK(c, result)
}
