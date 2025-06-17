package entity

import "gorm.io/gorm"

type Signature struct {
	gorm.Model
	DocumentID   uint      `json:"document_id"`
	Document     Document  `json:"document"`
	SignerID     string    `json:"signer_id"`
	Signer       User      `json:"signer"`
	SignatureRaw string    `json:"signature_raw"`
	Algorithm    string    `json:"algorithm"`
	SignedAt     int64     `json:"signed_at"`
	PrivateKey   string    `json:"private_key"`
}
