package repositories

import "github.com/anglesson/simple-web-server/internal/library/models"

type InMemoryEbookRepository struct {
	store []models.Ebook
}

func NewInMemoryEbookRepository(connection any) *InMemoryEbookRepository {
	return &InMemoryEbookRepository{}
}
