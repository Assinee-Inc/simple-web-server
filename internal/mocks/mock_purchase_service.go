package mocks

import (
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
