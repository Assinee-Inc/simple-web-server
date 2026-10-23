package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	"github.com/Assinee-Inc/docffy/internal/mocks"
	salesmodel "github.com/Assinee-Inc/docffy/internal/sales/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestTransactionHandler_ResendDownloadLink_InvalidMethod(t *testing.T) {
	handler := &TransactionHandler{}

	req := httptest.NewRequest(http.MethodGet, "/transactions/resend-download-link", nil)
	w := httptest.NewRecorder()

	handler.ResendDownloadLink(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestTransactionHandler_ResendDownloadLink_InvalidTransactionID(t *testing.T) {
	handler := &TransactionHandler{}

	req := httptest.NewRequest(http.MethodPost, "/transactions/resend-download-link", nil)
	w := httptest.NewRecorder()

	handler.ResendDownloadLink(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestTransactionHandler_TransactionDetail_SuccessJSON(t *testing.T) {
	mockTxSvc := &mocks.MockTransactionService{}
	mockCreatorSvc := &mocks.MockCreatorService{}
	mockResendSvc := &mocks.MockResendDownloadLinkService{}

	creator := &accountmodel.Creator{Email: "creator@test.com"}
	creator.ID = 1
	tx := &salesmodel.Transaction{
		Model:     gorm.Model{ID: 5},
		CreatorID: 1,
		Status:    salesmodel.TransactionStatusCompleted,
	}
	tx.PublicID = "tx_123"

	mockCreatorSvc.On("FindCreatorByEmail", "creator@test.com").Return(creator, nil)
	mockTxSvc.On("GetTransactionByPublicID", "tx_123").Return(tx, nil)

	h := NewTransactionHandler(mockTxSvc, nil, mockCreatorSvc, mockResendSvc)

	req := httptest.NewRequest(http.MethodGet, "/transactions/detail?id=tx_123", nil)
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, "creator@test.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.TransactionDetail(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	var body map[string]any
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, true, body["success"])
}

func TestTransactionStatus_ValidateCompletedTransactions(t *testing.T) {
	completedTransaction := &salesmodel.Transaction{
		Model:  gorm.Model{ID: 1},
		Status: salesmodel.TransactionStatusCompleted,
	}

	pendingTransaction := &salesmodel.Transaction{
		Model:  gorm.Model{ID: 2},
		Status: salesmodel.TransactionStatusPending,
	}

	assert.Equal(t, salesmodel.TransactionStatusCompleted, completedTransaction.Status)
	assert.NotEqual(t, salesmodel.TransactionStatusCompleted, pendingTransaction.Status)
}
