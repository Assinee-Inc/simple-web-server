package repositories

import "github.com/anglesson/simple-web-server/internal/library/domain"

type InMemoryEbookRepository struct {
	store []domain.Ebook
}

func NewInMemoryEbookRepository(connection any) *InMemoryEbookRepository {
	return &InMemoryEbookRepository{}
}
