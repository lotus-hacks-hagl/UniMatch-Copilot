package handler

import (
	"net/http"

	"unimatch-be/internal/service"
	"unimatch-be/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DocumentHandler struct {
	svc service.CaseDocumentService
}

func NewDocumentHandler(svc service.CaseDocumentService) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

// @Summary Upload a document for a case
// @Tags Documents
// @Accept multipart/form-data
// @Produce json
// @Param case_id path string true "Case ID"
// @Param file formData file true "File to upload"
// @Success 201 {object} model.CaseDocument
// @Router /cases/{case_id}/documents [post]
func (h *DocumentHandler) Upload(c *gin.Context) {
	caseIDStr := c.Param("case_id")
	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_CASE_ID", "Invalid case ID")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "MISSING_FILE", "File is required")
		return
	}

	src, err := file.Open()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "FILE_OPEN_ERROR", "Failed to open file")
		return
	}
	defer src.Close()

	var userID *uuid.UUID
	uid, exists := c.Get("user_id")
	if exists {
		parsedUID, _ := uuid.Parse(uid.(string))
		userID = &parsedUID
	}

	doc, appErr := h.svc.Upload(c.Request.Context(), caseID, userID, file.Filename, file.Header.Get("Content-Type"), file.Size, src)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	response.Created(c, doc)
}

// @Summary List documents for a case
// @Tags Documents
// @Produce json
// @Param case_id path string true "Case ID"
// @Success 200 {array} model.CaseDocument
// @Router /cases/{case_id}/documents [get]
func (h *DocumentHandler) List(c *gin.Context) {
	caseIDStr := c.Param("case_id")
	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_CASE_ID", "Invalid case ID")
		return
	}

	docs, appErr := h.svc.List(c.Request.Context(), caseID)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	response.OK(c, docs)
}

// @Summary Download a document
// @Tags Documents
// @Param id path string true "Document ID"
// @Success 200 {file} file
// @Router /documents/{id} [get]
func (h *DocumentHandler) Download(c *gin.Context) {
	docIDStr := c.Param("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_DOC_ID", "Invalid document ID")
		return
	}

	doc, appErr := h.svc.GetByID(c.Request.Context(), docID)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	physicalPath := h.svc.GetPhysicalPath(doc)
	c.Header("Content-Disposition", "attachment; filename="+doc.FileName)
	c.Header("Content-Type", doc.FileType)
	c.File(physicalPath)
}

// @Summary Delete a document
// @Tags Documents
// @Param id path string true "Document ID"
// @Success 200 {object} response.StandardResponse
// @Router /documents/{id} [delete]
func (h *DocumentHandler) Delete(c *gin.Context) {
	docIDStr := c.Param("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_DOC_ID", "Invalid document ID")
		return
	}

	appErr := h.svc.Delete(c.Request.Context(), docID)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	response.OK(c, nil)
}
