package ebook

import (
	"context"
	"time"
)

// Ebook é a entidade de domínio do pacote ebook.
type Ebook struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	SalesDescription string `json:"sales_description"`
	// preços são representados em centavos (int64) conforme as regras do projeto
	Price            int64     `json:"price"`
	PromotionalPrice int64     `json:"promotional_price"`
	CoverImage       string    `json:"cover_image"`
	InfoProducerID   string    `json:"info_producer_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	FileIDs          []string  `json:"file_ids"`
}

// contrato do repositório do domínio ebook
type Repository interface {
	Save(ctx context.Context, ebook *Ebook) error
	FindByParams(ctx context.Context, params ...any) ([]*Ebook, error)
}

// note: ManagerPort pode ser usado por testes/mocks externos
type ManagerPort interface {
	CreateEbook(ctx context.Context, ebook *Ebook) (*Ebook, error)
}
