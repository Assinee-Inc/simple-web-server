package service

import (
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
	data := map[string]interface{}{
		"Name":               name,
		"Title":              "Confirm your account!",
		"AppName":            config.AppConfig.AppName,
		"Contact":            config.AppConfig.MailFromAddress,
		"ConfirmAccountLink": "/account-confirmation?token=" + token + "&name=" + name + "&email=" + email,
	}
	s.prepareAndSendEmail(email, "Confirm your account", "account_confirmation", data)
}

func (s *EmailService) prepareAndSendEmail(to, subject, template string, data any) {
	s.mailer.From(config.AppConfig.AppName, config.AppConfig.MailFromAddress)
	s.mailer.To(to)
	s.mailer.Subject(subject)
	s.mailer.Body(mail.NewEmail(template, data))
	s.mailer.Send()
}

