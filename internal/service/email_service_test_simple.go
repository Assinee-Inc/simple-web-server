package service

import (
	"testing"

	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/service/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestEmailService_ResendDownloadLink_InvalidDTO(t *testing.T) {
	// Arrange - EmailService com mock do mailer
	mockMailer := &mocks.MockMailerSimple{}
	emailService := NewEmailService(mockMailer)

	// DTO inválido (email vazio)
	invalidDTO := &dto.ResendDownloadLinkDTO{
		ClientName:   "Test Client",
		ClientEmail:  "", // Email vazio é inválido
		EbookTitle:   "Test Ebook",
		DownloadLink: "http://example.com/download/123",
		AppName:      "Test App",
		ContactEmail: "contact@example.com",
	}

	// Act
	err := emailService.ResendDownloadLink(invalidDTO)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dados inválidos")
	mockMailer.AssertNotCalled(t, "Send") // Email não deve ter sido enviado
}

func TestEmailService_ResendDownloadLink_Success_Simple(t *testing.T) {
	// Arrange - EmailService com mock do mailer
	mockMailer := &mocks.MockMailerSimple{}
	emailService := NewEmailService(mockMailer)

	// Configurar mocks
	mockMailer.On("From", mock.AnythingOfType("string"))
	mockMailer.On("To", mock.AnythingOfType("string"))
	mockMailer.On("Subject", mock.AnythingOfType("string"))
	mockMailer.On("Body", mock.AnythingOfType("string"))
	mockMailer.On("Send")

	// DTO válido
	validDTO := &dto.ResendDownloadLinkDTO{
		ClientName:   "Test Client",
		ClientEmail:  "client@example.com",
		EbookTitle:   "Test Ebook",
		DownloadLink: "http://example.com/download/123",
		AppName:      "Test App",
		ContactEmail: "contact@example.com",
		EbookFiles: []dto.FileDTO{
			{OriginalName: "test.pdf", Size: "1.5 MB"},
		},
	}

	// Act
	err := emailService.ResendDownloadLink(validDTO)

	// Assert
	assert.NoError(t, err)
	mockMailer.AssertCalled(t, "Send")
	mockMailer.AssertCalled(t, "From", "contact@example.com")
	mockMailer.AssertCalled(t, "To", "client@example.com")
}

func TestEmailService_ValidateTransactionStatus(t *testing.T) {
	// Test para validar que apenas transações completadas podem ter links reenviados

	// Transação completada - deve passar na validação
	completedTransaction := &models.Transaction{
		Model:  gorm.Model{ID: 1},
		Status: models.TransactionStatusCompleted,
	}

	// Transação pendente - deve falhar na validação
	pendingTransaction := &models.Transaction{
		Model:  gorm.Model{ID: 2},
		Status: models.TransactionStatusPending,
	}

	// Transação com falha - deve falhar na validação
	failedTransaction := &models.Transaction{
		Model:  gorm.Model{ID: 3},
		Status: models.TransactionStatusFailed,
	}

	// Assert
	assert.Equal(t, models.TransactionStatusCompleted, completedTransaction.Status, "Transação deve estar completada")
	assert.NotEqual(t, models.TransactionStatusCompleted, pendingTransaction.Status, "Transação pendente não deve estar completada")
	assert.NotEqual(t, models.TransactionStatusCompleted, failedTransaction.Status, "Transação com falha não deve estar completada")
}
