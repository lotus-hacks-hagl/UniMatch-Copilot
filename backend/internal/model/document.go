package model

import (
	"github.com/google/uuid"
)

type CaseDocument struct {
	Base
	CaseID       uuid.UUID `json:"case_id" gorm:"type:uuid;not null"`
	FileName     string    `json:"file_name" gorm:"not null"`
	FileType     string    `json:"file_type"`
	FileSize     int64     `json:"file_size"`
	FilePath     string    `json:"file_path" gorm:"not null"`
	UploadedByID *uuid.UUID `json:"uploaded_by_id" gorm:"type:uuid"`
	UploadedBy   *User      `json:"uploaded_by,omitempty" gorm:"foreignKey:UploadedByID"`
}

func (CaseDocument) TableName() string { return "case_documents" }
