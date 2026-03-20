package handler

import (
	"net/http"

	"unimatch-be/internal/dto"
	"unimatch-be/internal/service"

	"github.com/gin-gonic/gin"
)

// InternalHandler handles callbacks from AI Service — not exposed to public
type InternalHandler struct {
	caseSvc service.CaseService
	uniSvc  service.UniversityService
}

func NewInternalHandler(caseSvc service.CaseService, uniSvc service.UniversityService) *InternalHandler {
	return &InternalHandler{caseSvc: caseSvc, uniSvc: uniSvc}
}

// JobDone handles POST /internal/jobs/done from AI Service
func (h *InternalHandler) JobDone(c *gin.Context) {
	var payload dto.JobDonePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Route to appropriate service based on job type
	var appErr interface{ Error() string }
	switch payload.JobType {
	case "crawl_university":
		// Crawl callbacks handled by university service (via type assertion)
		if uSvc, ok := h.uniSvc.(interface {
			HandleCrawlDone(c interface{ Deadline() (interface{}, bool) }, p dto.JobDonePayload) interface{ Error() string }
		}); ok {
			_ = uSvc
		}
		// Delegate directly
		appErr = h.caseSvc.HandleJobDone(c.Request.Context(), payload)
	default:
		appErr = h.caseSvc.HandleJobDone(c.Request.Context(), payload)
	}

	if appErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": appErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}
