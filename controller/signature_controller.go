package controller

import (
	"net/http"
	"strconv"

	"github.com/PhanPhuc2609/be-sign-file/entity"
	"github.com/PhanPhuc2609/be-sign-file/service"
	"github.com/gin-gonic/gin"
)

type SignatureController interface {
	CreateSignature(c *gin.Context)
	GetSignatureByID(c *gin.Context)
	GetSignaturesByDocumentID(c *gin.Context)
	DeleteSignature(c *gin.Context)
}

type signatureController struct {
	service service.SignatureService
}

func NewSignatureController(service service.SignatureService) SignatureController {
	return &signatureController{service: service}
}

// POST /api/signatures
func (ctrl *signatureController) CreateSignature(c *gin.Context) {
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
	var req struct {
		DocumentID uint `json:"document_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sig := entity.Signature{
		DocumentID: req.DocumentID,
		SignerID:   userIDStr,
	}
	createdSig, err := ctrl.service.CreateSignature(c.Request.Context(), sig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, createdSig)
}

// GET /api/signatures/:id
func (ctrl *signatureController) GetSignatureByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	sig, err := ctrl.service.GetSignatureByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Signature not found"})
		return
	}
	c.JSON(http.StatusOK, sig)
}

// GET /api/signatures/document/:doc_id
func (ctrl *signatureController) GetSignaturesByDocumentID(c *gin.Context) {
	docID, _ := strconv.ParseUint(c.Param("doc_id"), 10, 64)
	sigs, err := ctrl.service.GetSignaturesByDocumentID(c.Request.Context(), uint(docID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sigs)
}

// DELETE /api/signatures/:id
func (ctrl *signatureController) DeleteSignature(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	err := ctrl.service.DeleteSignature(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Signature deleted"})
}
