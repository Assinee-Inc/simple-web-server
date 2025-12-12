package mocks

import (
	"github.com/anglesson/simple-web-server/internal/library/models"
	"github.com/stretchr/testify/mock"
)

type MockEbookRepository struct {
	mock.Mock
}

func (m *MockEbookRepository) Save(ebook *models.Ebook) error {
	args := m.Called(ebook)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}

func (m *MockEbookRepository) FindByParams(params ...any) ([]*models.Ebook, error) {
	args := m.Called(params...)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Ebook), args.Error(1)
}
