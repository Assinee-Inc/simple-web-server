package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	librarymodel "github.com/Assinee-Inc/docffy/internal/library/model"
	salesmodel "github.com/Assinee-Inc/docffy/internal/sales/model"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockDownloadService struct {
	mock.Mock
}

func (m *MockDownloadService) FindPurchaseByHash(hashID string) (*salesmodel.Purchase, error) {
	args := m.Called(hashID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*salesmodel.Purchase), args.Error(1)
}

func (m *MockDownloadService) GetEbookFile(hashID string, filePublicID string) (string, error) {
	args := m.Called(hashID, filePublicID)
	return args.String(0), args.Error(1)
}

func (m *MockDownloadService) GetEbookFiles(purchaseID int) ([]*librarymodel.File, error) {
	args := m.Called(purchaseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*librarymodel.File), args.Error(1)
}

func TestShowLimitExceededPage(t *testing.T) {
	purchase := &salesmodel.Purchase{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
		},
		EbookID:       1,
		ClientID:      1,
		DownloadsUsed: 5,
		DownloadLimit: 5,
		Ebook: librarymodel.Ebook{
			Title: "Test Ebook",
		},
		Client: salesmodel.Client{
			Name:  "Test Client",
			Email: "client@test.com",
		},
	}

	req := httptest.NewRequest("GET", "/purchase/download/1", nil)
	w := httptest.NewRecorder()

	mockDownloadService := new(MockDownloadService)
	handler := NewDownloadHandler(mockDownloadService)

	handler.showLimitExceededPage(w, req, purchase)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var body map[string]any
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, false, body["success"])
	assert.Equal(t, "DOWNLOAD_LIMIT_EXCEEDED", body["code"])
}

func TestShowExpiredDownloadPage(t *testing.T) {
	expiredTime := time.Now().Add(-24 * time.Hour)
	purchase := &salesmodel.Purchase{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
		},
		EbookID:   1,
		ClientID:  1,
		ExpiresAt: expiredTime,
		Ebook: librarymodel.Ebook{
			Title: "Test Ebook",
		},
		Client: salesmodel.Client{
			Name:  "Test Client",
			Email: "client@test.com",
		},
	}

	req := httptest.NewRequest("GET", "/purchase/download/1", nil)
	w := httptest.NewRecorder()

	mockDownloadService := new(MockDownloadService)
	handler := NewDownloadHandler(mockDownloadService)

	handler.showExpiredDownloadPage(w, req, purchase)

	assert.Equal(t, http.StatusGone, w.Code)
	var body map[string]any
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, false, body["success"])
	assert.Equal(t, "DOWNLOAD_EXPIRED", body["code"])
}

func TestShowEbookFilesWithLimitExceeded(t *testing.T) {
	purchase := &salesmodel.Purchase{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
		},
		EbookID:       1,
		ClientID:      1,
		DownloadsUsed: 5,
		DownloadLimit: 5,
		Ebook: librarymodel.Ebook{
			Title: "Test Ebook",
		},
		Client: salesmodel.Client{
			Name:  "Test Client",
			Email: "client@test.com",
		},
	}

	assert.False(t, purchase.AvailableDownloads())
	assert.Equal(t, 5, purchase.DownloadsUsed)
	assert.Equal(t, 5, purchase.DownloadLimit)
}

func TestShowEbookFilesWithExpiredPurchase(t *testing.T) {
	expiredTime := time.Now().Add(-24 * time.Hour)
	purchase := &salesmodel.Purchase{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
		},
		EbookID:   1,
		ClientID:  1,
		ExpiresAt: expiredTime,
		Ebook: librarymodel.Ebook{
			Title: "Test Ebook",
		},
		Client: salesmodel.Client{
			Name:  "Test Client",
			Email: "client@test.com",
		},
	}

	assert.True(t, purchase.IsExpired())
	assert.True(t, purchase.ExpiresAt.Before(time.Now()))
}

func TestShowEbookFilesWithValidPurchase(t *testing.T) {
	futureTime := time.Now().Add(30 * 24 * time.Hour)
	purchase := &salesmodel.Purchase{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
		},
		EbookID:       1,
		ClientID:      1,
		DownloadsUsed: 2,
		DownloadLimit: 5,
		ExpiresAt:     futureTime,
		Ebook: librarymodel.Ebook{
			Title: "Test Ebook",
		},
		Client: salesmodel.Client{
			Name:  "Test Client",
			Email: "client@test.com",
		},
	}

	assert.True(t, purchase.AvailableDownloads())
	assert.False(t, purchase.IsExpired())
	assert.Equal(t, 2, purchase.DownloadsUsed)
	assert.Equal(t, 5, purchase.DownloadLimit)
	assert.True(t, purchase.ExpiresAt.After(time.Now()))
}

