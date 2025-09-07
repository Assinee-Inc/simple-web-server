package service

import (
	"fmt"
	"log"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/internal/service/dto"
)

// ResendDownloadLinkService gerencia o reenvio de links de download
type ResendDownloadLinkService struct {
	transactionRepo repository.TransactionRepository
	purchaseRepo    *repository.PurchaseRepository
	emailService    IEmailService
}

// NewResendDownloadLinkService cria uma nova instância do serviço
func NewResendDownloadLinkService(
	transactionRepo repository.TransactionRepository,
	purchaseRepo *repository.PurchaseRepository,
	emailService IEmailService,
) *ResendDownloadLinkService {
	return &ResendDownloadLinkService{
		transactionRepo: transactionRepo,
		purchaseRepo:    purchaseRepo,
		emailService:    emailService,
	}
}

// ResendDownloadLinkByTransactionID reenvia o link de download com base no ID da transação
func (s *ResendDownloadLinkService) ResendDownloadLinkByTransactionID(transactionID uint) error {
	log.Printf("🔄 ResendDownloadLinkByTransactionID chamado para transactionID=%d", transactionID)

	// Buscar a transação pelo ID
	transaction, err := s.transactionRepo.FindByID(transactionID)
	if err != nil {
		log.Printf("❌ ERRO: Não foi possível encontrar transação ID=%d: %v", transactionID, err)
		return fmt.Errorf("transação não encontrada: %v", err)
	}

	// Verificar se a transação está completa
	if transaction.Status != models.TransactionStatusCompleted {
		log.Printf("❌ ERRO: Tentativa de reenvio para transação não completada. Status=%s", transaction.Status)
		return fmt.Errorf("não é possível reenviar link para transação com status: %s", transaction.Status)
	}

	// Buscar a purchase relacionada
	purchase, err := s.purchaseRepo.FindByID(transaction.PurchaseID)
	if err != nil {
		log.Printf("❌ ERRO: Não foi possível encontrar purchase ID=%d: %v", transaction.PurchaseID, err)
		return fmt.Errorf("compra não encontrada: %v", err)
	}

	// Verificar se o cliente foi carregado
	if purchase.Client.ID == 0 {
		log.Printf("❌ ERRO: Cliente não foi carregado! Purchase.ClientID=%d", purchase.ClientID)
		return fmt.Errorf("dados do cliente não encontrados")
	}

	// Verificar se o email está presente
	if purchase.Client.Email == "" {
		log.Printf("❌ ERRO: Email do cliente está vazio! ClientID=%d", purchase.ClientID)
		return fmt.Errorf("email do cliente não encontrado")
	}

	// Verificar se o ebook tem arquivos
	if len(purchase.Ebook.Files) == 0 {
		log.Printf("❌ ERRO: Ebook não possui arquivos! EbookID=%d", purchase.EbookID)
		return fmt.Errorf("ebook não possui arquivos para download")
	}

	// Converter arquivos para DTO
	filesDTOs := make([]dto.FileDTO, len(purchase.Ebook.Files))
	for i, file := range purchase.Ebook.Files {
		filesDTOs[i] = dto.FileDTO{
			OriginalName: file.OriginalName,
			Size:         file.GetFileSizeFormatted(),
		}
	}

	// Criar DTO para o EmailService
	downloadDTO := &dto.ResendDownloadLinkDTO{
		ClientName:   purchase.Client.Name,
		ClientEmail:  purchase.Client.Email,
		EbookTitle:   purchase.Ebook.Title,
		EbookFiles:   filesDTOs,
		DownloadLink: fmt.Sprintf("%s:%s/purchase/download/%d", config.AppConfig.Host, config.AppConfig.Port, purchase.ID),
		AppName:      config.AppConfig.AppName,
		ContactEmail: config.AppConfig.MailFromAddress,
	}

	// Enviar email através do EmailService
	err = s.emailService.ResendDownloadLink(downloadDTO)
	if err != nil {
		log.Printf("❌ ERRO: Falha ao enviar email: %v", err)
		return fmt.Errorf("falha ao enviar email: %v", err)
	}

	log.Printf("✅ Reenvio de link processado com sucesso para transactionID=%d", transactionID)
	return nil
}
