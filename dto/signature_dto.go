package dto

type SignDocumentRequest struct {
	Algorithm string `json:"algorithm" binding:"required"`
}

type SignatureResponse struct {
	ID           uint   `json:"id"`
	DocumentID   uint   `json:"document_id"`
	SignerID     uint   `json:"signer_id"`
	SignatureRaw string `json:"signature_raw"`
	Algorithm    string `json:"algorithm"`
	SignedAt     int64  `json:"signed_at"`
}
