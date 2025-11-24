package domain

import "errors"

type Ebook struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	SalesDescription    string `json:"sales_description"`
	Price               int64
	PromotionalPrice    int64  `json:"promotional_price"`
	CoverImage          string `json:"cover_image"`
	Available           string `json:"available"`
	ShowSalesStatistics bool   `json:"show_sales_statistics"`
	InfoProducerID      string `json:"info_produtor_id"`
}

func NewEbook(title, description, salesDescription, coverImage, available string, price int64, promotionalPrice int64, showSalesStatistics bool, infoProducerID string) *Ebook {
	return &Ebook{
		Title:               title,
		Description:         description,
		SalesDescription:    salesDescription,
		Price:               price,
		PromotionalPrice:    promotionalPrice,
		CoverImage:          coverImage,
		Available:           available,
		ShowSalesStatistics: showSalesStatistics,
	}
}

// Validate: ebook creation
func (e *Ebook) Validate() error {
	if !e.promotionalValueIsLessThanValue() {
		return errors.New("promotional Price cannot be greater than value")
	}

	if len(e.Title) > 50 {
		return errors.New("title cannot be longer than 50 characters")
	}

	if len(e.Description) > 120 {
		return errors.New("description cannot be longer than 120 characters")
	}

	if len(e.SalesDescription) > 120 {
		return errors.New("sales Description cannot be longer than 120 characters")
	}

	return nil
}

func (e *Ebook) promotionalValueIsLessThanValue() bool {
	if e.PromotionalPrice == 0 {
		return true
	}
	return e.PromotionalPrice < e.Price
}
