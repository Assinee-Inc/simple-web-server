package handler

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anglesson/simple-web-server/internal/mocks"
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDownloadService é um mock do DownloadService para testes.
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

	mockTemplateRenderer := new(mocks.MockTemplateRenderer)
	mockTemplateRenderer.On("ViewWithoutLayout", w, req, "ebook/download-limit-exceeded", mock.AnythingOfType("map[string]interface {}")).Return()

	mockDownloadService := new(MockDownloadService)
	handler := NewDownloadHandler(mockDownloadService, mockTemplateRenderer)

	handler.showLimitExceededPage(w, req, purchase)

	mockTemplateRenderer.AssertExpectations(t)
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

	mockTemplateRenderer := new(mocks.MockTemplateRenderer)
	mockTemplateRenderer.On("ViewWithoutLayout", w, req, "ebook/download-expired", mock.AnythingOfType("map[string]interface {}")).Return()

	mockDownloadService := new(MockDownloadService)
	handler := NewDownloadHandler(mockDownloadService, mockTemplateRenderer)

	handler.showExpiredDownloadPage(w, req, purchase)

	mockTemplateRenderer.AssertExpectations(t)
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
