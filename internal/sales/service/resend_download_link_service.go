package service

import (
	"fmt"
	"log"

	"github.com/anglesson/simple-web-server/internal/config"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	salesrepo "github.com/anglesson/simple-web-server/internal/sales/repository"
	salesdto "github.com/anglesson/simple-web-server/internal/sales/service/dto"
)

// ResendDownloadLinkServiceInterface define a interface para o serviço
type ResendDownloadLinkServiceInterface interface {
	ResendDownloadLinkByTransactionID(transactionID uint) error
	ResendDownloadLinkByPurchaseID(purchaseID uint, newEmail string) error
}

// ResendDownloadLinkService gerencia o reenvio de links de download
type ResendDownloadLinkService struct {
	transactionRepo salesrepo.TransactionRepository
	purchaseRepo    *salesrepo.PurchaseRepository
	emailService    IEmailService
}

// NewResendDownloadLinkService cria uma nova instância do serviço
func NewResendDownloadLinkService(
	transactionRepo salesrepo.TransactionRepository,
	purchaseRepo *salesrepo.PurchaseRepository,
	emailService IEmailService,
) ResendDownloadLinkServiceInterface {
	return &ResendDownloadLinkService{
		transactionRepo: transactionRepo,
		purchaseRepo:    purchaseRepo,
		emailService:    emailService,
	}
}

// ResendDownloadLinkByTransactionID reenvia o link de download com base no ID da transação
func (s *ResendDownloadLinkService) ResendDownloadLinkByTransactionID(transactionID uint) error {
	log.Printf("🔄 ResendDownloadLinkByTransactionID chamado para transactionID=%d", transactionID)

	transaction, err := s.transactionRepo.FindByID(transactionID)
	if err != nil {
		log.Printf("❌ ERRO: Não foi possível encontrar transação ID=%d: %v", transactionID, err)
		return fmt.Errorf("transação não encontrada: %v", err)
	}

	if transaction.Status != salesmodel.TransactionStatusCompleted {
		log.Printf("❌ ERRO: Tentativa de reenvio para transação não completada. Status=%s", transaction.Status)
		return fmt.Errorf("não é possível reenviar link para transação com status: %s", transaction.Status)
	}

	purchase, err := s.purchaseRepo.FindByID(transaction.PurchaseID)
	if err != nil {
		log.Printf("❌ ERRO: Não foi possível encontrar purchase ID=%d: %v", transaction.PurchaseID, err)
		return fmt.Errorf("compra não encontrada: %v", err)
	}

	if purchase.Client.ID == 0 {
		log.Printf("❌ ERRO: Cliente não foi carregado! Purchase.ClientID=%d", purchase.ClientID)
		return fmt.Errorf("dados do cliente não encontrados")
	}

	if purchase.Client.Email == "" {
		log.Printf("❌ ERRO: Email do cliente está vazio! ClientID=%d", purchase.ClientID)
		return fmt.Errorf("email do cliente não encontrado")
	}

	if len(purchase.Ebook.Files) == 0 {
		log.Printf("❌ ERRO: Ebook não possui arquivos! EbookID=%d", purchase.EbookID)
		return fmt.Errorf("ebook não possui arquivos para download")
	}

	filesDTOs := make([]salesdto.FileDTO, len(purchase.Ebook.Files))
	for i, file := range purchase.Ebook.Files {
		filesDTOs[i] = salesdto.FileDTO{
			OriginalName: file.OriginalName,
			Size:         file.GetFileSizeFormatted(),
		}
	}

	downloadDTO := &salesdto.ResendDownloadLinkDTO{
		ClientName:   purchase.Client.Name,
		ClientEmail:  purchase.Client.Email,
		EbookTitle:   purchase.Ebook.Title,
		EbookFiles:   filesDTOs,
		DownloadLink: fmt.Sprintf("%s:%s/purchase/download/%d", config.AppConfig.Host, config.AppConfig.Port, purchase.ID),
		AppName:      config.AppConfig.AppName,
		ContactEmail: config.AppConfig.MailFromAddress,
	}

	err = s.emailService.ResendDownloadLink(downloadDTO)
	if err != nil {
		log.Printf("❌ ERRO: Falha ao enviar email: %v", err)
		return fmt.Errorf("falha ao enviar email: %v", err)
	}

	log.Printf("✅ Reenvio de link processado com sucesso para transactionID=%d", transactionID)
	return nil
}

// ResendDownloadLinkByPurchaseID reenvia o link de download com base no ID da purchase
func (s *ResendDownloadLinkService) ResendDownloadLinkByPurchaseID(purchaseID uint, newEmail string) error {
	log.Printf("🔄 ResendDownloadLinkByPurchaseID chamado para purchaseID=%d, newEmail=%s", purchaseID, newEmail)

	purchase, err := s.purchaseRepo.FindByID(purchaseID)
	if err != nil {
		log.Printf("❌ ERRO: Não foi possível encontrar purchase ID=%d: %v", purchaseID, err)
		return fmt.Errorf("compra não encontrada: %v", err)
	}

	if purchase.Client.ID == 0 {
		log.Printf("❌ ERRO: Cliente não foi carregado! Purchase.ClientID=%d", purchase.ClientID)
		return fmt.Errorf("dados do cliente não encontrados")
	}

	emailToUse := purchase.Client.Email
	if newEmail != "" {
		emailToUse = newEmail
	}

	if emailToUse == "" {
		log.Printf("❌ ERRO: Email não fornecido! ClientID=%d", purchase.ClientID)
		return fmt.Errorf("email não encontrado")
	}

	if len(purchase.Ebook.Files) == 0 {
		log.Printf("❌ ERRO: Ebook não possui arquivos! EbookID=%d", purchase.EbookID)
		return fmt.Errorf("ebook não possui arquivos para download")
	}

	filesDTOs := make([]salesdto.FileDTO, len(purchase.Ebook.Files))
	for i, file := range purchase.Ebook.Files {
		filesDTOs[i] = salesdto.FileDTO{
			OriginalName: file.OriginalName,
			Size:         file.GetFileSizeFormatted(),
		}
	}

	downloadDTO := &salesdto.ResendDownloadLinkDTO{
		ClientName:   purchase.Client.Name,
		ClientEmail:  emailToUse,
		EbookTitle:   purchase.Ebook.Title,
		EbookFiles:   filesDTOs,
		DownloadLink: fmt.Sprintf("%s:%s/purchase/download/%d", config.AppConfig.Host, config.AppConfig.Port, purchase.ID),
		AppName:      config.AppConfig.AppName,
		ContactEmail: config.AppConfig.MailFromAddress,
	}

	err = s.emailService.ResendDownloadLink(downloadDTO)
	if err != nil {
		log.Printf("❌ ERRO: Falha ao enviar email: %v", err)
		return fmt.Errorf("falha ao enviar email: %v", err)
	}

	log.Printf("✅ Reenvio de link processado com sucesso para purchaseID=%d", purchaseID)
	return nil
}
