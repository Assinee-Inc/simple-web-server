package mocks

import (
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	"github.com/stretchr/testify/mock"
)

// MockTransactionRepository é um mock para o repository de transações
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) CreateTransaction(transaction *salesmodel.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) UpdateTransaction(transaction *salesmodel.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) FindByID(id uint) (*salesmodel.Transaction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*salesmodel.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) FindByPublicID(publicID string) (*salesmodel.Transaction, error) {
	args := m.Called(publicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*salesmodel.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) FindByCreatorID(creatorID uint, page, limit int) ([]*salesmodel.Transaction, int64, error) {
	args := m.Called(creatorID, page, limit)
	return args.Get(0).([]*salesmodel.Transaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionRepository) FindByCreatorIDWithFilters(creatorID uint, page, limit int, search, status string) ([]*salesmodel.Transaction, int64, error) {
	args := m.Called(creatorID, page, limit, search, status)
	return args.Get(0).([]*salesmodel.Transaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionRepository) FindByPurchaseID(purchaseID uint) (*salesmodel.Transaction, error) {
	args := m.Called(purchaseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*salesmodel.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) UpdateTransactionStatus(id uint, status salesmodel.TransactionStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}
