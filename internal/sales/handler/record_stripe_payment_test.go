package handler

import (
	"errors"
	"testing"
	"time"

	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	"github.com/anglesson/simple-web-server/internal/mocks"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func buildHandlerForRecordStripePayment(
	txSvc *mocks.MockTransactionService,
) *CheckoutHandler {
	return &CheckoutHandler{
		transactionService: txSvc,
	}
}

func newCompletedTransaction(id uint, paymentIntentID string) *salesmodel.Transaction {
	return &salesmodel.Transaction{
		Model:                 gorm.Model{ID: id},
		Status:                salesmodel.TransactionStatusCompleted,
		StripePaymentIntentID: paymentIntentID,
	}
}

func newPendingTransaction(id uint) *salesmodel.Transaction {
	return &salesmodel.Transaction{
		Model:  gorm.Model{ID: id},
		Status: salesmodel.TransactionStatusPending,
	}
}

// TestRecordStripePayment_PendingTransaction_UpdatesToCOMpleted verifica o fluxo normal:
// transação pendente → atualizada para completed com o payment intent.
func TestRecordStripePayment_PendingTransaction_UpdatesToCompleted(t *testing.T) {
	tx := newPendingTransaction(1)
	mockTx := new(mocks.MockTransactionService)
	mockTx.On("FindTransactionByPurchaseID", uint(1)).Return(tx, nil)
	mockTx.On("UpdateTransactionToCompleted", uint(1), "pi_new_intent").Return(nil)

	ebook := &librarymodel.Ebook{Value: 50}
	handler := buildHandlerForRecordStripePayment(mockTx)

	handler.recordStripePayment(1, 10, ebook, "pi_new_intent")

	mockTx.AssertCalled(t, "UpdateTransactionToCompleted", uint(1), "pi_new_intent")
	mockTx.AssertNotCalled(t, "CreateDirectTransaction", mock.Anything)
	mockTx.AssertExpectations(t)
}

// TestRecordStripePayment_NoTransaction_UpdatesToCOMpleted verifica que quando não há
// transação prévia, UpdateTransactionToCompleted é chamado (retornará erro que é logado).
func TestRecordStripePayment_NoTransaction_CallsUpdate(t *testing.T) {
	mockTx := new(mocks.MockTransactionService)
	mockTx.On("FindTransactionByPurchaseID", uint(2)).Return((*salesmodel.Transaction)(nil), errors.New("not found"))
	mockTx.On("UpdateTransactionToCompleted", uint(2), "pi_fresh").Return(nil)

	ebook := &librarymodel.Ebook{Value: 100}
	handler := buildHandlerForRecordStripePayment(mockTx)

	handler.recordStripePayment(2, 20, ebook, "pi_fresh")

	mockTx.AssertCalled(t, "UpdateTransactionToCompleted", uint(2), "pi_fresh")
	mockTx.AssertNotCalled(t, "CreateDirectTransaction", mock.Anything)
	mockTx.AssertExpectations(t)
}

// TestRecordStripePayment_CompletedWithDifferentIntent_CreatesNewTransaction é o teste
// que previne a regressão do bug: Stripe payment not recorded when transaction already completed.
func TestRecordStripePayment_CompletedWithDifferentIntent_CreatesNewTransaction(t *testing.T) {
	existingTx := newCompletedTransaction(1, "pi_old_intent")
	mockTx := new(mocks.MockTransactionService)
	mockTx.On("FindTransactionByPurchaseID", uint(1)).Return(existingTx, nil)
	mockTx.On("CreateDirectTransaction", mock.MatchedBy(func(tx *salesmodel.Transaction) bool {
		return tx.StripePaymentIntentID == "pi_new_different_intent" &&
			tx.Status == salesmodel.TransactionStatusCompleted &&
			tx.PurchaseID == 1 &&
			tx.CreatorID == 10 &&
			tx.ProcessedAt != nil
	})).Return(nil)

	ebook := &librarymodel.Ebook{Value: 50}
	handler := buildHandlerForRecordStripePayment(mockTx)

	handler.recordStripePayment(1, 10, ebook, "pi_new_different_intent")

	mockTx.AssertCalled(t, "CreateDirectTransaction", mock.Anything)
	mockTx.AssertNotCalled(t, "UpdateTransactionToCompleted", mock.Anything, mock.Anything)
	mockTx.AssertExpectations(t)
}

// TestRecordStripePayment_CompletedWithSameIntent_UpdatesToCOMpleted verifica idempotência:
// mesmo payment intent → segue o fluxo de update (que não fará nada, pois já está completed).
func TestRecordStripePayment_CompletedWithSameIntent_CallsUpdate(t *testing.T) {
	existingTx := newCompletedTransaction(5, "pi_same_intent")
	mockTx := new(mocks.MockTransactionService)
	mockTx.On("FindTransactionByPurchaseID", uint(5)).Return(existingTx, nil)
	mockTx.On("UpdateTransactionToCompleted", uint(5), "pi_same_intent").Return(nil)

	ebook := &librarymodel.Ebook{Value: 30}
	handler := buildHandlerForRecordStripePayment(mockTx)

	handler.recordStripePayment(5, 20, ebook, "pi_same_intent")

	mockTx.AssertCalled(t, "UpdateTransactionToCompleted", uint(5), "pi_same_intent")
	mockTx.AssertNotCalled(t, "CreateDirectTransaction", mock.Anything)
	mockTx.AssertExpectations(t)
}

// TestRecordStripePayment_CompletedWithDifferentIntent_CreateFails_LogsError verifica que
// erros na criação da nova transação são tratados sem panic.
func TestRecordStripePayment_CompletedWithDifferentIntent_CreateFails(t *testing.T) {
	existingTx := newCompletedTransaction(7, "pi_original")
	mockTx := new(mocks.MockTransactionService)
	mockTx.On("FindTransactionByPurchaseID", uint(7)).Return(existingTx, nil)
	mockTx.On("CreateDirectTransaction", mock.Anything).Return(errors.New("db error"))

	ebook := &librarymodel.Ebook{Value: 75}
	handler := buildHandlerForRecordStripePayment(mockTx)

	// Não deve entrar em panic
	assert.NotPanics(t, func() {
		handler.recordStripePayment(7, 30, ebook, "pi_another_payment")
	})

	mockTx.AssertExpectations(t)
}

// TestRecordStripePayment_UpdateFails_LogsError verifica que erros no update
// não causam panic.
func TestRecordStripePayment_UpdateFails_LogsError(t *testing.T) {
	tx := newPendingTransaction(9)
	mockTx := new(mocks.MockTransactionService)
	mockTx.On("FindTransactionByPurchaseID", uint(9)).Return(tx, nil)
	mockTx.On("UpdateTransactionToCompleted", uint(9), "pi_x").Return(errors.New("update error"))

	ebook := &librarymodel.Ebook{Value: 20}
	handler := buildHandlerForRecordStripePayment(mockTx)

	assert.NotPanics(t, func() {
		handler.recordStripePayment(9, 40, ebook, "pi_x")
	})

	mockTx.AssertExpectations(t)
}

// TestRecordStripePayment_NewTransactionHasCorrectFields verifica que a nova transação
// criada para pagamento duplicado tem todos os campos essenciais preenchidos.
func TestRecordStripePayment_NewTransactionHasCorrectFields(t *testing.T) {
	existingTx := newCompletedTransaction(3, "pi_first_payment")
	var capturedTx *salesmodel.Transaction

	mockTx := new(mocks.MockTransactionService)
	mockTx.On("FindTransactionByPurchaseID", uint(3)).Return(existingTx, nil)
	mockTx.On("CreateDirectTransaction", mock.Anything).Run(func(args mock.Arguments) {
		capturedTx = args.Get(0).(*salesmodel.Transaction)
	}).Return(nil)

	ebook := &librarymodel.Ebook{Value: 100, PromotionalValue: 80}
	handler := buildHandlerForRecordStripePayment(mockTx)
	beforeCall := time.Now()

	handler.recordStripePayment(3, 50, ebook, "pi_second_payment")

	assert.NotNil(t, capturedTx)
	assert.Equal(t, "pi_second_payment", capturedTx.StripePaymentIntentID)
	assert.Equal(t, salesmodel.TransactionStatusCompleted, capturedTx.Status)
	assert.Equal(t, uint(3), capturedTx.PurchaseID)
	assert.Equal(t, uint(50), capturedTx.CreatorID)
	assert.NotNil(t, capturedTx.ProcessedAt)
	assert.True(t, capturedTx.ProcessedAt.After(beforeCall) || capturedTx.ProcessedAt.Equal(beforeCall))
	assert.Greater(t, capturedTx.TotalAmount, int64(0))
}
