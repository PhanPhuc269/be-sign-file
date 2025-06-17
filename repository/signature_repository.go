package repository

import (
	"context"

	"github.com/PhanPhuc2609/be-sign-file/entity"
	"gorm.io/gorm"
)

type SignatureRepository interface {
	Create(ctx context.Context, tx *gorm.DB, sig entity.Signature) (entity.Signature, error)
	FindByID(ctx context.Context, tx *gorm.DB, id uint) (entity.Signature, error)
	FindByDocumentID(ctx context.Context, tx *gorm.DB, docID uint) ([]entity.Signature, error)
	Update(ctx context.Context, tx *gorm.DB, sig entity.Signature) (entity.Signature, error)
	Delete(ctx context.Context, tx *gorm.DB, id uint) error
}

type signatureRepository struct {
	db *gorm.DB
}

func NewSignatureRepository(db *gorm.DB) SignatureRepository {
	return &signatureRepository{db: db}
}

func (r *signatureRepository) Create(ctx context.Context, tx *gorm.DB, sig entity.Signature) (entity.Signature, error) {
	if tx == nil {
		tx = r.db
	}
	if err := tx.WithContext(ctx).Create(&sig).Error; err != nil {
		return entity.Signature{}, err
	}
	return sig, nil
}

func (r *signatureRepository) FindByID(ctx context.Context, tx *gorm.DB, id uint) (entity.Signature, error) {
	if tx == nil {
		tx = r.db
	}
	var sig entity.Signature
	if err := tx.WithContext(ctx).Preload("Document").Preload("Signer").Where("id = ?", id).First(&sig).Error; err != nil {
		return entity.Signature{}, err
	}
	return sig, nil
}

func (r *signatureRepository) FindByDocumentID(ctx context.Context, tx *gorm.DB, docID uint) ([]entity.Signature, error) {
	if tx == nil {
		tx = r.db
	}
	var sigs []entity.Signature
	if err := tx.WithContext(ctx).Where("document_id = ?", docID).Find(&sigs).Error; err != nil {
		return nil, err
	}
	return sigs, nil
}

func (r *signatureRepository) Update(ctx context.Context, tx *gorm.DB, sig entity.Signature) (entity.Signature, error) {
	if tx == nil {
		tx = r.db
	}
	if err := tx.WithContext(ctx).Save(&sig).Error; err != nil {
		return entity.Signature{}, err
	}
	return sig, nil
}

func (r *signatureRepository) Delete(ctx context.Context, tx *gorm.DB, id uint) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Delete(&entity.Signature{}, id).Error
}
