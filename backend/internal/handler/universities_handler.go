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

type UniversitiesHandler struct {
	svc       service.UniversityService
	validator *validator.Validate
}

func NewUniversitiesHandler(svc service.UniversityService) *UniversitiesHandler {
	return &UniversitiesHandler{svc: svc, validator: validator.New()}
}

// List godoc
// @Summary List universities
// @Description Fetch paginated list of universities with optional filters
// @Tags universities
// @Produce json
// @Param country query string false "Filter by country"
// @Param search query string false "Search by name"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Response{meta=response.Meta}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Security BearerAuth
// @Router /universities [get]
func (h *UniversitiesHandler) List(c *gin.Context) {
	country := c.Query("country")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	unis, total, appErr := h.svc.List(c.Request.Context(), country, search, page, limit)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	response.Paginated(c, unis, response.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	})
}

// Create godoc
// @Summary Create university
// @Description Create a new university record manually
// @Tags universities
// @Accept json
// @Produce json
// @Param request body dto.CreateUniversityRequest true "University Information"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response{error=swagger.SwaggerError}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Security BearerAuth
// @Router /universities [post]
func (h *UniversitiesHandler) Create(c *gin.Context) {
	var req dto.CreateUniversityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if err := h.validator.Struct(&req); err != nil {
		response.FailWithDetails(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation failed", err.Error())
		return
	}

	uni, appErr := h.svc.Create(c.Request.Context(), req)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.Created(c, uni)
}

// CrawlAll godoc
// @Summary Trigger crawl for all universities
// @Description Submits async crawl jobs for universities that haven't been crawled in 24h
// @Tags universities
// @Produce json
// @Success 200 {object} response.Response{data=dto.CrawlAllResponse}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Security BearerAuth
// @Router /universities/crawl-all [post]
func (h *UniversitiesHandler) CrawlAll(c *gin.Context) {
	count, appErr := h.svc.CrawlAll(c.Request.Context())
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.OK(c, gin.H{"triggered": count, "message": "crawl jobs submitted"})
}

// CrawlActiveCount godoc
// @Summary Count active crawl jobs
// @Description Get the number of universities currently pending crawl results
// @Tags universities
// @Produce json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Security BearerAuth
// @Router /universities/crawl-active [get]
func (h *UniversitiesHandler) CrawlActiveCount(c *gin.Context) {
	count, appErr := h.svc.CountActiveCrawls(c.Request.Context())
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.OK(c, gin.H{"active_crawls": count})
}

// Crawl godoc
// @Summary Trigger crawl for a single university
// @Description Submits an async crawl job for a specific university
// @Tags universities
// @Produce json
// @Param id path string true "University ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response{error=swagger.SwaggerError}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Security BearerAuth
// @Router /universities/{id}/crawl [post]
func (h *UniversitiesHandler) Crawl(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_ID", "invalid university id")
		return
	}

	appErr := h.svc.Crawl(c.Request.Context(), id)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.OK(c, gin.H{"message": "crawl job submitted"})
}

// Delete godoc
// @Summary Delete university
// @Description Delete a university record and sync with AI graph
// @Tags universities
// @Produce json
// @Param id path string true "University ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response{error=swagger.SwaggerError}
// @Failure 404 {object} response.Response{error=swagger.SwaggerError}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Security BearerAuth
// @Router /universities/{id} [delete]
func (h *UniversitiesHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_ID", "invalid university id")
		return
	}

	appErr := h.svc.Delete(c.Request.Context(), id)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	response.OK(c, gin.H{"message": "university deleted and graph synced"})
}
