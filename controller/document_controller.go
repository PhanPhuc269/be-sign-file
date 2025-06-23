package controller

import (
	"net/http"
	"strconv"

	"github.com/PhanPhuc2609/be-sign-file/entity"
	"github.com/PhanPhuc2609/be-sign-file/service"
	"github.com/gin-gonic/gin"
)

type DocumentController interface {
	UploadDocument(c *gin.Context)
	GetDocumentByID(c *gin.Context)
	GetDocumentsByUserID(c *gin.Context)
	DeleteDocument(c *gin.Context)
	UploadAndVerifyDocument(c *gin.Context)
}

type documentController struct {
	service service.DocumentService
}

func NewDocumentController(service service.DocumentService) DocumentController {
	return &documentController{service: service}
}

// POST /api/documents/upload
func (ctrl *documentController) UploadDocument(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user_id in context"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
		return
	}
	uploadPath := "uploads/" + file.Filename
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot save file"})
		return
	}
	doc := entity.Document{
		UserID:   userIDStr,
		FileName: file.Filename,
		FilePath: uploadPath,
		Status:   "uploaded",
	}
	createdDoc, err := ctrl.service.CreateDocument(c.Request.Context(), doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, createdDoc)
}

// GET /api/documents/:id
func (ctrl *documentController) GetDocumentByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	doc, err := ctrl.service.GetDocumentByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}
	c.JSON(http.StatusOK, doc)
}

// GET /api/documents/user/:user_id
func (ctrl *documentController) GetDocumentsByUserID(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user_id in context"})
		return
	}
	docs, err := ctrl.service.GetDocumentsByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, docs)
}

// POST /api/documents/verify
func (ctrl *documentController) UploadAndVerifyDocument(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user_id in context"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
		return
	}

	verified, message, err := ctrl.service.UploadAndVerifyDocumentService(c.Request.Context(), userIDStr, file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"verified": false, "message": message, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"verified": verified, "message": message})
}

// DELETE /api/documents/:id
func (ctrl *documentController) DeleteDocument(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	err := ctrl.service.DeleteDocument(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Document deleted"})
}
