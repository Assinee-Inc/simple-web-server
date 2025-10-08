package service

import (
	"testing"

	"github.com/anglesson/simple-web-server/internal/service/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMailer é um mock do Mailer
type MockMailer struct {
	mock.Mock
}

func (m *MockMailer) From(email string) {
	m.Called(email)
}

func (m *MockMailer) To(email string) {
	m.Called(email)
}

func (m *MockMailer) Subject(subject string) {
	m.Called(subject)
}

func (m *MockMailer) Body(body string) {
	m.Called(body)
}

func (m *MockMailer) Send() {
	m.Called()
}

func TestEmailService_ResendDownloadLink_Success(t *testing.T) {
	// Usar helper para acessar templates na raiz do projeto
	cleanup := changeToProjectRoot(t)
	defer cleanup()

	// Arrange
	mockMailer := &MockMailer{}
	emailService := NewEmailService(mockMailer)

	downloadDTO := &dto.ResendDownloadLinkDTO{
		ClientName:  "João Silva",
		ClientEmail: "joao@teste.com",
		EbookTitle:  "Ebook de Teste",
		EbookFiles: []dto.FileDTO{
			{OriginalName: "arquivo1.pdf", Size: "2.5 MB"},
		},
		DownloadLink: "https://example.com/download/123",
		AppName:      "MeuApp",
		ContactEmail: "contato@exemplo.com",
	}

	// Setup mocks
	mockMailer.On("From", "contato@exemplo.com")
	mockMailer.On("To", "joao@teste.com")
	mockMailer.On("Subject", "Link de Download Reenviado - Ebook de Teste")
	mockMailer.On("Body", mock.AnythingOfType("string"))
	mockMailer.On("Send")

	// Act
	err := emailService.ResendDownloadLink(downloadDTO)

	// Assert
	assert.NoError(t, err)
	mockMailer.AssertExpectations(t)
}

func TestEmailService_ResendDownloadLink_InvalidData(t *testing.T) {
	// Arrange
	mockMailer := &MockMailer{}
	emailService := NewEmailService(mockMailer)

	// DTO sem email do cliente
	downloadDTO := &dto.ResendDownloadLinkDTO{
		ClientName:  "João Silva",
		ClientEmail: "", // Email vazio
		EbookTitle:  "Ebook de Teste",
		EbookFiles: []dto.FileDTO{
			{OriginalName: "arquivo1.pdf", Size: "2.5 MB"},
		},
		DownloadLink: "https://example.com/download/123",
		AppName:      "MeuApp",
		ContactEmail: "contato@exemplo.com",
	}

	// Act
	err := emailService.ResendDownloadLink(downloadDTO)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dados inválidos")
	mockMailer.AssertNotCalled(t, "Send")
}

func TestEmailService_ResendDownloadLink_NoFiles(t *testing.T) {
	// Arrange
	mockMailer := &MockMailer{}
	emailService := NewEmailService(mockMailer)

	// DTO sem arquivos
	downloadDTO := &dto.ResendDownloadLinkDTO{
		ClientName:   "João Silva",
		ClientEmail:  "joao@teste.com",
		EbookTitle:   "Ebook de Teste",
		EbookFiles:   []dto.FileDTO{}, // Sem arquivos
		DownloadLink: "https://example.com/download/123",
		AppName:      "MeuApp",
		ContactEmail: "contato@exemplo.com",
	}

	// Act
	err := emailService.ResendDownloadLink(downloadDTO)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ebook deve ter pelo menos um arquivo")
	mockMailer.AssertNotCalled(t, "Send")
}
