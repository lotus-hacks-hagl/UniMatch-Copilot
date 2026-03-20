package handler

import (
	"strconv"

	"unimatch-be/internal/service"
	"unimatch-be/pkg/response"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	svc service.DashboardService
}

func NewDashboardHandler(svc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// Stats godoc
// @Summary Get dashboard stats
// @Description Get high-level application statistics
// @Tags dashboard
// @Produce json
// @Success 200 {object} response.Response{data=dto.DashboardStats}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /dashboard/stats [get]
func (h *DashboardHandler) Stats(c *gin.Context) {
	stats, appErr := h.svc.GetStats(c.Request.Context())
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.OK(c, stats)
}

// CasesByDay godoc
// @Summary Get cases trend
// @Description Get cases created per day chart data
// @Tags dashboard
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.CasesByDay}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /dashboard/cases-by-day [get]
func (h *DashboardHandler) CasesByDay(c *gin.Context) {
	data, appErr := h.svc.GetCasesByDay(c.Request.Context())
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.OK(c, data)
}

// EscalationTrend godoc
// @Summary Get escalation trend
// @Description Get escalations per day chart data
// @Tags dashboard
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.EscalationTrend}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /dashboard/escalation-trend [get]
func (h *DashboardHandler) EscalationTrend(c *gin.Context) {
	data, appErr := h.svc.GetEscalationTrend(c.Request.Context())
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.OK(c, data)
}

// Analytics godoc
// @Summary Get detailed analytics
// @Description Get auto-approval rates and top universities/countries distributions
// @Tags dashboard
// @Produce json
// @Success 200 {object} response.Response{data=dto.Analytics}
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /dashboard/analytics [get]
func (h *DashboardHandler) Analytics(c *gin.Context) {
	data, appErr := h.svc.GetAnalytics(c.Request.Context())
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.OK(c, data)
}

// ActivityLog godoc
// @Summary Get recent activity log
// @Description Fetch paginated recent system events
// @Tags dashboard
// @Produce json
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response{error=swagger.SwaggerError}
// @Router /activity-log [get]
func (h *DashboardHandler) ActivityLog(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	logs, appErr := h.svc.GetActivityLog(c.Request.Context(), limit)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.OK(c, logs)
}
