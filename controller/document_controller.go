package controller

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	userIDStr, _ := userIDVal.(string)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
		return
	}

	tempPath := "uploads/verify_" + file.Filename
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot save file"})
		return
	}
	defer func() { _ = os.Remove(tempPath) }()

	signedContent, err := os.ReadFile(tempPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot read uploaded file"})
		return
	}
	parts := strings.Split(string(signedContent), "---BEGIN SIGNATURE---")
	if len(parts) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No signature found in file"})
		return
	}
	signedParts := strings.SplitN(parts[1], "---END SIGNATURE---", 2)
	if len(signedParts) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No signature end marker in file"})
		return
	}
	sigBase64 := strings.TrimSpace(signedParts[0])
	originalContent := []byte(parts[0])

	digest := fmt.Sprintf("%x", sha256.Sum256(originalContent))

	// Tìm document theo digest
	doc, err := ctrl.service.FindDocumentByDigest(c.Request.Context(), digest, userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found by digest"})
		return
	}
	sigs, err := ctrl.service.GetSignaturesByDocumentID(c.Request.Context(), doc.ID)
	if err != nil || len(sigs) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No signature found for this document"})
		return
	}

	ok, err := ctrl.service.VerifySignatureRaw(c.Request.Context(), sigBase64, originalContent, sigs[0])
	if ok {
		c.JSON(http.StatusOK, gin.H{"verified": true, "message": "Signature is valid"})
	} else {
		c.JSON(http.StatusOK, gin.H{"verified": false, "message": err.Error()})
	}
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
