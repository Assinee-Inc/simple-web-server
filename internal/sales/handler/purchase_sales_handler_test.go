package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	librarymodel "github.com/Assinee-Inc/docffy/internal/library/model"
	"github.com/Assinee-Inc/docffy/internal/mocks"
	salesmodel "github.com/Assinee-Inc/docffy/internal/sales/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPurchaseSalesHandler_PurchaseSalesList_SessionError(t *testing.T) {
	mockPurchaseService := &mocks.MockPurchaseService{}
	mockSessionService := &mocks.MockSessionService{}
	mockCreatorService := &mocks.MockCreatorService{}
	mockEbookService := &mocks.MockEbookService{}
	mockResendService := &mocks.MockResendDownloadLinkService{}
	mockTransactionService := &mocks.MockTransactionService{}

	handler := NewPurchaseSalesHandler(
		mockPurchaseService,
		mockSessionService,
		mockCreatorService,
		mockEbookService,
		mockResendService,
		mockTransactionService,
	)

	mockSessionService.On("GetUserEmailFromSession", mock.AnythingOfType("*http.Request")).Return("", assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/purchase/sales", nil)
	w := httptest.NewRecorder()

	handler.PurchaseSalesList(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestPurchaseSalesHandler_PurchaseSalesList_SuccessJSON(t *testing.T) {
	mockPurchaseService := &mocks.MockPurchaseService{}
	mockSessionService := &mocks.MockSessionService{}
	mockCreatorService := &mocks.MockCreatorService{}
	mockEbookService := &mocks.MockEbookService{}
	mockResendService := &mocks.MockResendDownloadLinkService{}
	mockTransactionService := &mocks.MockTransactionService{}

	handler := NewPurchaseSalesHandler(
		mockPurchaseService,
		mockSessionService,
		mockCreatorService,
		mockEbookService,
		mockResendService,
		mockTransactionService,
	)

	creator := &accountmodel.Creator{Email: "creator@test.com"}
	creator.ID = 1
	mockCreatorService.On("FindCreatorByEmail", "creator@test.com").Return(creator, nil)
	mockPurchaseService.On("GetPurchasesByCreatorIDWithFilters", uint(1), (*uint)(nil), "", "", 1, 10).Return([]salesmodel.Purchase{}, int64(0), nil)
	mockEbookService.On("GetEbooksByCreatorID", uint(1)).Return([]*librarymodel.Ebook{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/purchase/sales", nil)
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, "creator@test.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.PurchaseSalesList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	var body map[string]any
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, true, body["success"])
}

func TestPurchaseSalesHandler_BlockDownload_InvalidMethod(t *testing.T) {
	handler := NewPurchaseSalesHandler(
		&mocks.MockPurchaseService{},
		&mocks.MockSessionService{},
		&mocks.MockCreatorService{},
		&mocks.MockEbookService{},
		&mocks.MockResendDownloadLinkService{},
		&mocks.MockTransactionService{},
	)

	req := httptest.NewRequest(http.MethodGet, "/purchase/sales/block-download", nil)
	w := httptest.NewRecorder()

	handler.BlockDownload(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestPurchaseSalesHandler_BlockDownload_EmptyPurchaseID(t *testing.T) {
	handler := NewPurchaseSalesHandler(
		&mocks.MockPurchaseService{},
		&mocks.MockSessionService{},
		&mocks.MockCreatorService{},
		&mocks.MockEbookService{},
		&mocks.MockResendDownloadLinkService{},
		&mocks.MockTransactionService{},
	)

	form := url.Values{}
	form.Add("purchase_id", "")
	req := httptest.NewRequest(http.MethodPost, "/purchase/sales/block-download", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.BlockDownload(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPurchaseSalesHandler_UnblockDownload_InvalidMethod(t *testing.T) {
	handler := NewPurchaseSalesHandler(
		&mocks.MockPurchaseService{},
		&mocks.MockSessionService{},
		&mocks.MockCreatorService{},
		&mocks.MockEbookService{},
		&mocks.MockResendDownloadLinkService{},
		&mocks.MockTransactionService{},
	)

	req := httptest.NewRequest(http.MethodGet, "/purchase/sales/unblock-download", nil)
	w := httptest.NewRecorder()

	handler.UnblockDownload(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestPurchaseSalesHandler_ResendDownloadLink_InvalidMethod(t *testing.T) {
	handler := NewPurchaseSalesHandler(
		&mocks.MockPurchaseService{},
		&mocks.MockSessionService{},
		&mocks.MockCreatorService{},
		&mocks.MockEbookService{},
		&mocks.MockResendDownloadLinkService{},
		&mocks.MockTransactionService{},
	)

	req := httptest.NewRequest(http.MethodGet, "/purchase/sales/resend-link", nil)
	w := httptest.NewRecorder()

	handler.ResendDownloadLink(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestPurchaseSalesHandler_ResendDownloadLink_EmptyPurchaseID(t *testing.T) {
	handler := NewPurchaseSalesHandler(
		&mocks.MockPurchaseService{},
		&mocks.MockSessionService{},
		&mocks.MockCreatorService{},
		&mocks.MockEbookService{},
		&mocks.MockResendDownloadLinkService{},
		&mocks.MockTransactionService{},
	)

	form := url.Values{}
	form.Add("purchase_id", "")
	req := httptest.NewRequest(http.MethodPost, "/purchase/sales/resend-link", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.ResendDownloadLink(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPurchaseSalesHandler_Integration_Basic(t *testing.T) {
	mockPurchaseService := &mocks.MockPurchaseService{}
	mockSessionService := &mocks.MockSessionService{}
	mockCreatorService := &mocks.MockCreatorService{}
	mockEbookService := &mocks.MockEbookService{}
	mockResendService := &mocks.MockResendDownloadLinkService{}
	mockTransactionService := &mocks.MockTransactionService{}

	handler := NewPurchaseSalesHandler(
		mockPurchaseService,
		mockSessionService,
		mockCreatorService,
		mockEbookService,
		mockResendService,
		mockTransactionService,
	)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.purchaseService)
	assert.NotNil(t, handler.creatorService)
	assert.NotNil(t, handler.ebookService)
	assert.NotNil(t, handler.resendDownloadLinkService)
}
