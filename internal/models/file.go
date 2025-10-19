package models

import (
	"fmt"

	"github.com/anglesson/simple-web-server/pkg/utils"
	"gorm.io/gorm"
)

type File struct {
	gorm.Model
	Name                  string   `json:"name"`
	OriginalName          string   `json:"original_name"`
	NameNormalized        string   `json:"name_normalized" gorm:"type:text;index"`
	Description           string   `json:"description"`
	DescriptionNormalized string   `json:"description_normalized" gorm:"type:text;index"`
	FileType              string   `json:"file_type"` // pdf, doc, image, etc.
	FileSize              int64    `json:"file_size"` // em bytes
	S3Key                 string   `json:"s3_key"`
	S3URL                 string   `json:"s3_url"`
	Status                bool     `json:"status"` // ativo/inativo
	CreatorID             uint     `json:"creator_id"`
	Creator               Creator  `gorm:"foreignKey:CreatorID"`
	Ebooks                []*Ebook `gorm:"many2many:ebook_files"`
}

func NewFile(name, originalName, description, fileType, s3Key, s3URL string, fileSize int64, creatorID uint) *File {
	return &File{
		Name:                  name,
		OriginalName:          originalName,
		NameNormalized:        utils.NormalizeText(originalName),
		Description:           description,
		DescriptionNormalized: utils.NormalizeText(description),
		FileType:              fileType,
		FileSize:              fileSize,
		S3Key:                 s3Key,
		S3URL:                 s3URL,
		Status:                true,
		CreatorID:             creatorID,
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

func (f *File) SetName(name string) {
	f.Name = name
	f.NameNormalized = utils.NormalizeText(name)
}

func (f *File) SetDescription(description string) {
	f.Description = description
	f.DescriptionNormalized = utils.NormalizeText(description)
}
