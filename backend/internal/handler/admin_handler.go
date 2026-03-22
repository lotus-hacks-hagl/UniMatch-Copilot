package handler

import (
	"net/http"
	"unimatch-be/internal/service"
	"unimatch-be/pkg/response"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminSvc service.AdminService
}

func NewAdminHandler(adminSvc service.AdminService) *AdminHandler {
	return &AdminHandler{adminSvc: adminSvc}
}

// ListTeachers godoc
// @Summary List all teachers
// @Description Get a list of all registered teachers (Admin only)
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=[]model.User}
// @Router /admin/teachers [get]
func (h *AdminHandler) ListTeachers(c *gin.Context) {
	teachers, appErr := h.adminSvc.ListTeachers(c.Request.Context())
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.OK(c, teachers)
}

type VerifyTeacherRequest struct {
	IsVerified bool `json:"is_verified"`
}

// VerifyTeacher godoc
// @Summary Verify/Unverify a teacher
// @Description Update the verification status of a teacher (Admin only)
// @Tags admin
// @Security BearerAuth
// @Accept json
// @Param id path string true "Teacher ID"
// @Param request body VerifyTeacherRequest true "Verification status"
// @Success 200 {object} response.Response
// @Router /admin/teachers/{id}/verify [put]
func (h *AdminHandler) VerifyTeacher(c *gin.Context) {
	teacherID := c.Param("id")
	var req VerifyTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	appErr := h.adminSvc.VerifyTeacher(c.Request.Context(), teacherID, req.IsVerified)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.OK(c, nil)
}
