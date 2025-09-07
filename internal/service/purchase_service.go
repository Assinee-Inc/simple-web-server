package service

import (
	"errors"
	"fmt"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
)

type PurchaseService struct {
	purchaseRepository *repository.PurchaseRepository
	mailService        IEmailService
}

func NewPurchaseService(purchaseRepository *repository.PurchaseRepository, mailService IEmailService) *PurchaseService {
	return &PurchaseService{
		purchaseRepository: purchaseRepository,
		mailService:        mailService,
	}
}

func (ps *PurchaseService) CreatePurchase(ebookId uint, clients []uint) error {
	var purchases []*models.Purchase

	for _, clientId := range clients {
		if clientId != 0 && ebookId != 0 {
			// Verificar se já existe uma purchase para este cliente/ebook
			existingPurchase, err := ps.purchaseRepository.FindExistingPurchase(ebookId, uint(clientId))
			if err == nil && existingPurchase != nil {
				// Purchase já existe, não criar duplicata
				continue
			}
			purchases = append(purchases, models.NewPurchase(ebookId, uint(clientId)))
		}
	}

	if len(purchases) == 0 {
		// Todas as purchases já existem ou não há purchases válidas para criar
		return nil
	}

	err := ps.purchaseRepository.CreateManyPurchases(purchases)
	if err != nil {
		return errors.New(err.Error())
	}

	go ps.mailService.SendLinkToDownload(purchases)
	return nil
}

// CreatePurchaseWithResult cria purchase e retorna a purchase criada ou existente
func (ps *PurchaseService) CreatePurchaseWithResult(ebookId uint, clientId uint) (*models.Purchase, error) {
	if clientId == 0 || ebookId == 0 {
		return nil, errors.New("clientId e ebookId devem ser válidos")
	}

	// Verificar se já existe uma purchase para este cliente/ebook
	existingPurchase, err := ps.purchaseRepository.FindExistingPurchase(ebookId, clientId)
	if err == nil && existingPurchase != nil {
		// Purchase já existe, retornar a existente
		return existingPurchase, nil
	}

	// Criar nova purchase
	purchase := models.NewPurchase(ebookId, clientId)
	purchases := []*models.Purchase{purchase}

	err = ps.purchaseRepository.CreateManyPurchases(purchases)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	go ps.mailService.SendLinkToDownload(purchases)
	return purchase, nil
}

func (ps *PurchaseService) GetPurchaseByID(id uint) (*models.Purchase, error) {
	return ps.purchaseRepository.FindByID(id)
}

func (ps *PurchaseService) GetEbookFile(purchaseID int, fileID uint) (string, error) {
	purchase, err := ps.purchaseRepository.FindByID(uint(purchaseID))
	if err != nil {
		return "", errors.New(err.Error())
	}

	if !purchase.AvailableDownloads() {
		return "", errors.New("não é possível realizar o download, limite de downloads atingido")
	}

	if purchase.IsExpired() {
		return "", errors.New("não é possível realizar o download, o pedido está expirado")
	}

	// Buscar o arquivo específico do ebook
	var targetFile *models.File
	for _, file := range purchase.Ebook.Files {
		if file.ID == fileID {
			targetFile = file
			break
		}
	}

	if targetFile == nil {
		return "", errors.New("arquivo não encontrado neste ebook")
	}

	// Aplicar marca d'água no arquivo
	watermarkText := fmt.Sprintf("%s - %s - %s", purchase.Client.Name, purchase.Client.CPF, purchase.Client.Email)
	outputFilePath, err := ApplyWatermark(targetFile.S3Key, watermarkText)
	if err != nil {
		return "", err
	}

	purchase.UseDownload()
	ps.purchaseRepository.Update(purchase)

	return outputFilePath, nil
}

// GetEbookFiles retorna todos os arquivos do ebook para um cliente
func (ps *PurchaseService) GetEbookFiles(purchaseID int) ([]*models.File, error) {
	purchase, err := ps.purchaseRepository.FindByID(uint(purchaseID))
	if err != nil {
		return nil, errors.New(err.Error())
	}

	if !purchase.AvailableDownloads() {
		return nil, errors.New("não é possível realizar o download, limite de downloads atingido")
	}

	if purchase.IsExpired() {
		return nil, errors.New("não é possível realizar o download, o pedido está expirado")
	}

	if len(purchase.Ebook.Files) == 0 {
		return nil, errors.New("nenhum arquivo encontrado neste ebook")
	}

	return purchase.Ebook.Files, nil
}
