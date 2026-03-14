package service_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestMain muda o diretório de trabalho para a raiz do projeto,
// necessário para que mail.NewEmail encontre os templates em web/mails/.
func TestMain(m *testing.M) {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..", "..")
	if err := os.Chdir(projectRoot); err != nil {
		panic("não foi possível navegar até a raiz do projeto: " + err.Error())
	}
	os.Exit(m.Run())
}

// TestSendAccountConfirmation_LinkIsAbsoluteURL garante que o e-mail de confirmação
// contém uma URL completa no botão, não um caminho relativo.
// REGRESSÃO: sem o host, o botão "Confirmar" chegava sem href funcional.
func TestSendAccountConfirmation_LinkIsAbsoluteURL(t *testing.T) {
	config.AppConfig.Host = "http://localhost"
	config.AppConfig.Port = "8080"
	config.AppConfig.AppName = "TestApp"
	config.AppConfig.MailFromAddress = "no-reply@testapp.com"
	config.AppConfig.AppMode = "development"

	mockMailer := new(mocks.MockMailerSimple)
	emailService := authsvc.NewEmailService(mockMailer)

	var capturedBody string
	mockMailer.On("From", mock.Anything).Return()
	mockMailer.On("To", "user@example.com").Return()
	mockMailer.On("Subject", mock.Anything).Return()
	mockMailer.On("Body", mock.MatchedBy(func(body string) bool {
		capturedBody = body
		return true
	})).Return()
	mockMailer.On("Send").Return()

	emailService.SendAccountConfirmation("Test User", "user@example.com", "abc123token")

	assert.NotEmpty(t, capturedBody, "o body do e-mail não deve estar vazio")

	assert.True(t,
		strings.Contains(capturedBody, "http://localhost:8080/account-confirmation?token=abc123token"),
		"REGRESSÃO: o link de confirmação deve ser uma URL completa com host e porta")

	assert.False(t,
		strings.Contains(capturedBody, `href="/account-confirmation`),
		"REGRESSÃO: link relativo detectado — o botão 'Confirmar' chegaria sem href funcional no e-mail")
}

// TestSendAccountConfirmation_LinkIsAbsoluteURL_Production garante que em produção
// o link usa apenas o host, sem porta.
func TestSendAccountConfirmation_LinkIsAbsoluteURL_Production(t *testing.T) {
	config.AppConfig.Host = "https://app.example.com"
	config.AppConfig.Port = "8080"
	config.AppConfig.AppName = "TestApp"
	config.AppConfig.MailFromAddress = "no-reply@testapp.com"
	config.AppConfig.AppMode = "production"

	mockMailer := new(mocks.MockMailerSimple)
	emailService := authsvc.NewEmailService(mockMailer)

	var capturedBody string
	mockMailer.On("From", mock.Anything).Return()
	mockMailer.On("To", "user@example.com").Return()
	mockMailer.On("Subject", mock.Anything).Return()
	mockMailer.On("Body", mock.MatchedBy(func(body string) bool {
		capturedBody = body
		return true
	})).Return()
	mockMailer.On("Send").Return()

	emailService.SendAccountConfirmation("Test User", "user@example.com", "prodtoken456")

	assert.True(t,
		strings.Contains(capturedBody, "https://app.example.com/account-confirmation?token=prodtoken456"),
		"em produção o link deve usar apenas o host sem porta")

	assert.False(t,
		strings.Contains(capturedBody, ":8080"),
		"em produção o link não deve conter a porta")
}
