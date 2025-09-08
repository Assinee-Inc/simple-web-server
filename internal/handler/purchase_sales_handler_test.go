package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPurchaseSalesHandler_PurchaseSalesList_SessionError(t *testing.T) {
	// Arrange
	mockTemplateRenderer := &mocks.MockTemplateRenderer{}
	mockPurchaseService := &mocks.MockPurchaseService{}
	mockSessionService := &mocks.MockSessionService{}
	mockCreatorService := &mocks.MockCreatorService{}
	mockEbookService := &mocks.MockEbookService{}
	mockResendService := &mocks.MockResendDownloadLinkService{}
	mockTransactionService := &mocks.MockTransactionService{}

	handler := NewPurchaseSalesHandler(
		mockTemplateRenderer,
		mockPurchaseService,
		mockSessionService,
		mockCreatorService,
		mockEbookService,
		mockResendService,
		mockTransactionService,
	)

	// Setup expectation for session error
	mockSessionService.On("GetUserEmailFromSession", mock.AnythingOfType("*http.Request")).Return("", assert.AnError)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/purchase/sales", nil)
	w := httptest.NewRecorder()

	// Act
	handler.PurchaseSalesList(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPurchaseSalesHandler_BlockDownload_InvalidMethod(t *testing.T) {
	// Arrange
	mockTemplateRenderer := &mocks.MockTemplateRenderer{}
	mockPurchaseService := &mocks.MockPurchaseService{}
	mockSessionService := &mocks.MockSessionService{}
	mockCreatorService := &mocks.MockCreatorService{}
	mockEbookService := &mocks.MockEbookService{}
	mockResendService := &mocks.MockResendDownloadLinkService{}
	mockTransactionService := &mocks.MockTransactionService{}

	handler := NewPurchaseSalesHandler(
		mockTemplateRenderer,
		mockPurchaseService,
		mockSessionService,
		mockCreatorService,
		mockEbookService,
		mockResendService,
		mockTransactionService,
	)

	// Create GET request (should be POST)
	req := httptest.NewRequest(http.MethodGet, "/purchase/sales/block-download", nil)
	w := httptest.NewRecorder()

	// Act
	handler.BlockDownload(w, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestPurchaseSalesHandler_BlockDownload_InvalidPurchaseID(t *testing.T) {
	// Arrange
	mockTemplateRenderer := &mocks.MockTemplateRenderer{}
	mockPurchaseService := &mocks.MockPurchaseService{}
	mockSessionService := &mocks.MockSessionService{}
	mockCreatorService := &mocks.MockCreatorService{}
	mockEbookService := &mocks.MockEbookService{}
	mockResendService := &mocks.MockResendDownloadLinkService{}
	mockTransactionService := &mocks.MockTransactionService{}

	handler := NewPurchaseSalesHandler(
		mockTemplateRenderer,
		mockPurchaseService,
		mockSessionService,
		mockCreatorService,
		mockEbookService,
		mockResendService,
		mockTransactionService,
	)

	// Create form with invalid purchase ID
	form := url.Values{}
	form.Add("purchase_id", "invalid")
	req := httptest.NewRequest(http.MethodPost, "/purchase/sales/block-download", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Act
	handler.BlockDownload(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPurchaseSalesHandler_UnblockDownload_InvalidMethod(t *testing.T) {
	// Arrange
	mockTemplateRenderer := &mocks.MockTemplateRenderer{}
	mockPurchaseService := &mocks.MockPurchaseService{}
	mockSessionService := &mocks.MockSessionService{}
	mockCreatorService := &mocks.MockCreatorService{}
	mockEbookService := &mocks.MockEbookService{}
	mockResendService := &mocks.MockResendDownloadLinkService{}
	mockTransactionService := &mocks.MockTransactionService{}

	handler := NewPurchaseSalesHandler(
		mockTemplateRenderer,
		mockPurchaseService,
		mockSessionService,
		mockCreatorService,
		mockEbookService,
		mockResendService,
		mockTransactionService,
	)

	// Create GET request (should be POST)
	req := httptest.NewRequest(http.MethodGet, "/purchase/sales/unblock-download", nil)
	w := httptest.NewRecorder()

	// Act
	handler.UnblockDownload(w, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestPurchaseSalesHandler_ResendDownloadLink_InvalidMethod(t *testing.T) {
	// Arrange
	mockTemplateRenderer := &mocks.MockTemplateRenderer{}
	mockPurchaseService := &mocks.MockPurchaseService{}
	mockSessionService := &mocks.MockSessionService{}
	mockCreatorService := &mocks.MockCreatorService{}
	mockEbookService := &mocks.MockEbookService{}
	mockResendService := &mocks.MockResendDownloadLinkService{}
	mockTransactionService := &mocks.MockTransactionService{}

	handler := NewPurchaseSalesHandler(
		mockTemplateRenderer,
		mockPurchaseService,
		mockSessionService,
		mockCreatorService,
		mockEbookService,
		mockResendService,
		mockTransactionService,
	)

	// Create GET request (should be POST)
	req := httptest.NewRequest(http.MethodGet, "/purchase/sales/resend-link", nil)
	w := httptest.NewRecorder()

	// Act
	handler.ResendDownloadLink(w, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestPurchaseSalesHandler_ResendDownloadLink_InvalidPurchaseID(t *testing.T) {
	// Arrange
	mockTemplateRenderer := &mocks.MockTemplateRenderer{}
	mockPurchaseService := &mocks.MockPurchaseService{}
	mockSessionService := &mocks.MockSessionService{}
	mockCreatorService := &mocks.MockCreatorService{}
	mockEbookService := &mocks.MockEbookService{}
	mockResendService := &mocks.MockResendDownloadLinkService{}
	mockTransactionService := &mocks.MockTransactionService{}

	handler := NewPurchaseSalesHandler(
		mockTemplateRenderer,
		mockPurchaseService,
		mockSessionService,
		mockCreatorService,
		mockEbookService,
		mockResendService,
		mockTransactionService,
	)

	// Create form with invalid purchase ID
	form := url.Values{}
	form.Add("purchase_id", "invalid")
	req := httptest.NewRequest(http.MethodPost, "/purchase/sales/resend-link", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Act
	handler.ResendDownloadLink(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Teste de integração básica para verificar se o fluxo funciona
func TestPurchaseSalesHandler_Integration_Basic(t *testing.T) {
	// Este é um teste básico que verifica se a estrutura está correta
	// Em um ambiente real, seria necessário configurar um banco de dados de teste

	// Arrange
	mockTemplateRenderer := &mocks.MockTemplateRenderer{}
	mockPurchaseService := &mocks.MockPurchaseService{}
	mockSessionService := &mocks.MockSessionService{}
	mockCreatorService := &mocks.MockCreatorService{}
	mockEbookService := &mocks.MockEbookService{}
	mockResendService := &mocks.MockResendDownloadLinkService{}
	mockTransactionService := &mocks.MockTransactionService{}

	// Verificar se o handler pode ser criado sem erros
	handler := NewPurchaseSalesHandler(
		mockTemplateRenderer,
		mockPurchaseService,
		mockSessionService,
		mockCreatorService,
		mockEbookService,
		mockResendService,
		mockTransactionService,
	)

	// Assert
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.templateRenderer)
	assert.NotNil(t, handler.purchaseService)
	assert.NotNil(t, handler.sessionService)
	assert.NotNil(t, handler.creatorService)
	assert.NotNil(t, handler.ebookService)
	assert.NotNil(t, handler.resendDownloadLinkService)
}
