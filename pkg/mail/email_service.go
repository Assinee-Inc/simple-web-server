package mail

import (
	"fmt"
	"log"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/models"
)

type EmailService struct {
	mailer Mailer
}

func NewEmailService(mailer Mailer) *EmailService {
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

	s.mailer.From(config.AppConfig.MailFromAddress)
	s.mailer.To(email)
	s.mailer.Subject("Recover your password!")
	s.mailer.Body(NewEmail("reset_password", data))
	s.mailer.Send()
}

func (s *EmailService) SendAccountConfirmation(name, email, token string) {
	data := map[string]interface{}{
		"Name":               name,
		"Title":              "Confirm your account!",
		"AppName":            config.AppConfig.AppName,
		"Contact":            config.AppConfig.MailFromAddress,
		"ConfirmAccountLink": "/account-confirmation?token=" + token + "&name=" + name + "&email=" + email,
	}

	s.mailer.From(config.AppConfig.MailFromAddress)
	s.mailer.To(email)
	s.mailer.Subject("Confirm your account")
	s.mailer.Body(NewEmail("account_confirmation", data))
	s.mailer.Send()
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

		data := map[string]interface{}{
			"Name":              purchase.Client.Name,
			"Title":             "Seu e-book chegou!",
			"AppName":           config.AppConfig.AppName,
			"Contact":           config.AppConfig.MailFromAddress,
			"EbookDownloadLink": fmt.Sprintf("%s:%s/purchase/download/%d", config.AppConfig.Host, config.AppConfig.Port, purchase.ID),
			"Ebook":             purchase.Ebook,
			"Files":             purchase.Ebook.Files,
			"FileCount":         len(purchase.Ebook.Files),
		}

		log.Printf("Configurando email para: %s", purchase.Client.Email)
		s.mailer.From(config.AppConfig.MailFromAddress)
		s.mailer.To(purchase.Client.Email)
		s.mailer.Subject("Seu e-book chegou!")
		s.mailer.Body(NewEmail("ebook_download", data))
		s.mailer.Send()
	}
}
