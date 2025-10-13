package models

import (
	"fmt"

	"gorm.io/gorm"
)

type File struct {
	gorm.Model
	Name         string   `json:"name"`
	OriginalName string   `json:"original_name"`
	Description  string   `json:"description"`
	FileType     string   `json:"file_type"` // pdf, doc, image, etc.
	FileSize     int64    `json:"file_size"` // em bytes
	S3Key        string   `json:"s3_key"`
	S3URL        string   `json:"s3_url"`
	Status       bool     `json:"status"` // ativo/inativo
	CreatorID    uint     `json:"creator_id"`
	Creator      Creator  `gorm:"foreignKey:CreatorID"`
	Ebooks       []*Ebook `gorm:"many2many:ebook_files"`
}

func NewFile(name, originalName, description, fileType, s3Key, s3URL string, fileSize int64, creatorID uint) *File {
	return &File{
		Name:         name,
		OriginalName: originalName,
		Description:  description,
		FileType:     fileType,
		FileSize:     fileSize,
		S3Key:        s3Key,
		S3URL:        s3URL,
		Status:       true,
		CreatorID:    creatorID,
	}
}

func (f *File) GetFileSizeFormatted() string {
	const unit = 1024
	if f.FileSize < unit {
		return fmt.Sprintf("%d B", f.FileSize)
	}
	div, exp := int64(unit), 0
	for n := f.FileSize / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(f.FileSize)/float64(div), "KMGTPE"[exp])
}

func (f *File) IsPDF() bool {
	return f.FileType == "pdf"
}

func (f *File) IsImage() bool {
	return f.FileType == "image"
}

func (f *File) InUse() bool {
	return len(f.Ebooks) > 0
}
