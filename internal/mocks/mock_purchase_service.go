package mocks

import (
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockPurchaseService is a mock implementation of a PurchaseService.

type MockPurchaseService struct {
	mock.Mock
}

func (m *MockPurchaseService) CreatePurchase(ebookId uint, clients []uint) error {
	args := m.Called(ebookId, clients)
	return args.Error(0)
}

func (m *MockPurchaseService) CreatePurchaseWithResult(ebookId uint, clientId uint) (*models.Purchase, error) {
	args := m.Called(ebookId, clientId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Purchase), args.Error(1)
}

func (m *MockPurchaseService) GetPurchasesByCreatorIDWithFilters(creatorID uint, ebookID *uint, clientName, clientEmail string, page, limit int) ([]models.Purchase, int64, error) {
	args := m.Called(creatorID, ebookID, clientName, clientEmail, page, limit)
	return args.Get(0).([]models.Purchase), args.Get(1).(int64), args.Error(2)
}

func (m *MockPurchaseService) BlockDownload(purchaseID uint, creatorID uint, block bool) error {
	args := m.Called(purchaseID, creatorID, block)
	return args.Error(0)
}

func (m *MockPurchaseService) GetPurchaseByID(id uint) (*models.Purchase, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Purchase), args.Error(1)
}

func (m *MockPurchaseService) GetEbookFile(purchaseID int, fileID uint) (string, error) {
	args := m.Called(purchaseID, fileID)
	return args.String(0), args.Error(1)
}

func (m *MockPurchaseService) GetEbookFiles(purchaseID int) ([]*models.File, error) {
	args := m.Called(purchaseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.File), args.Error(1)
}
