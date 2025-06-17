package repository

import (
	"context"

	"github.com/PhanPhuc2609/be-sign-file/entity"
	"gorm.io/gorm"
)

type DocumentRepository interface {
	GetSignaturesByDocumentID(ctx context.Context, tx *gorm.DB, docID uint) ([]entity.Signature, error)

	Create(ctx context.Context, tx *gorm.DB, doc entity.Document) (entity.Document, error)
	FindByID(ctx context.Context, tx *gorm.DB, id uint) (entity.Document, error)
	FindByUserID(ctx context.Context, tx *gorm.DB, userID string) ([]entity.Document, error)
	FindByDigest(ctx context.Context, tx *gorm.DB, digest string, userID string) (entity.Document, error)
	Update(ctx context.Context, tx *gorm.DB, doc entity.Document) (entity.Document, error)
	Delete(ctx context.Context, tx *gorm.DB, id uint) error
}

type documentRepository struct {
	db *gorm.DB
}

// GetSignaturesByDocumentID returns all signatures for a document
func (r *documentRepository) GetSignaturesByDocumentID(ctx context.Context, tx *gorm.DB, docID uint) ([]entity.Signature, error) {
	var dbConn *gorm.DB
	if tx != nil {
		dbConn = tx
	} else {
		dbConn = r.db
	}
	var sigs []entity.Signature
	err := dbConn.WithContext(ctx).Where("document_id = ?", docID).Find(&sigs).Error
	return sigs, err
}

func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) Create(ctx context.Context, tx *gorm.DB, doc entity.Document) (entity.Document, error) {
	if tx == nil {
		tx = r.db
	}
	if err := tx.WithContext(ctx).Create(&doc).Error; err != nil {
		return entity.Document{}, err
	}
	return doc, nil
}

func (r *documentRepository) FindByID(ctx context.Context, tx *gorm.DB, id uint) (entity.Document, error) {
	if tx == nil {
		tx = r.db
	}
	var doc entity.Document
	if err := tx.WithContext(ctx).Preload("User").Where("id = ?", id).First(&doc).Error; err != nil {
		return entity.Document{}, err
	}
	return doc, nil
}

func (r *documentRepository) FindByUserID(ctx context.Context, tx *gorm.DB, userID string) ([]entity.Document, error) {
	if tx == nil {
		tx = r.db
	}
	var docs []entity.Document
	if err := tx.WithContext(ctx).Where("user_id = ?", userID).Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *documentRepository) FindByDigest(ctx context.Context, tx *gorm.DB, digest string, userID string) (entity.Document, error) {
	if tx == nil {
		tx = r.db
	}
	var doc entity.Document
	if err := tx.WithContext(ctx).Preload("User").Where("digest = ? AND user_id = ?", digest, userID).First(&doc).Error; err != nil {
		return entity.Document{}, err
	}
	return doc, nil
}

func (r *documentRepository) Update(ctx context.Context, tx *gorm.DB, doc entity.Document) (entity.Document, error) {
	if tx == nil {
		tx = r.db
	}
	if err := tx.WithContext(ctx).Save(&doc).Error; err != nil {
		return entity.Document{}, err
	}
	return doc, nil
}

func (r *documentRepository) Delete(ctx context.Context, tx *gorm.DB, id uint) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Delete(&entity.Document{}, id).Error
}
