package service

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"

	"github.com/PhanPhuc2609/be-sign-file/entity"
	"github.com/PhanPhuc2609/be-sign-file/repository"
	"gorm.io/gorm"
)

type DocumentService interface {
	CreateDocument(ctx context.Context, doc entity.Document) (entity.Document, error)
	GetDocumentsByUserID(ctx context.Context, userID string) ([]entity.Document, error)
	UpdateDocument(ctx context.Context, doc entity.Document) (entity.Document, error)
	DeleteDocument(ctx context.Context, id uint) error
	GetDocumentByID(ctx context.Context, id uint) (entity.Document, error)
	FindDocumentByDigest(ctx context.Context, digest string, userID string) (entity.Document, error)
	GetSignaturesByDocumentID(ctx context.Context, docID uint) ([]entity.Signature, error)
	VerifySignature(ctx context.Context, sig entity.Signature, doc entity.Document) (bool, error)
	VerifySignatureRaw(ctx context.Context, sigBase64 string, content []byte, sig entity.Signature) (bool, error)
	UploadAndVerifyDocumentService(ctx context.Context, userID string, fileHeader *multipart.FileHeader) (bool, string, error)
}

type documentService struct {
	docRepo repository.DocumentRepository
	db      *gorm.DB
}

func (s *documentService) FindDocumentByDigest(ctx context.Context, digest string, userID string) (entity.Document, error) {
	return s.docRepo.FindByDigest(ctx, nil, digest, userID)
}

func NewDocumentService(docRepo repository.DocumentRepository, db *gorm.DB) DocumentService {
	return &documentService{
		docRepo: docRepo,
		db:      db,
	}
}

func (s *documentService) CreateDocument(ctx context.Context, doc entity.Document) (entity.Document, error) {
	// Đọc nội dung file tài liệu
	content, err := os.ReadFile(doc.FilePath)
	if err != nil {
		return entity.Document{}, errors.New("cannot read document file")
	}
	// Tạo digest SHA-256
	hash := sha256.Sum256(content)
	doc.Digest = hex.EncodeToString(hash[:])

	return s.docRepo.Create(ctx, nil, doc)
}

func (s *documentService) GetDocumentByID(ctx context.Context, id uint) (entity.Document, error) {
	return s.docRepo.FindByID(ctx, nil, id)
}

func (s *documentService) GetDocumentsByUserID(ctx context.Context, userID string) ([]entity.Document, error) {
	return s.docRepo.FindByUserID(ctx, nil, userID)
}

func (s *documentService) UpdateDocument(ctx context.Context, doc entity.Document) (entity.Document, error) {
	// Ensure document exists
	_, err := s.docRepo.FindByID(ctx, nil, doc.ID)
	if err != nil {
		return entity.Document{}, errors.New("document not found")
	}
	return s.docRepo.Update(ctx, nil, doc)
}

func (s *documentService) DeleteDocument(ctx context.Context, id uint) error {
	return s.docRepo.Delete(ctx, nil, id)
}

// Get signatures by document ID
func (s *documentService) GetSignaturesByDocumentID(ctx context.Context, docID uint) ([]entity.Signature, error) {
	return s.docRepo.GetSignaturesByDocumentID(ctx, nil, docID)
}

// Verify signature for a document
func (s *documentService) VerifySignature(ctx context.Context, sig entity.Signature, doc entity.Document) (bool, error) {
	// Giữ lại cho tương thích cũ
	return false, errors.New("Not implemented, use VerifySignatureRaw")
}

// Verify signature from raw signature in file
func (s *documentService) VerifySignatureRaw(ctx context.Context, sigBase64 string, content []byte, sig entity.Signature) (bool, error) {
	// Lấy public key từ chữ ký (ở đây demo dùng private key để lấy public key)
	privBytes, err := base64.StdEncoding.DecodeString(sig.PrivateKey)
	if err != nil {
		return false, errors.New("invalid private key encoding")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(privBytes)
	if err != nil {
		return false, errors.New("invalid private key")
	}
	publicKey := &privateKey.PublicKey

	// Để tương thích với cách ký: ký hash của hex digest
	// 1. Tính lại digest của nội dung file upload
	fileDigest := sha256.Sum256(content)
	fileDigestHex := hex.EncodeToString(fileDigest[:])
	// 2. Hash lại chuỗi hex digest này
	hashed := sha256.Sum256([]byte(fileDigestHex))

	signatureBytes, err := base64.StdEncoding.DecodeString(sigBase64)
	if err != nil {
		return false, errors.New("invalid signature encoding")
	}

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signatureBytes)
	if err != nil {
		return false, errors.New("signature verification failed")
	}
	return true, nil
}

// Upload and verify document logic moved from controller
func (s *documentService) UploadAndVerifyDocumentService(ctx context.Context, userID string, fileHeader *multipart.FileHeader) (bool, string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return false, "Cannot open uploaded file", err
	}
	defer file.Close()

	tempPath := "uploads/verify_" + fileHeader.Filename
	out, err := os.Create(tempPath)
	if err != nil {
		return false, "Cannot save file", err
	}
	_, err = io.Copy(out, file)
	out.Close()
	if err != nil {
		return false, "Cannot save file", err
	}
	defer os.Remove(tempPath)

	signedContent, err := os.ReadFile(tempPath)
	if err != nil {
		return false, "Cannot read uploaded file", err
	}
	parts := strings.Split(string(signedContent), "---BEGIN SIGNATURE---")
	if len(parts) < 2 {
		return false, "No signature found in file", errors.New("no signature")
	}
	signedParts := strings.SplitN(parts[1], "---END SIGNATURE---", 2)
	if len(signedParts) < 1 {
		return false, "No signature end marker in file", errors.New("no signature end marker")
	}
	sigBase64 := strings.TrimSpace(signedParts[0])
	originalContent := []byte(parts[0])

	digest := fmt.Sprintf("%x", sha256.Sum256(originalContent))

	doc, err := s.FindDocumentByDigest(ctx, digest, userID)
	if err != nil {
		return false, "Document not found by digest", err
	}
	sigs, err := s.GetSignaturesByDocumentID(ctx, doc.ID)
	if err != nil || len(sigs) == 0 {
		return false, "No signature found for this document", errors.New("no signature in db")
	}

	ok, err := s.VerifySignatureRaw(ctx, sigBase64, originalContent, sigs[0])
	if ok {
		return true, "Signature is valid", nil
	} else {
		return false, err.Error(), err
	}
}
