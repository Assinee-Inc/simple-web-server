package service

import (
	"testing"
	"time"

	"github.com/anglesson/simple-web-server/internal/mocks"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindTransactionByPurchaseID(t *testing.T) {
	mockRepo := new(mocks.MockTransactionRepository)
	service := &transactionServiceImpl{
		transactionRepo: mockRepo,
	}

	// Arrange
	purchaseID := uint(123)
	expectedTransaction := &salesmodel.Transaction{
		PurchaseID: purchaseID,
		Status:     salesmodel.TransactionStatusPending,
	}
	expectedTransaction.ID = 1

	mockRepo.On("FindByPurchaseID", purchaseID).Return(expectedTransaction, nil)

	// Act
	result, err := service.FindTransactionByPurchaseID(purchaseID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedTransaction.ID, result.ID)
	assert.Equal(t, purchaseID, result.PurchaseID)
	mockRepo.AssertExpectations(t)
}

func TestUpdateTransactionToCompleted_Success(t *testing.T) {
	mockRepo := new(mocks.MockTransactionRepository)
	service := &transactionServiceImpl{
		transactionRepo: mockRepo,
	}

	// Arrange
	purchaseID := uint(123)
	stripePaymentIntentID := "pi_test_123"

	existingTransaction := &salesmodel.Transaction{
		PurchaseID: purchaseID,
		Status:     salesmodel.TransactionStatusPending,
	}
	existingTransaction.ID = 1

	mockRepo.On("FindByPurchaseID", purchaseID).Return(existingTransaction, nil)
	mockRepo.On("UpdateTransaction", mock.MatchedBy(func(t *salesmodel.Transaction) bool {
		return t.ID == 1 &&
			t.Status == salesmodel.TransactionStatusCompleted &&
			t.StripePaymentIntentID == stripePaymentIntentID &&
			t.ProcessedAt != nil
	})).Return(nil)

	// Act
	err := service.UpdateTransactionToCompleted(purchaseID, stripePaymentIntentID)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateTransactionToCompleted_TransactionNotFound(t *testing.T) {
	mockRepo := new(mocks.MockTransactionRepository)
	service := &transactionServiceImpl{
		transactionRepo: mockRepo,
	}

	// Arrange
	purchaseID := uint(123)
	stripePaymentIntentID := "pi_test_123"

	mockRepo.On("FindByPurchaseID", purchaseID).Return(nil, nil)

	// Act
	err := service.UpdateTransactionToCompleted(purchaseID, stripePaymentIntentID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transação não encontrada")
	mockRepo.AssertExpectations(t)
}

func TestUpdateTransactionToCompleted_AlreadyCompleted(t *testing.T) {
	mockRepo := new(mocks.MockTransactionRepository)
	service := &transactionServiceImpl{
		transactionRepo: mockRepo,
	}

	// Arrange
	purchaseID := uint(123)
	stripePaymentIntentID := "pi_test_123"

	completedTransaction := &salesmodel.Transaction{
		PurchaseID:            purchaseID,
		Status:                salesmodel.TransactionStatusCompleted,
		StripePaymentIntentID: "pi_old_123",
		ProcessedAt:           &time.Time{},
	}
	completedTransaction.ID = 1

	mockRepo.On("FindByPurchaseID", purchaseID).Return(completedTransaction, nil)

	// Act
	err := service.UpdateTransactionToCompleted(purchaseID, stripePaymentIntentID)

	// Assert
	assert.NoError(t, err) // Should not error, just skip update
	mockRepo.AssertExpectations(t)
	// UpdateTransaction should not be called because transaction is already completed
	mockRepo.AssertNotCalled(t, "UpdateTransaction", mock.Anything)
}
