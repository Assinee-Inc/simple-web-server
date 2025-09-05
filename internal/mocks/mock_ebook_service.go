package mocks

import (
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/stretchr/testify/mock"
)

type MockEbookService struct {
	mock.Mock
}

func (m *MockEbookService) FindByID(id uint) (*models.Ebook, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ebook), args.Error(1)
}

func (m *MockEbookService) FindBySlug(slug string) (*models.Ebook, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ebook), args.Error(1)
}

func (m *MockEbookService) ListEbooksForUser(userID uint, query repository.EbookQuery) (*[]models.Ebook, error) {
	args := m.Called(userID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.Ebook), args.Error(1)
}

func (m *MockEbookService) Update(ebook *models.Ebook) error {
	args := m.Called(ebook)
	return args.Error(0)
}

func (m *MockEbookService) Create(ebook *models.Ebook) error {
	args := m.Called(ebook)
	return args.Error(0)
}

func (m *MockEbookService) CreateEbook(input interface{}) (*models.Ebook, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ebook), args.Error(1)
}

func (m *MockEbookService) UpdateEbook(ebook *models.Ebook) error {
	args := m.Called(ebook)
	return args.Error(0)
}

func (m *MockEbookService) DeleteEbook(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockEbookService) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}
