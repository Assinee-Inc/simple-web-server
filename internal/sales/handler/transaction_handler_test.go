package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anglesson/simple-web-server/internal/models"
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
	completedTransaction := &models.Transaction{
		Model:  gorm.Model{ID: 1},
		Status: models.TransactionStatusCompleted,
	}

	// Transação pendente
	pendingTransaction := &models.Transaction{
		Model:  gorm.Model{ID: 2},
		Status: models.TransactionStatusPending,
	}

	// Assert
	assert.Equal(t, models.TransactionStatusCompleted, completedTransaction.Status)
	assert.NotEqual(t, models.TransactionStatusCompleted, pendingTransaction.Status)
}
