package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Mock para o repository
type MockPurchaseRepository struct {
	mock.Mock
}

func (m *MockPurchaseRepository) FindByID(id uint) (*models.Purchase, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Purchase), args.Error(1)
}

func (m *MockPurchaseRepository) CreateManyPurchases(purchases []*models.Purchase) error {
	args := m.Called(purchases)
	return args.Error(0)
}

func (m *MockPurchaseRepository) Update(purchase *models.Purchase) error {
	args := m.Called(purchase)
	return args.Error(0)
}

func TestShowLimitExceededPage(t *testing.T) {
	// Criar uma compra com limite atingido
	purchase := &models.Purchase{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
		},
		EbookID:       1,
		ClientID:      1,
		DownloadsUsed: 5,
		DownloadLimit: 5,
		Ebook: models.Ebook{
			Title: "Test Ebook",
			Creator: models.Creator{
				Name:  "Test Creator",
				Email: "creator@test.com",
			},
		},
		Client: models.Client{
			Name:  "Test Client",
			Email: "client@test.com",
		},
	}

	// Criar request e response
	req := httptest.NewRequest("GET", "/purchase/download/1", nil)
	w := httptest.NewRecorder()

	// Chamar a função
	showLimitExceededPage(w, req, purchase)

	// Verificar se a resposta foi bem-sucedida
	assert.Equal(t, http.StatusOK, w.Code)

	// Verificar se o conteúdo contém informações sobre limite excedido
	body := w.Body.String()
	assert.Contains(t, body, "Limite de Downloads Atingido")
	assert.Contains(t, body, "Test Ebook")
	assert.Contains(t, body, "Test Creator")
	assert.Contains(t, body, "5") // Downloads realizados
	assert.Contains(t, body, "0") // Downloads disponíveis
}

func TestShowExpiredDownloadPage(t *testing.T) {
	// Criar uma compra expirada
	expiredTime := time.Now().Add(-24 * time.Hour) // Expirada há 1 dia
	purchase := &models.Purchase{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now().Add(-30 * 24 * time.Hour), // Compra há 30 dias
		},
		EbookID:   1,
		ClientID:  1,
		ExpiresAt: expiredTime,
		Ebook: models.Ebook{
			Title: "Test Ebook",
			Creator: models.Creator{
				Name:  "Test Creator",
				Email: "creator@test.com",
			},
		},
		Client: models.Client{
			Name:  "Test Client",
			Email: "client@test.com",
		},
	}

	// Criar request e response
	req := httptest.NewRequest("GET", "/purchase/download/1", nil)
	w := httptest.NewRecorder()

	// Chamar a função
	showExpiredDownloadPage(w, req, purchase)

	// Verificar se a resposta foi bem-sucedida
	assert.Equal(t, http.StatusOK, w.Code)

	// Verificar se o conteúdo contém informações sobre expiração
	body := w.Body.String()
	assert.Contains(t, body, "Download Expirado")
	assert.Contains(t, body, "Test Ebook")
	assert.Contains(t, body, "Test Creator")
	assert.Contains(t, body, "Expirado há")
}

func TestShowEbookFilesWithLimitExceeded(t *testing.T) {
	// Criar uma compra com limite atingido
	purchase := &models.Purchase{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
		},
		EbookID:       1,
		ClientID:      1,
		DownloadsUsed: 5,
		DownloadLimit: 5,
		Ebook: models.Ebook{
			Title: "Test Ebook",
			Creator: models.Creator{
				Name:  "Test Creator",
				Email: "creator@test.com",
			},
		},
		Client: models.Client{
			Name:  "Test Client",
			Email: "client@test.com",
		},
	}

	// Mock do repository
	mockRepo := &MockPurchaseRepository{}
	mockRepo.On("FindByID", uint(1)).Return(purchase, nil)

	// Verificar se a compra tem limite atingido
	assert.False(t, purchase.AvailableDownloads())
	assert.Equal(t, 5, purchase.DownloadsUsed)
	assert.Equal(t, 5, purchase.DownloadLimit)
}

func TestShowEbookFilesWithExpiredPurchase(t *testing.T) {
	// Criar uma compra expirada
	expiredTime := time.Now().Add(-24 * time.Hour) // Expirada há 1 dia
	purchase := &models.Purchase{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now().Add(-30 * 24 * time.Hour), // Compra há 30 dias
		},
		EbookID:   1,
		ClientID:  1,
		ExpiresAt: expiredTime,
		Ebook: models.Ebook{
			Title: "Test Ebook",
			Creator: models.Creator{
				Name:  "Test Creator",
				Email: "creator@test.com",
			},
		},
		Client: models.Client{
			Name:  "Test Client",
			Email: "client@test.com",
		},
	}

	// Verificar se a compra está expirada
	assert.True(t, purchase.IsExpired())
	assert.True(t, purchase.ExpiresAt.Before(time.Now()))
}

func TestShowEbookFilesWithValidPurchase(t *testing.T) {
	// Criar uma compra válida
	futureTime := time.Now().Add(30 * 24 * time.Hour) // Válida por mais 30 dias
	purchase := &models.Purchase{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
		},
		EbookID:       1,
		ClientID:      1,
		DownloadsUsed: 2,
		DownloadLimit: 5,
		ExpiresAt:     futureTime,
		Ebook: models.Ebook{
			Title: "Test Ebook",
			Creator: models.Creator{
				Name:  "Test Creator",
				Email: "creator@test.com",
			},
		},
		Client: models.Client{
			Name:  "Test Client",
			Email: "client@test.com",
		},
	}

	// Verificar se a compra é válida
	assert.True(t, purchase.AvailableDownloads())
	assert.False(t, purchase.IsExpired())
	assert.Equal(t, 2, purchase.DownloadsUsed)
	assert.Equal(t, 5, purchase.DownloadLimit)
	assert.True(t, purchase.ExpiresAt.After(time.Now()))
}
