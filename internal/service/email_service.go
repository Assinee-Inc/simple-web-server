package service

import (
	"fmt"
	"log"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/service/dto"
	"github.com/anglesson/simple-web-server/pkg/mail"
)

type IEmailService interface {
	SendPasswordResetEmail(name, email, resetLink string)
	SendAccountConfirmation(name, email, token string)
	SendLinkToDownload(purchases []*models.Purchase)
	ResendDownloadLink(dto *dto.ResendDownloadLinkDTO) error
}

type EmailService struct {
	mailer mail.Mailer
}

func NewEmailService(mailer mail.Mailer) *EmailService {
	return &EmailService{
		mailer: mailer,
	}
}

func (s *EmailService) SendPasswordResetEmail(name, email string, resetLink string) {
	data := map[string]interface{}{
		"ResetLink": resetLink,
		"Name":      name,
		"Title":     "Recover your password!",
	}

	s.prepareAndSendEmail(
		email,
		"Recover your password!",
		"reset_password",
		data,
	)
}

func (s *EmailService) SendAccountConfirmation(name, email, token string) {
	data := map[string]interface{}{
		"Name":               name,
		"Title":              "Confirm your account!",
		"AppName":            config.AppConfig.AppName,
		"Contact":            config.AppConfig.MailFromAddress,
		"ConfirmAccountLink": "/account-confirmation?token=" + token + "&name=" + name + "&email=" + email,
	}

	s.prepareAndSendEmail(
		email,
		"Confirm your account",
		"account_confirmation",
		data,
	)
}

func (s *EmailService) SendLinkToDownload(purchases []*models.Purchase) {
	log.Printf("📧 SendLinkToDownload chamado com %d purchase(s)", len(purchases))

	for i, purchase := range purchases {
		log.Printf("📧 Processando purchase %d/%d", i+1, len(purchases))
		log.Printf("📧 Purchase ID=%d, ClientID=%d", purchase.ID, purchase.ClientID)
		log.Printf("📧 Client struct: %+v", purchase.Client)
		log.Printf("📧 Client ID=%d, Name='%s', Email='%s'",
			purchase.Client.ID, purchase.Client.Name, purchase.Client.Email)

		// Verificar se o cliente foi carregado
		if purchase.Client.ID == 0 {
			log.Printf("❌ ERRO: Cliente não foi carregado! Client.ID=0")
			continue
		}

		// Verificar se o email está vazio
		if purchase.Client.Email == "" {
			log.Printf("❌ ERRO: Email do cliente está vazio! ClientID=%d", purchase.ClientID)
			continue
		}

		downloadLink := s.buildDownloadURL(purchase.HashID)

		data := map[string]interface{}{
			"Name":              purchase.Client.Name,
			"Title":             "Seu e-book chegou!",
			"AppName":           config.AppConfig.AppName,
			"Contact":           config.AppConfig.MailFromAddress,
			"EbookDownloadLink": downloadLink,
			"Ebook":             purchase.Ebook,
			"Files":             purchase.Ebook.Files,
		}

		log.Printf("Configurando email para: %s", purchase.Client.Email)
		s.prepareAndSendEmail(
			purchase.Client.Email,
			"Seu e-book chegou!",
			"ebook_download",
			data,
		)
	}
}

func (s *EmailService) buildDownloadURL(hashID string) string {
	if config.AppConfig.IsProduction() {
		return fmt.Sprintf("%s/purchase/download/%s", config.AppConfig.Host, hashID)
	}
	return fmt.Sprintf("%s:%s/purchase/download/%s", config.AppConfig.Host, config.AppConfig.Port, hashID)
}

func (s *EmailService) ResendDownloadLink(downloadDTO *dto.ResendDownloadLinkDTO) error {
	log.Printf("📧 ResendDownloadLink chamado para cliente: %s", downloadDTO.ClientEmail)

	// Validar DTO
	if err := downloadDTO.Validate(); err != nil {
		log.Printf("❌ ERRO: Validação do DTO falhou: %v", err)
		return fmt.Errorf("dados inválidos: %v", err)
	}

	log.Printf("📧 Reenviando link de download para cliente: %s", downloadDTO.ClientEmail)

	// Preparar dados para o email
	data := map[string]interface{}{
		"Name":              downloadDTO.ClientName,
		"Title":             "Link de Download Reenviado - " + downloadDTO.EbookTitle,
		"AppName":           downloadDTO.AppName,
		"Contact":           downloadDTO.ContactEmail,
		"EbookDownloadLink": downloadDTO.DownloadLink,
		"Ebook":             map[string]interface{}{"Title": downloadDTO.EbookTitle},
		"Files":             downloadDTO.EbookFiles,
	}

	// Enviar email
	s.prepareAndSendEmail(
		downloadDTO.ClientEmail,
		"Link de Download Reenviado - "+downloadDTO.EbookTitle,
		"ebook_download",
		data,
	)

	log.Printf("✅ Link de download reenviado com sucesso para %s", downloadDTO.ClientEmail)
	return nil
}

func (s *EmailService) prepareAndSendEmail(to, subject, template string, data any) {
	s.mailer.From(config.AppConfig.AppName)
	s.mailer.To(to)
	s.mailer.Subject(subject)
	s.mailer.Body(mail.NewEmail(template, data))
	s.mailer.Send()
}
