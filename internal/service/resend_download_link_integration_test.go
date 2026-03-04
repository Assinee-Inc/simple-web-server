package service

import (
	"testing"

	"github.com/anglesson/simple-web-server/internal/service/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMailerForIntegration é um mock que captura dados do email sem processar templates
type MockMailerForIntegration struct {
	mock.Mock
	SentEmails []SentEmailData
}

type SentEmailData struct {
	From    string
	To      string
	Subject string
	Body    string
}

func (m *MockMailerForIntegration) From(email string) {
	m.Called(email)
	if len(m.SentEmails) == 0 {
		m.SentEmails = append(m.SentEmails, SentEmailData{})
	}
	m.SentEmails[len(m.SentEmails)-1].From = email
}

func (m *MockMailerForIntegration) To(email string) {
	m.Called(email)
	if len(m.SentEmails) == 0 {
		m.SentEmails = append(m.SentEmails, SentEmailData{})
	}
	m.SentEmails[len(m.SentEmails)-1].To = email
}

func (m *MockMailerForIntegration) Subject(subject string) {
	m.Called(subject)
	if len(m.SentEmails) == 0 {
		m.SentEmails = append(m.SentEmails, SentEmailData{})
	}
	m.SentEmails[len(m.SentEmails)-1].Subject = subject
}

func (m *MockMailerForIntegration) Body(body string) {
	m.Called(body)
	if len(m.SentEmails) == 0 {
		m.SentEmails = append(m.SentEmails, SentEmailData{})
	}
	m.SentEmails[len(m.SentEmails)-1].Body = body
}

func (m *MockMailerForIntegration) Send() {
	m.Called()
}

// TestEmailService_ResendDownloadLink_FunctionalTest testa a lógica principal com templates
func TestEmailService_ResendDownloadLink_FunctionalTest(t *testing.T) {
	// Usar helper para acessar templates na raiz do projeto
	cleanup := changeToProjectRoot(t)
	defer cleanup()

	t.Run("Deve processar DTO válido e chamar mailer corretamente", func(t *testing.T) {
		// Arrange
		mockMailer := &MockMailerForIntegration{}
		emailService := NewEmailService(mockMailer)

		validDTO := &dto.ResendDownloadLinkDTO{
			ClientName:  "João Silva",
			ClientEmail: "joao@teste.com",
			EbookTitle:  "Ebook de Teste",
			EbookFiles: []dto.FileDTO{
				{OriginalName: "arquivo1.pdf", Size: "2.5 MB"},
				{OriginalName: "arquivo2.pdf", Size: "1.8 MB"},
			},
			DownloadLink: "https://exemplo.com/download/123",
			AppName:      "MeuApp",
			ContactEmail: "contato@exemplo.com",
		}

		// Setup mocks - intercepta antes do processamento de template
		mockMailer.On("From", mock.AnythingOfType("string"))
		mockMailer.On("To", "joao@teste.com")
		mockMailer.On("Subject", "Link de Download Reenviado - Ebook de Teste")
		mockMailer.On("Body", mock.AnythingOfType("string")) // Aceita qualquer string (incluindo template)
		mockMailer.On("Send")

		// Act - Agora deve funcionar com templates
		err := emailService.ResendDownloadLink(validDTO)

		// Assert - Deve ter sucesso
		assert.NoError(t, err)
		mockMailer.AssertExpectations(t)

		t.Logf("✅ SUCESSO: Lógica principal funcionando!")
		t.Logf("📧 Email configurado para: %s", validDTO.ClientEmail)
		t.Logf("📚 Ebook: %s", validDTO.EbookTitle)
		t.Logf("📄 Arquivos: %d", len(validDTO.EbookFiles))
		t.Logf("🔗 Link: %s", validDTO.DownloadLink)
	})

	t.Run("Deve falhar para DTO inválido (antes do template)", func(t *testing.T) {
		// Arrange
		mockMailer := &MockMailerForIntegration{}
		emailService := NewEmailService(mockMailer)

		invalidDTO := &dto.ResendDownloadLinkDTO{
			ClientName:   "João Silva",
			ClientEmail:  "", // Email vazio - deve falhar na validação
			EbookTitle:   "Ebook de Teste",
			DownloadLink: "https://exemplo.com/download/123",
			AppName:      "MeuApp",
			ContactEmail: "contato@exemplo.com",
		}

		// Act
		err := emailService.ResendDownloadLink(invalidDTO)

		// Assert - Deve falhar na validação (antes do template)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dados inválidos")
		assert.Contains(t, err.Error(), "email do cliente é obrigatório")

		// Verificar que o mailer NÃO foi chamado
		mockMailer.AssertNotCalled(t, "Send")

		t.Logf("✅ SUCESSO: Validação funcionando corretamente!")
		t.Logf("❌ Erro esperado: %v", err)
	})
}
