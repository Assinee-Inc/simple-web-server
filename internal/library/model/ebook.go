package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/anglesson/simple-web-server/pkg/utils"
	"gorm.io/gorm"
)

func (e *Ebook) BeforeCreate(tx *gorm.DB) error {
	if e.PublicID == "" {
		e.PublicID = utils.GeneratePublicID("ebk_")
	}
	return nil
}

type Ebook struct {
	gorm.Model
	PublicID              string  `json:"public_id" gorm:"type:varchar(40);uniqueIndex"`
	Title                 string  `json:"title"`
	TitleNormalized       string  `json:"title_normalized" gorm:"type:text;index"`
	Description           string  `json:"description"`
	DescriptionNormalized string  `json:"description_normalized" gorm:"type:text;index"`
	SalesPage             string  `json:"sales_page"` // Conteúdo da página de vendas
	Value                 float64 `json:"value"`
	PromotionalValue      float64 `json:"promotional_value"`
	Status                bool    `json:"status"`
	Image                 string  `json:"image"`
	Slug                  string  `json:"slug" gorm:"uniqueIndex"` // URL amigável
	CreatorID             uint    `json:"creator_id"`
	Files                 []*File `gorm:"many2many:ebook_files;"`
	Statistics            bool    `json:"statistics" gorm:"default:false"`

	// Campos para SEO e marketing
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	Keywords        string `json:"keywords"`

	// Estatísticas
	Views int `json:"views" gorm:"default:0"`
	Sales int `json:"sales" gorm:"default:0"`
}

func NewEbook(title, description, salesPage string, value, promotionalValue float64, creatorID uint, statistics bool) *Ebook {
	return &Ebook{
		Title:            title,
		Description:      description,
		SalesPage:        salesPage,
		Value:            value,
		PromotionalValue: promotionalValue,
		Status:           true,
		CreatorID:        creatorID,
		Slug:             GenerateSlug(title),
		Statistics:       statistics,
	}
}

func (e *Ebook) GetValue() string {
	return utils.FloatToBRL(e.Value)
}

func (e *Ebook) GetPromotionalValue() string {
	return fmt.Sprintf("%.2f", e.PromotionalValue)
}

func (e *Ebook) GetPromotionalValueBRL() string {
	return utils.FloatToBRL(e.PromotionalValue)
}

func (e *Ebook) GetLastUpdate() string {
	return e.UpdatedAt.Format("02-01-2006 15:04")
}

func (e *Ebook) AddFile(file *File) {
	e.Files = append(e.Files, file)
}

func (e *Ebook) RemoveFile(fileID uint) {
	for i, file := range e.Files {
		if file.ID == fileID {
			e.Files = append(e.Files[:i], e.Files[i+1:]...)
			break
		}
	}
}

func (e *Ebook) GetTotalFileSize() int64 {
	var total int64
	for _, file := range e.Files {
		total += file.FileSize
	}
	return total
}

func (e *Ebook) GetFileCount() int {
	return len(e.Files)
}

func (e *Ebook) IncrementViews() {
	e.Views++
}

func (e *Ebook) IncrementSales() {
	e.Sales++
}

// GetPresignedImageURL retorna a URL pré-assinada da imagem se disponível
func (e *Ebook) GetPresignedImageURL() string {
	if len(e.Image) > 100 {
		return e.Image
	}
	if e.Image == "" {
		return ""
	}
	return e.Image
}

// GenerateSlug cria uma URL amigável baseada no título
func GenerateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "ç", "c")
	slug = strings.ReplaceAll(slug, "ã", "a")
	slug = strings.ReplaceAll(slug, "á", "a")
	slug = strings.ReplaceAll(slug, "à", "a")
	slug = strings.ReplaceAll(slug, "â", "a")
	slug = strings.ReplaceAll(slug, "é", "e")
	slug = strings.ReplaceAll(slug, "ê", "e")
	slug = strings.ReplaceAll(slug, "í", "i")
	slug = strings.ReplaceAll(slug, "ó", "o")
	slug = strings.ReplaceAll(slug, "ô", "o")
	slug = strings.ReplaceAll(slug, "ú", "u")
	slug = strings.ReplaceAll(slug, "ü", "u")
	slug = strings.ReplaceAll(slug, "ñ", "n")

	reg := regexp.MustCompile("[^a-z0-9-]")
	slug = reg.ReplaceAllString(slug, "")

	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	slug = strings.Trim(slug, "-")

	return slug
}

func (e *Ebook) HasPromotion() bool {
	return e.PromotionalValue > 0
}

func (e *Ebook) ShowStatistics() bool {
	return e.Statistics
}

func (e *Ebook) GetFinalValue() float64 {
	if e.HasPromotion() {
		return e.PromotionalValue
	}
	return e.Value
}

func (e *Ebook) GetEconomy() string {
	savings := e.Value - e.PromotionalValue
	return utils.FloatToBRL(savings)
}
