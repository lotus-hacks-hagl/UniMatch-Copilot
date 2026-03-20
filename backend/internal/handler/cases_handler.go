package handler

import (
	"net/http"
	"strconv"

	"unimatch-be/internal/dto"
	"unimatch-be/internal/service"
	"unimatch-be/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CasesHandler struct {
	svc       service.CaseService
	validator *validator.Validate
}

func NewCasesHandler(svc service.CaseService) *CasesHandler {
	return &CasesHandler{svc: svc, validator: validator.New()}
}

// Create godoc
// @Summary Create a new case
// @Description Create student and case records, triggers AI profiling
// @Tags cases
// @Accept json
// @Produce json
// @Param request body dto.CreateCaseRequest true "Student Information"
// @Success 201 {object} response.Response{data=dto.CaseCreatedResponse}
// @Failure 400 {object} response.Response{error=swagger.SwaggerError}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /cases [post]
func (h *CasesHandler) Create(c *gin.Context) {
	var req dto.CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if err := h.validator.Struct(&req); err != nil {
		response.FailWithDetails(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation failed", err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
		return
	}

	result, appErr := h.svc.Create(c.Request.Context(), req)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.Created(c, result)
}

// List godoc
// @Summary List cases
// @Description Fetch paginated list of cases with optional status filter
// @Tags cases
// @Produce json
// @Param status query string false "Filter by case status: pending, processing, done, human_review, failed"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Response{meta=response.Meta}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /cases [get]
func (h *CasesHandler) List(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	cases, total, appErr := h.svc.List(c.Request.Context(), status, page, limit)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	response.Paginated(c, cases, response.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	})
}

// Count godoc
// @Summary Count cases
// @Description Get total cases count, optionally filtered by status
// @Tags cases
// @Produce json
// @Param status query string false "Filter by case status"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /cases/count [get]
func (h *CasesHandler) Count(c *gin.Context) {
	status := c.Query("status")
	count, appErr := h.svc.Count(c.Request.Context(), status)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.OK(c, gin.H{"count": count})
}

// GetByID godoc
// @Summary Get case details
// @Description Retrieve a case and its active recommendations by UUID
// @Tags cases
// @Produce json
// @Param id path string true "Case UUID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response{error=swagger.SwaggerError}
// @Failure 404 {object} response.Response{error=swagger.SwaggerError}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /cases/{id} [get]
func (h *CasesHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid case id")
		return
	}
	caseRecord, appErr := h.svc.GetByID(c.Request.Context(), id)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.OK(c, caseRecord)
}

// RequestReport godoc
// @Summary Request PDF report generation
// @Description Queue an async report generation job for a ready case
// @Tags cases
// @Produce json
// @Param id path string true "Case UUID"
// @Success 200 {object} response.Response{data=dto.ReportStatusResponse}
// @Failure 400 {object} response.Response{error=swagger.SwaggerError}
// @Failure 404 {object} response.Response{error=swagger.SwaggerError}
// @Failure 503 {object} response.Response{error=swagger.SwaggerError}
// @Router /cases/{id}/report [post]
func (h *CasesHandler) RequestReport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid case id")
		return
	}
	result, appErr := h.svc.RequestReport(c.Request.Context(), id)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.OK(c, result)
}
