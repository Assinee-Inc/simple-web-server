package model

type EbookRequest struct {
	Title            string  `validate:"required,min=5,max=120" json:"title"`
	Description      string  `validate:"required,max=120" json:"description"`
	SalesPage        string  `validate:"required" json:"sales_page"`
	Value            float64 `validate:"required,gt=0" json:"value"`
	PromotionalValue float64 `json:"promotional_value"`
	Status           bool    `json:"status"`
	Statistics       bool    `json:"statistics"`
}
