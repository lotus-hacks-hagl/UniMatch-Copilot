package handler

import (
	"net/http"
	"strconv"

	"unimatch-be/internal/service"
	"unimatch-be/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StudentHandler struct {
	svc service.StudentService
}

func NewStudentHandler(svc service.StudentService) *StudentHandler {
	return &StudentHandler{svc: svc}
}

func (h *StudentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	res, appErr := h.svc.List(c.Request.Context(), page, limit)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	response.Paginated(c, res.Data, res.Meta)
}

func (h *StudentHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_ID", "invalid student id")
		return
	}

	student, appErr := h.svc.GetByID(c.Request.Context(), id)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	response.OK(c, student)
}
