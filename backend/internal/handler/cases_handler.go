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
// @Security BearerAuth
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
// @Security BearerAuth
// @Router /cases [get]
func (h *CasesHandler) List(c *gin.Context) {
	status := c.Query("status")
	assignedTo := c.Query("assigned_to")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var assignedToID *uuid.UUID
	var filterNone bool

	if assignedTo == "me" {
		uid, exists := c.Get("user_id")
		if exists {
			parsedID, _ := uuid.Parse(uid.(string))
			assignedToID = &parsedID
		}
	} else if assignedTo == "none" {
		filterNone = true
	} else if assignedTo != "" {
		parsedID, err := uuid.Parse(assignedTo)
		if err == nil {
			assignedToID = &parsedID
		}
	}

	cases, total, appErr := h.svc.List(c.Request.Context(), status, assignedToID, filterNone, page, limit)
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

// Claim godoc
// @Summary Claim a case
// @Description Take a case from the public pool (Teacher only)
// @Tags cases
// @Produce json
// @Param id path string true "Case UUID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response{error=swagger.SwaggerError}
// @Failure 401 {object} response.Response{error=swagger.SwaggerError}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Security BearerAuth
// @Router /cases/{id}/claim [post]
func (h *CasesHandler) Claim(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid case id")
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "user not authenticated")
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	appErr := h.svc.Claim(c.Request.Context(), caseID, userID)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	response.OK(c, nil)
}

// Count godoc
// @Summary Count cases
// @Description Get total cases count, optionally filtered by status
// @Tags cases
// @Produce json
// @Param status query string false "Filter by case status"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Security BearerAuth
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
// @Security BearerAuth
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
// @Security BearerAuth
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

func (h *CasesHandler) Update(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Fail(c, 400, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.svc.Update(c.Request.Context(), id, updates); err != nil {
		response.Fail(c, err.HTTPStatus, err.Code, err.Message)
		return
	}

	response.OK(c, nil)
}
