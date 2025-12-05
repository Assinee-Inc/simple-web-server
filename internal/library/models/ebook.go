package models

import (
	"github.com/anglesson/simple-web-server/pkg/validator"
)

type Ebook struct {
	ID                  string `json:"id"`
	Title               string `json:"title" validate:"required,max=50"`
	Description         string `json:"description" validate:"max=120"`
	SalesDescription    string `json:"sales_description" validate:"max=120"`
	Price               int64  `json:"price" validate:"required,min=1"`
	PromotionalPrice    int64  `json:"promotional_price" validate:"ltfield=Price"`
	CoverImage          string `json:"cover_image"`
	Available           string `json:"available"`
	ShowSalesStatistics bool   `json:"show_sales_statistics"`
	InfoProducerID      string `json:"info_produtor_id" validate:"required"`
}

func NewEbook(
	title,
	description,
	salesDescription,
	coverImage,
	available string,
	price int64,
	promotionalPrice int64,
	showSalesStatistics bool,
	infoProducerID string,
) *Ebook {
	return &Ebook{
		Title:               title,
		Description:         description,
		SalesDescription:    salesDescription,
		Price:               price,
		PromotionalPrice:    promotionalPrice,
		CoverImage:          coverImage,
		Available:           available,
		ShowSalesStatistics: showSalesStatistics,
		InfoProducerID:      infoProducerID,
	}
}

// Validate: ebook creation
func (e *Ebook) Validate() error {
	if err := validator.Validate(e); err != nil {
		return err
	}
	return nil
}

func (e *Ebook) promotionalValueIsLessThanValue() bool {
	if e.PromotionalPrice == 0 {
		return true
	}
	return e.PromotionalPrice < e.Price
}
