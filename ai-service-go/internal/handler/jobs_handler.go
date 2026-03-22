package handler

import (
	"net/http"

	"ai-service-go/internal/dto"
	"ai-service-go/internal/service"
	"ai-service-go/pkg/response"

	"github.com/gin-gonic/gin"
)

type JobsHandler struct {
	svc service.JobService
}

func NewJobsHandler(svc service.JobService) *JobsHandler {
	return &JobsHandler{svc: svc}
}

// GetJob godoc
// @Summary Get local job debug state
// @Description Returns local in-memory debug state for a job. Intended for local or test environments only.
// @Tags jobs
// @Produce json
// @Param job_id path string true "Job ID"
// @Success 200 {object} response.Response{data=dto.JobDebugResponse}
// @Failure 404 {object} response.Response
// @Router /jobs/{job_id} [get]
func (h *JobsHandler) GetJob(c *gin.Context) {
	jobID := c.Param("job_id")
	result, appErr := h.svc.GetJob(c.Request.Context(), jobID)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.OK(c, result)
}

// Analyze godoc
// @Summary Queue analyze job
// @Description Accepts a student profile and candidate universities, then asynchronously sends analyze results back to the backend callback URL.
// @Tags jobs
// @Accept json
// @Produce json
// @Param request body dto.AnalyzeJobRequest true "Analyze request"
// @Success 202 {object} response.Response{data=dto.JobAcceptedResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /jobs/analyze [post]
func (h *JobsHandler) Analyze(c *gin.Context) {
	var req dto.AnalyzeJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if appErr := h.svc.EnqueueAnalyze(c.Request.Context(), req); appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	response.Accepted(c, dto.JobAcceptedResponse{
		JobID:   req.JobID,
		Status:  "queued",
		Message: "Analyze job accepted",
	})
}

// Crawl godoc
// @Summary Queue crawl job
// @Description Accepts a university crawl job and asynchronously posts crawl results to the backend callback URL.
// @Tags jobs
// @Accept json
// @Produce json
// @Param request body dto.CrawlJobRequest true "Crawl request"
// @Success 202 {object} response.Response{data=dto.JobAcceptedResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /jobs/crawl [post]
func (h *JobsHandler) Crawl(c *gin.Context) {
	var req dto.CrawlJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if appErr := h.svc.EnqueueCrawl(c.Request.Context(), req); appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	response.Accepted(c, dto.JobAcceptedResponse{
		JobID:   req.JobID,
		Status:  "queued",
		Message: "Crawl job accepted",
	})
}

// Report godoc
// @Summary Queue report job
// @Description Accepts recommendations and asynchronously posts generated report content to the backend callback URL.
// @Tags jobs
// @Accept json
// @Produce json
// @Param request body dto.ReportJobRequest true "Report request"
// @Success 202 {object} response.Response{data=dto.JobAcceptedResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /jobs/report [post]
func (h *JobsHandler) Report(c *gin.Context) {
	var req dto.ReportJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if appErr := h.svc.EnqueueReport(c.Request.Context(), req); appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	response.Accepted(c, dto.JobAcceptedResponse{
		JobID:   req.JobID,
		Status:  "queued",
		Message: "Report job accepted",
	})
}
