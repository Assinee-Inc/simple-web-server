package mocks

import (
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockTransactionRepository é um mock para o repository de transações
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) CreateTransaction(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) UpdateTransaction(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) FindByID(id uint) (*models.Transaction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) FindByCreatorID(creatorID uint, page, limit int) ([]*models.Transaction, int64, error) {
	args := m.Called(creatorID, page, limit)
	return args.Get(0).([]*models.Transaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionRepository) FindByPurchaseID(purchaseID uint) (*models.Transaction, error) {
	args := m.Called(purchaseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) UpdateTransactionStatus(id uint, status models.TransactionStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}
