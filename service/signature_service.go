package service

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/PhanPhuc2609/be-sign-file/entity"
	"github.com/PhanPhuc2609/be-sign-file/repository"
	"gorm.io/gorm"
)

type SignatureService interface {
	CreateSignature(ctx context.Context, sig entity.Signature) (entity.Signature, error)
	GetSignatureByID(ctx context.Context, id uint) (entity.Signature, error)
	GetSignaturesByDocumentID(ctx context.Context, docID uint) ([]entity.Signature, error)
	UpdateSignature(ctx context.Context, sig entity.Signature) (entity.Signature, error)
	DeleteSignature(ctx context.Context, id uint) error
}

type signatureService struct {
	sigRepo  repository.SignatureRepository
	docRepo  repository.DocumentRepository
	userRepo repository.UserRepository
	db       *gorm.DB
}

func NewSignatureService(sigRepo repository.SignatureRepository, docRepo repository.DocumentRepository, userRepo repository.UserRepository, db *gorm.DB) SignatureService {
	return &signatureService{
		sigRepo:  sigRepo,
		docRepo:  docRepo,
		userRepo: userRepo,
		db:       db,
	}
}

func (s *signatureService) CreateSignature(ctx context.Context, sig entity.Signature) (entity.Signature, error) {
	// Ensure document exists
	doc, err := s.docRepo.FindByID(ctx, nil, sig.DocumentID)
	if err != nil {
		return entity.Signature{}, errors.New("document not found")
	}
	// Ensure signer exists
	_, err = s.userRepo.GetUserById(ctx, nil, sig.SignerID)
	if err != nil {
		return entity.Signature{}, errors.New("signer not found")
	}

	// Mô phỏng sinh private key RSA (trong thực tế nên lưu key ở nơi an toàn)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return entity.Signature{}, errors.New("failed to generate private key")
	}

	// Digest của tài liệu (đã lưu trong doc.Digest)
	digestBytes := []byte(doc.Digest)
	hashed := sha256.Sum256(digestBytes)

	// Ký digest
	signatureBytes, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return entity.Signature{}, errors.New("failed to sign digest")
	}
	sig.SignatureRaw = base64.StdEncoding.EncodeToString(signatureBytes)
	sig.Algorithm = "RSA"
	sig.SignedAt = time.Now().Unix()

	// Lưu private key dưới dạng base64 (chỉ dùng cho mô phỏng/demo, KHÔNG dùng cho production)
	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	sig.PrivateKey = base64.StdEncoding.EncodeToString(privBytes)

	// Đính chữ ký vào file (tạo file mới .signed)
	signedFilePath := doc.FilePath + ".signed"
	originalContent, err := os.ReadFile(doc.FilePath)
	if err != nil {
		return entity.Signature{}, errors.New("cannot read original file to append signature")
	}
	// Thêm marker đúng chuẩn, không thêm thừa dòng trống
	var signedContent []byte
	if len(originalContent) > 0 && originalContent[len(originalContent)-1] == '\n' {
		signedContent = append(originalContent, []byte(fmt.Sprintf("---BEGIN SIGNATURE---\n%s\n---END SIGNATURE---\n", sig.SignatureRaw))...)
	} else {
		signedContent = append(originalContent, []byte(fmt.Sprintf("\n---BEGIN SIGNATURE---\n%s\n---END SIGNATURE---\n", sig.SignatureRaw))...)
	}
	if err := os.WriteFile(signedFilePath, signedContent, 0644); err != nil {
		return entity.Signature{}, errors.New("cannot write signed file")
	}

	return s.sigRepo.Create(ctx, nil, sig)
}

func (s *signatureService) GetSignatureByID(ctx context.Context, id uint) (entity.Signature, error) {
	return s.sigRepo.FindByID(ctx, nil, id)
}

func (s *signatureService) GetSignaturesByDocumentID(ctx context.Context, docID uint) ([]entity.Signature, error) {
	return s.sigRepo.FindByDocumentID(ctx, nil, docID)
}

func (s *signatureService) UpdateSignature(ctx context.Context, sig entity.Signature) (entity.Signature, error) {
	_, err := s.sigRepo.FindByID(ctx, nil, sig.ID)
	if err != nil {
		return entity.Signature{}, errors.New("signature not found")
	}
	return s.sigRepo.Update(ctx, nil, sig)
}

func (s *signatureService) DeleteSignature(ctx context.Context, id uint) error {
	return s.sigRepo.Delete(ctx, nil, id)
}
func (s *signatureService) VerifySignature(ctx context.Context, sig entity.Signature, doc entity.Document) (bool, error) {
	// Giải mã private key (ở demo bạn lưu private key, thực tế sẽ dùng public key)
	privBytes, err := base64.StdEncoding.DecodeString(sig.PrivateKey)
	if err != nil {
		return false, errors.New("invalid private key encoding")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(privBytes)
	if err != nil {
		return false, errors.New("invalid private key")
	}
	publicKey := &privateKey.PublicKey

	// Lấy digest gốc của tài liệu
	digestBytes := []byte(doc.Digest)
	hashed := sha256.Sum256(digestBytes)

	// Giải mã chữ ký
	signatureBytes, err := base64.StdEncoding.DecodeString(sig.SignatureRaw)
	if err != nil {
		return false, errors.New("invalid signature encoding")
	}

	// Xác minh chữ ký
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signatureBytes)
	if err != nil {
		return false, errors.New("signature verification failed")
	}
	return true, nil
}
