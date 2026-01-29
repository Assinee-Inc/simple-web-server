package ebook

import (
	"context"
)

// InMemoryRepository é uma implementação simples do EbookRepositoryPort usada em testes.
type inMemoryRepository struct {
	store []*Ebook
}

func NewInMemoryRepository() Repository {
	return &inMemoryRepository{store: make([]*Ebook, 0)}
}

func (r *inMemoryRepository) Save(ctx context.Context, ebook *Ebook) error {
	r.store = append(r.store, ebook)
	return nil
}

func (r *inMemoryRepository) FindByParams(ctx context.Context, params ...any) ([]*Ebook, error) {
	// implementação mínima: retornar todos os ebooks armazenados.
	result := make([]*Ebook, len(r.store))
	copy(result, r.store)
	return result, nil
}
