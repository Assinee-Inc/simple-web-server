package mocks

import (
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) CreateTransaction(purchase *models.Purchase, totalAmount int64) (*models.Transaction, error) {
	args := m.Called(purchase, totalAmount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionService) CreateDirectTransaction(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionService) ProcessPaymentWithSplit(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionService) GetTransactionByID(id uint) (*models.Transaction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetTransactionsByCreatorID(creatorID uint, page, limit int) ([]*models.Transaction, int64, error) {
	args := m.Called(creatorID, page, limit)
	return args.Get(0).([]*models.Transaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionService) GetTransactionsByCreatorIDWithFilters(creatorID uint, page, limit int, search, status string) ([]*models.Transaction, int64, error) {
	args := m.Called(creatorID, page, limit, search, status)
	return args.Get(0).([]*models.Transaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionService) FindTransactionByPurchaseID(purchaseID uint) (*models.Transaction, error) {
	args := m.Called(purchaseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionService) UpdateTransactionToCompleted(purchaseID uint, stripePaymentIntentID string) error {
	args := m.Called(purchaseID, stripePaymentIntentID)
	return args.Error(0)
}