func TestShowEbookFiles_PaymentPending_Returns403JSON(t *testing.T) {
	purchase := &salesmodel.Purchase{
		Model:         gorm.Model{ID: 1},
		EbookID:       1,
		ClientID:      1,
		DownloadLimit: -1,
		PaymentStatus: salesmodel.PaymentStatusPending,
		Ebook:         librarymodel.Ebook{Title: "Test Ebook"},
		Client:        salesmodel.Client{Name: "Test Client", Email: "client@test.com"},
	}

	req := httptest.NewRequest("GET", "/purchase/download/abc123", nil)
	w := httptest.NewRecorder()

	mockDownloadService := new(MockDownloadService)
	mockDownloadService.On("FindPurchaseByHash", "abc123").Return(purchase, nil)

	handler := NewDownloadHandler(mockDownloadService)
	handler.showEbookFiles(w, req, "abc123")

	assert.Equal(t, http.StatusForbidden, w.Code)
	var body map[string]any
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, "PAYMENT_PENDING", body["code"])
	mockDownloadService.AssertExpectations(t)
}

func TestShowEbookFiles_PaymentConfirmed_ShowsFiles(t *testing.T) {
	purchase := &salesmodel.Purchase{
		Model:         gorm.Model{ID: 1},
		EbookID:       1,
		ClientID:      1,
		DownloadLimit: -1,
		PaymentStatus: salesmodel.PaymentStatusConfirmed,
		Ebook:         librarymodel.Ebook{Title: "Test Ebook"},
		Client:        salesmodel.Client{Name: "Test Client", Email: "client@test.com"},
	}

	files := []*librarymodel.File{{PublicID: "file_1"}}

	req := httptest.NewRequest("GET", "/purchase/download/abc123", nil)
	w := httptest.NewRecorder()

	mockDownloadService := new(MockDownloadService)
	mockDownloadService.On("FindPurchaseByHash", "abc123").Return(purchase, nil)
	mockDownloadService.On("GetEbookFiles", 1).Return(files, nil)

	handler := NewDownloadHandler(mockDownloadService)
	handler.showEbookFiles(w, req, "abc123")

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, true, body["success"])
	assert.Equal(t, "Test Ebook", body["ebook_title"])
	mockDownloadService.AssertExpectations(t)
}

func TestShowEbookFiles_PaymentFailed_Returns403JSON(t *testing.T) {
	purchase := &salesmodel.Purchase{
		Model:         gorm.Model{ID: 1},
		EbookID:       1,
		ClientID:      1,
		DownloadLimit: -1,
		PaymentStatus: salesmodel.PaymentStatusFailed,
		Ebook:         librarymodel.Ebook{Title: "Test Ebook"},
		Client:        salesmodel.Client{Name: "Test Client", Email: "client@test.com"},
	}

	req := httptest.NewRequest("GET", "/purchase/download/abc123", nil)
	w := httptest.NewRecorder()

	mockDownloadService := new(MockDownloadService)
	mockDownloadService.On("FindPurchaseByHash", "abc123").Return(purchase, nil)

	handler := NewDownloadHandler(mockDownloadService)
	handler.showEbookFiles(w, req, "abc123")

	assert.Equal(t, http.StatusForbidden, w.Code)
	var body map[string]any
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, "PAYMENT_PENDING", body["code"])
	mockDownloadService.AssertExpectations(t)
}

func TestShowEbookFiles_HashNotFound_Returns404(t *testing.T) {
	req := httptest.NewRequest("GET", "/purchase/download/notfound", nil)
	w := httptest.NewRecorder()

	mockDownloadService := new(MockDownloadService)
	mockDownloadService.On("FindPurchaseByHash", "notfound").Return(nil, assert.AnError)

	handler := NewDownloadHandler(mockDownloadService)
	handler.showEbookFiles(w, req, "notfound")

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDownloadService.AssertExpectations(t)
}

func TestPurchaseDownloadHandler_PaymentPending_BlocksFileDownload(t *testing.T) {
	purchase := &salesmodel.Purchase{
		Model:         gorm.Model{ID: 1},
		EbookID:       1,
		ClientID:      1,
		DownloadLimit: -1,
		PaymentStatus: salesmodel.PaymentStatusPending,
		Ebook:         librarymodel.Ebook{Title: "Test Ebook"},
		Client:        salesmodel.Client{Name: "Test Client", Email: "client@test.com"},
	}

	req := httptest.NewRequest("GET", "/purchase/download/abc123?file_id=file_1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("hash_id", "abc123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	mockDownloadService := new(MockDownloadService)
	mockDownloadService.On("FindPurchaseByHash", "abc123").Return(purchase, nil)

	handler := NewDownloadHandler(mockDownloadService)
	handler.PurchaseDownloadHandler(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockDownloadService.AssertExpectations(t)
	mockDownloadService.AssertNotCalled(t, "GetEbookFile", mock.Anything, mock.Anything)
}
