package service

import (
	"testing"
	"time"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestPurchaseModelMethods(t *testing.T) {
	// Teste para AvailableDownloads com limite ilimitado
	purchaseUnlimited := &models.Purchase{
		DownloadLimit: -1,
		DownloadsUsed: 10,
	}
	assert.True(t, purchaseUnlimited.AvailableDownloads())

	// Teste para AvailableDownloads com limite atingido
	purchaseLimited := &models.Purchase{
		DownloadLimit: 5,
		DownloadsUsed: 5,
	}
	assert.False(t, purchaseLimited.AvailableDownloads())

	// Teste para AvailableDownloads com downloads disponíveis
	purchaseAvailable := &models.Purchase{
		DownloadLimit: 5,
		DownloadsUsed: 2,
	}
	assert.True(t, purchaseAvailable.AvailableDownloads())

	// Teste para IsExpired com data de expiração no passado
	expiredTime := time.Now().Add(-24 * time.Hour)
	purchaseExpired := &models.Purchase{
		ExpiresAt: expiredTime,
	}
	assert.True(t, purchaseExpired.IsExpired())

	// Teste para IsExpired com data de expiração no futuro
	futureTime := time.Now().Add(24 * time.Hour)
	purchaseValid := &models.Purchase{
		ExpiresAt: futureTime,
	}
	assert.False(t, purchaseValid.IsExpired())

	// Teste para IsExpired sem data de expiração
	purchaseNoExpiry := &models.Purchase{
		ExpiresAt: time.Time{},
	}
	assert.False(t, purchaseNoExpiry.IsExpired())

	// Teste para UseDownload
	purchase := &models.Purchase{
		DownloadsUsed: 0,
	}
	purchase.UseDownload()
	assert.Equal(t, 1, purchase.DownloadsUsed)
	assert.Len(t, purchase.Downloads, 1)
}

func TestPurchaseValidationLogic(t *testing.T) {
	// Teste para compra com limite atingido
	purchaseLimitExceeded := &models.Purchase{
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
		},
		Client: models.Client{
			Name:  "Test Client",
			Email: "client@test.com",
		},
	}

	// Verificar se a compra tem limite atingido
	assert.False(t, purchaseLimitExceeded.AvailableDownloads())
	assert.Equal(t, 5, purchaseLimitExceeded.DownloadsUsed)
	assert.Equal(t, 5, purchaseLimitExceeded.DownloadLimit)

	// Teste para compra expirada
	expiredTime := time.Now().Add(-24 * time.Hour) // Expirada há 1 dia
	purchaseExpired := &models.Purchase{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now().Add(-30 * 24 * time.Hour), // Compra há 30 dias
		},
		EbookID:   1,
		ClientID:  1,
		ExpiresAt: expiredTime,
		Ebook: models.Ebook{
			Title: "Test Ebook",
		},
		Client: models.Client{
			Name:  "Test Client",
			Email: "client@test.com",
		},
	}

	// Verificar se a compra está expirada
	assert.True(t, purchaseExpired.IsExpired())
	assert.True(t, purchaseExpired.ExpiresAt.Before(time.Now()))

	// Teste para compra válida
	futureTime := time.Now().Add(30 * 24 * time.Hour) // Válida por mais 30 dias
	purchaseValid := &models.Purchase{
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
		},
		Client: models.Client{
			Name:  "Test Client",
			Email: "client@test.com",
		},
	}

	// Verificar se a compra é válida
	assert.True(t, purchaseValid.AvailableDownloads())
	assert.False(t, purchaseValid.IsExpired())
	assert.Equal(t, 2, purchaseValid.DownloadsUsed)
	assert.Equal(t, 5, purchaseValid.DownloadLimit)
	assert.True(t, purchaseValid.ExpiresAt.After(time.Now()))
}
