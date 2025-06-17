package entity

import "gorm.io/gorm"

type Document struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	gorm.Model
	UserID   string `json:"user_id"`
	User     User   `json:"user"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	Digest   string `json:"digest"`
	Status   string `json:"status"` // uploaded, signed, etc.
}
