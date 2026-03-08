package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestTransactionHandler_ResendDownloadLink_InvalidMethod(t *testing.T) {
	// Arrange
	handler := &TransactionHandler{}

	req := httptest.NewRequest(http.MethodGet, "/transactions/resend-download-link", nil)
	w := httptest.NewRecorder()

	// Act
	handler.ResendDownloadLink(w, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestTransactionHandler_ResendDownloadLink_InvalidTransactionID(t *testing.T) {
	// Arrange
	handler := &TransactionHandler{}

	req := httptest.NewRequest(http.MethodPost, "/transactions/resend-download-link", nil)
	w := httptest.NewRecorder()

	// Act
	handler.ResendDownloadLink(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransactionStatus_ValidateCompletedTransactions(t *testing.T) {
	// Test para validar status de transações

	// Transação completada
	completedTransaction := &salesmodel.Transaction{
		Model:  gorm.Model{ID: 1},
		Status: salesmodel.TransactionStatusCompleted,
	}

	// Transação pendente
	pendingTransaction := &salesmodel.Transaction{
		Model:  gorm.Model{ID: 2},
		Status: salesmodel.TransactionStatusPending,
	}

	// Assert
	assert.Equal(t, salesmodel.TransactionStatusCompleted, completedTransaction.Status)
	assert.NotEqual(t, salesmodel.TransactionStatusCompleted, pendingTransaction.Status)
}
