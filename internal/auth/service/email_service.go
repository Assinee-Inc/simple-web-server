package service

import (
	"github.com/Assinee-Inc/docffy/pkg/mail"
)

// IEmailService define o contrato de envio de e-mails do módulo de autenticação.
// Fluxos legados de recuperação de senha e confirmação manual de e-mail foram removidos (FR-002 / FR-003).
type IEmailService interface {
}

type EmailService struct {
	mailer mail.Mailer
}

func NewEmailService(mailer mail.Mailer) *EmailService {
	return &EmailService{mailer: mailer}
}
