package service

import (
	"fmt"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/pkg/mail"
)

type IEmailService interface {
	SendPasswordResetEmail(name, email, resetLink string)
	SendAccountConfirmation(name, email, token string)
}

type EmailService struct {
	mailer mail.Mailer
}

func NewEmailService(mailer mail.Mailer) *EmailService {
	return &EmailService{mailer: mailer}
}

func (s *EmailService) SendPasswordResetEmail(name, email, resetLink string) {
	data := map[string]interface{}{
		"ResetLink": resetLink,
		"Name":      name,
		"Title":     "Recover your password!",
	}
	s.prepareAndSendEmail(email, "Recover your password!", "reset_password", data)
}

func (s *EmailService) SendAccountConfirmation(name, email, token string) {
	var baseURL string
	if config.AppConfig.IsProduction() {
		baseURL = config.AppConfig.Host
	} else {
		baseURL = fmt.Sprintf("%s:%s", config.AppConfig.Host, config.AppConfig.Port)
	}

	data := map[string]interface{}{
		"Name":               name,
		"Title":              "Confirme sua conta!",
		"appName":            config.AppConfig.AppName,
		"Contact":            config.AppConfig.MailFromAddress,
		"ConfirmAccountLink": fmt.Sprintf("%s/account-confirmation?token=%s", baseURL, token),
	}
	s.prepareAndSendEmail(email, "Confirme sua conta — "+config.AppConfig.AppName, "account_confirmation", data)
}

func (s *EmailService) prepareAndSendEmail(to, subject, template string, data any) {
	s.mailer.From(config.AppConfig.AppName, config.AppConfig.MailFromAddress)
	s.mailer.To(to)
	s.mailer.Subject(subject)
	s.mailer.Body(mail.NewEmail(template, data))
	s.mailer.Send()
}

