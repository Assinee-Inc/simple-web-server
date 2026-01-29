package ebook_test

import (
	"context"
	"github.com/anglesson/simple-web-server/internal/ebook"
	"github.com/stretchr/testify/mock"
)

type MockEbookRepository struct {
	mock.Mock
}

func (m *MockEbookRepository) Save(ctx context.Context, ebook *ebook.Ebook) error {
	args := m.Called(ctx, ebook)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}

func (m *MockEbookRepository) FindByParams(ctx context.Context, params ...any) ([]*ebook.Ebook, error) {
	// Prepend ctx to params for the mock call
	allArgs := append([]interface{}{ctx}, params...)
	args := m.Called(allArgs...)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ebook.Ebook), args.Error(1)
}
