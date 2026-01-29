package ebook_test

import (
	"context"
	"github.com/anglesson/simple-web-server/internal/ebook"
	"github.com/stretchr/testify/mock"
)

type MockEbookManager struct {
	mock.Mock
}

func (m *MockEbookManager) CreateEbook(ctx context.Context, eb *ebook.Ebook) (*ebook.Ebook, error) {
	args := m.Called(ctx, eb)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ebook.Ebook), args.Error(1)
}
