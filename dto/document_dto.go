package dto

type UploadDocumentRequest struct {
	FileName string `json:"file_name" binding:"required"`
	FileData []byte `json:"file_data" binding:"required"`
}

type DocumentResponse struct {
	ID       uint   `json:"id"`
	UserID   uint   `json:"user_id"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	Digest   string `json:"digest"`
	Status   string `json:"status"`
}
